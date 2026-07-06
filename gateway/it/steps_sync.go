/*
 * Copyright (c) 2025, WSO2 LLC. (https://www.wso2.com).
 *
 * WSO2 LLC. licenses this file to you under the Apache License,
 * Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package it

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	adminapi "github.com/wso2/api-platform/gateway/gateway-controller/pkg/api/admin"
	"github.com/wso2/api-platform/gateway/it/steps"
)

const (
	// policyPropagationPollInterval is how often awaitPolicyPropagation probes
	// the xds_sync_status endpoints.
	policyPropagationPollInterval = 100 * time.Millisecond

	// policyGracePeriod bounds how long awaitPolicyPropagation holds out for
	// the ideal signal (version bump plus policy-engine sync) before accepting
	// a weaker one: mutations that never touch the policy snapshot show no
	// bump, and snapshot updates that only delete resources are never pushed
	// to the policy engine (LinearCache does not notify state-of-the-world
	// watches for pure deletions), so equality may never be reached.
	policyGracePeriod = 5 * time.Second

	// policyPropagationTimeout bounds the total wait for a mutation to reach
	// the policy engine.
	policyPropagationTimeout = 15 * time.Second
)

// policySnapshotAdminBase returns the admin base URL of the controller that
// feeds xDS to gateway-runtime. In the two-controller Postgres topology this
// is gateway-controller-xds; otherwise the management controller.
func policySnapshotAdminBase(state *TestState) string {
	if state.Config.PolicySnapshotControllerAdminURL != "" {
		return state.Config.PolicySnapshotControllerAdminURL
	}
	return state.Config.GatewayControllerAdminURL
}

// fetchControllerPolicyVersion reads the controller's latest published
// policy-chain version from its admin xds_sync_status endpoint.
func fetchControllerPolicyVersion(state *TestState) (string, error) {
	url := fmt.Sprintf("%s/xds_sync_status", policySnapshotAdminBase(state))
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}

	if admin, ok := state.Config.Users["admin"]; ok {
		req.SetBasicAuth(admin.Username, admin.Password)
	}

	resp, err := state.HTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("controller xds sync endpoint returned status %d", resp.StatusCode)
	}

	return decodePolicyChainVersion(resp)
}

// fetchPolicyEnginePolicyVersion reads the policy-chain version the policy
// engine last received via xDS.
func fetchPolicyEnginePolicyVersion(state *TestState) (string, error) {
	url := fmt.Sprintf("%s/xds_sync_status", state.Config.PolicyEngineURL)
	resp, err := state.HTTPClient.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("policy engine xds sync endpoint returned status %d", resp.StatusCode)
	}

	return decodePolicyChainVersion(resp)
}

func decodePolicyChainVersion(resp *http.Response) (string, error) {
	var payload adminapi.XDSSyncStatusResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", err
	}
	if payload.PolicyChainVersion == nil {
		return "", fmt.Errorf("policy chain version is nil in response")
	}
	return *payload.PolicyChainVersion, nil
}

// capturePolicyVersion snapshots the controller's current policy-chain version
// so a following awaitPolicyPropagation call can detect the mutation's version
// bump. Returns "" when the version cannot be read; awaitPolicyPropagation
// then degrades to the equality-only check.
func capturePolicyVersion(state *TestState) string {
	version, err := fetchControllerPolicyVersion(state)
	if err != nil {
		return ""
	}
	return version
}

// awaitPolicyPropagation blocks until a mutating controller call (create or
// update) has propagated to the policy engine. Controller mutations publish an
// event and update the policy xDS snapshot asynchronously, so this polls until
// the controller's policy-chain version has moved past preVersion (captured
// via capturePolicyVersion before the mutation) and the policy engine reports
// the same version. When the last response is not 2xx the mutation was
// rejected and there is nothing to propagate, so it returns immediately.
func awaitPolicyPropagation(state *TestState, httpSteps *steps.HTTPSteps, preVersion string) error {
	return awaitPolicyChange(state, httpSteps, preVersion, true)
}

// awaitPolicyDeletion blocks until a deleting controller call has been applied
// to the controller's policy xDS snapshot. Unlike creates and updates, the
// resulting snapshot update removes resources without changing any, and
// LinearCache does not notify state-of-the-world watches for pure deletions —
// the policy engine never sees the new version, so only the controller-side
// version bump is waited for. Route removal reaches Envoy through its own xDS
// snapshot in the same update.
func awaitPolicyDeletion(state *TestState, httpSteps *steps.HTTPSteps, preVersion string) error {
	return awaitPolicyChange(state, httpSteps, preVersion, false)
}

func awaitPolicyChange(state *TestState, httpSteps *steps.HTTPSteps, preVersion string, requireEngineSync bool) error {
	resp := httpSteps.LastResponse()
	if resp == nil || resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil
	}

	start := time.Now()
	var controllerVersion, engineVersion string
	var controllerErr, engineErr error
	for {
		controllerVersion, controllerErr = fetchControllerPolicyVersion(state)
		bumped := controllerErr == nil && controllerVersion != "" &&
			(preVersion == "" || controllerVersion != preVersion)

		synced := true
		if requireEngineSync {
			engineVersion, engineErr = fetchPolicyEnginePolicyVersion(state)
			synced = engineErr == nil && engineVersion == controllerVersion
		}

		if bumped && synced {
			return nil
		}

		elapsed := time.Since(start)
		if elapsed >= policyGracePeriod && controllerErr == nil {
			if bumped {
				// The controller applied the change but the policy engine
				// never echoed the new version — a deletion-only snapshot
				// update, which LinearCache does not push to SotW watches.
				log.Printf("Policy version %s not echoed by policy engine %.1fs after mutation (deletion-only update); proceeding", controllerVersion, elapsed.Seconds())
				return nil
			}
			if !requireEngineSync || (engineErr == nil && engineVersion == controllerVersion) {
				// Controller (and policy engine, when required) agree but the
				// version never moved past preVersion: the mutation did not
				// change the policy snapshot.
				log.Printf("Policy version %s unchanged %.1fs after mutation; treating as no-op", controllerVersion, elapsed.Seconds())
				return nil
			}
		}
		if elapsed >= policyPropagationTimeout {
			return fmt.Errorf("policy change did not propagate within %s: pre_version=%q controller_version=%q policy_engine_version=%q controller_err=%v policy_engine_err=%v",
				policyPropagationTimeout, preVersion, controllerVersion, engineVersion, controllerErr, engineErr)
		}

		time.Sleep(policyPropagationPollInterval)
	}
}

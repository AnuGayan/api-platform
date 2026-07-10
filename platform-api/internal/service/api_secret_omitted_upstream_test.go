/*
 *  Copyright (c) 2026, WSO2 LLC. (http://www.wso2.org) All Rights Reserved.
 *
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *  http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 *
 */

// Tests covering credential preservation on REST API updates that omit the
// upstream block (or the sandbox endpoint). Auth values are write-only —
// redacted on read — so an omitted upstream/endpoint must be treated as "keep
// the stored credential", never as "delete it". Without preservation the
// stored credential is persisted empty and its secret is deprecated by the
// rotation-cleanup path.

package service

import (
	"errors"
	"testing"

	"github.com/wso2/api-platform/platform-api/api"
	"github.com/wso2/api-platform/platform-api/internal/model"
	"github.com/wso2/api-platform/platform-api/internal/utils"
)

// TestApplyAPIUpdates_OmittedUpstream_PreservesStoredCredential proves an
// update that omits the upstream block (e.g. renaming the API) does not wipe
// the stored upstream credential.
func TestApplyAPIUpdates_OmittedUpstream_PreservesStoredCredential(t *testing.T) {
	service := &APIService{
		apiRepo:     &mockAPIRepository{},
		projectRepo: &mockProjectRepository{projectByUUID: &model.Project{ID: "11111111-1111-1111-1111-111111111111", Handle: "test-project"}},
		apiUtil:     &utils.APIUtil{},
		identity:    newTestIdentityService(),
	}

	existing := &model.API{
		Handle:    "pets-api",
		ProjectID: "11111111-1111-1111-1111-111111111111",
		Version:   "v1",
		Configuration: model.RestAPIConfig{
			Upstream: model.UpstreamConfig{
				Main: &model.UpstreamEndpoint{
					URL:  "https://backend.internal/api",
					Auth: &model.UpstreamAuth{Type: "bearer", Header: "Authorization", Value: `{{ secret "handle-a" }}`},
				},
			},
		},
	}

	req := &api.RESTAPI{DisplayName: "Renamed API"} // upstream omitted entirely

	updated, err := service.applyAPIUpdates(existing, req, "org-1")
	if err != nil {
		t.Fatalf("applyAPIUpdates() error = %v", err)
	}

	if updated.Upstream.Main.Auth == nil || updated.Upstream.Main.Auth.Value == nil {
		t.Fatalf("stored upstream credential was wiped on an update that omitted upstream: auth=%+v", updated.Upstream.Main.Auth)
	}
	if *updated.Upstream.Main.Auth.Value != `{{ secret "handle-a" }}` {
		t.Errorf("expected stored credential preserved, got %q", *updated.Upstream.Main.Auth.Value)
	}
	if updated.Upstream.Main.Url == nil || *updated.Upstream.Main.Url != "https://backend.internal/api" {
		t.Errorf("expected stored upstream URL preserved, got %v", updated.Upstream.Main.Url)
	}
}

// TestApplyAPIUpdates_SandboxOmittedWhileMainPresent_PreservesSandboxCredential
// proves an update that sends Main but omits Sandbox keeps the stored sandbox
// endpoint and its credential.
func TestApplyAPIUpdates_SandboxOmittedWhileMainPresent_PreservesSandboxCredential(t *testing.T) {
	service := &APIService{
		apiRepo:     &mockAPIRepository{},
		projectRepo: &mockProjectRepository{projectByUUID: &model.Project{ID: "11111111-1111-1111-1111-111111111111", Handle: "test-project"}},
		apiUtil:     &utils.APIUtil{},
		identity:    newTestIdentityService(),
	}

	existing := &model.API{
		Handle:    "pets-api",
		ProjectID: "11111111-1111-1111-1111-111111111111",
		Version:   "v1",
		Configuration: model.RestAPIConfig{
			Upstream: model.UpstreamConfig{
				Main: &model.UpstreamEndpoint{
					URL:  "https://backend.internal/api",
					Auth: &model.UpstreamAuth{Type: "bearer", Header: "Authorization", Value: `{{ secret "main-handle" }}`},
				},
				Sandbox: &model.UpstreamEndpoint{
					URL:  "https://sandbox.internal/api",
					Auth: &model.UpstreamAuth{Type: "bearer", Header: "Authorization", Value: `{{ secret "sandbox-handle" }}`},
				},
			},
		},
	}

	req := &api.RESTAPI{
		Upstream: api.Upstream{
			Main: api.UpstreamDefinition{Url: utils.StringPtrIfNotEmpty("https://backend-v2.internal/api")},
		},
	}

	updated, err := service.applyAPIUpdates(existing, req, "org-1")
	if err != nil {
		t.Fatalf("applyAPIUpdates() error = %v", err)
	}

	if updated.Upstream.Sandbox == nil || updated.Upstream.Sandbox.Auth == nil || updated.Upstream.Sandbox.Auth.Value == nil {
		t.Fatalf("sandbox credential was wiped when Main was sent but Sandbox omitted: sandbox=%+v", updated.Upstream.Sandbox)
	}
	if *updated.Upstream.Sandbox.Auth.Value != `{{ secret "sandbox-handle" }}` {
		t.Errorf("expected sandbox credential preserved, got %q", *updated.Upstream.Sandbox.Auth.Value)
	}
	if updated.Upstream.Main.Url == nil || *updated.Upstream.Main.Url != "https://backend-v2.internal/api" {
		t.Errorf("expected Main URL updated, got %v", updated.Upstream.Main.Url)
	}
}

// TestAPIServiceUpdate_OmittedUpstream_KeepsSecret_Integration proves, against a
// real DB, that updating a REST API without the upstream block does not
// deprecate or free the referenced secret.
func TestAPIServiceUpdate_OmittedUpstream_KeepsSecret_Integration(t *testing.T) {
	apiSvc, secretSvc, cleanup := setupAPISecretTestEnv(t)
	defer cleanup()

	createTestSecret(t, secretSvc, apiSecretITOrgUUID, "keep-me", "sk-real-token")

	createReq := &api.CreateRESTAPIRequest{
		DisplayName: "Keep Secret API",
		Context:     "/keep-secret-api",
		Version:     "v1",
		ProjectId:   "api-secret-it-proj",
		Upstream: api.Upstream{
			Main: api.UpstreamDefinition{
				Url: utils.StringPtrIfNotEmpty("https://backend.internal/api"),
				Auth: &api.UpstreamAuth{
					Type:   upstreamAuthTypePtr("bearer"),
					Header: ptr("Authorization"),
					Value:  ptr(`{{ secret "keep-me" }}`),
				},
			},
		},
	}
	created, err := apiSvc.CreateAPI(createReq, apiSecretITOrgUUID, "alice")
	if err != nil {
		t.Fatalf("CreateAPI failed: %v", err)
	}

	updateReq := &api.RESTAPI{DisplayName: "Renamed API"} // no upstream block
	if _, err := apiSvc.UpdateAPIByHandle(*created.Id, updateReq, apiSecretITOrgUUID, "alice"); err != nil {
		t.Fatalf("UpdateAPI (display-name only) failed: %v", err)
	}

	secret, err := secretSvc.Get(apiSecretITOrgUUID, "keep-me")
	if err != nil {
		t.Fatalf("secret 'keep-me' was lost after an update that omitted upstream: %v", err)
	}
	if secret.Status != string(model.SecretStatusActive) {
		t.Errorf("expected secret to remain ACTIVE, got status=%q", secret.Status)
	}

	if err := secretSvc.Delete(apiSecretITOrgUUID, "keep-me", "alice"); err == nil {
		t.Fatal("secret is no longer referenced after an update that omitted upstream — credential was wiped")
	} else {
		var inUse *SecretInUseError
		if !errors.As(err, &inUse) {
			t.Errorf("expected SecretInUseError proving the reference survived, got: %v", err)
		}
	}
}

// TestAPIServiceUpdate_GenuineRotation_StillDeprecatesOldSecret_Integration is a
// control: a real credential swap must still deprecate the old secret, proving
// the preservation fix does not disable rotation cleanup.
func TestAPIServiceUpdate_GenuineRotation_StillDeprecatesOldSecret_Integration(t *testing.T) {
	apiSvc, secretSvc, cleanup := setupAPISecretTestEnv(t)
	defer cleanup()

	createTestSecret(t, secretSvc, apiSecretITOrgUUID, "rot-a", "sk-a")
	createTestSecret(t, secretSvc, apiSecretITOrgUUID, "rot-b", "sk-b")

	createReq := &api.CreateRESTAPIRequest{
		DisplayName: "Rotate API",
		Context:     "/rotate-api",
		Version:     "v1",
		ProjectId:   "api-secret-it-proj",
		Upstream: api.Upstream{
			Main: api.UpstreamDefinition{
				Url:  utils.StringPtrIfNotEmpty("https://backend.internal/api"),
				Auth: &api.UpstreamAuth{Type: upstreamAuthTypePtr("bearer"), Header: ptr("Authorization"), Value: ptr(`{{ secret "rot-a" }}`)},
			},
		},
	}
	created, err := apiSvc.CreateAPI(createReq, apiSecretITOrgUUID, "alice")
	if err != nil {
		t.Fatalf("CreateAPI failed: %v", err)
	}

	rotateReq := &api.RESTAPI{
		DisplayName: "Rotate API",
		Context:     "/rotate-api",
		Version:     "v1",
		Upstream: api.Upstream{
			Main: api.UpstreamDefinition{
				Url:  utils.StringPtrIfNotEmpty("https://backend.internal/api"),
				Auth: &api.UpstreamAuth{Type: upstreamAuthTypePtr("bearer"), Header: ptr("Authorization"), Value: ptr(`{{ secret "rot-b" }}`)},
			},
		},
	}
	if _, err := apiSvc.UpdateAPIByHandle(*created.Id, rotateReq, apiSecretITOrgUUID, "alice"); err != nil {
		t.Fatalf("UpdateAPI (rotate) failed: %v", err)
	}

	a, err := secretSvc.Get(apiSecretITOrgUUID, "rot-a")
	if err != nil {
		t.Fatalf("failed to fetch rot-a: %v", err)
	}
	if a.Status != string(model.SecretStatusDeprecated) {
		t.Errorf("expected rotated-away secret to be DEPRECATED, got %q", a.Status)
	}
	b, err := secretSvc.Get(apiSecretITOrgUUID, "rot-b")
	if err != nil {
		t.Fatalf("failed to fetch rot-b: %v", err)
	}
	if b.Status != string(model.SecretStatusActive) {
		t.Errorf("expected new secret to remain ACTIVE, got %q", b.Status)
	}
}

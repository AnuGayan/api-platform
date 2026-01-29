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

package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wso2/api-platform/gateway/gateway-controller/pkg/config"
	"github.com/wso2/api-platform/gateway/gateway-controller/pkg/constants"
	policyenginev1 "github.com/wso2/api-platform/sdk/gateway/policyengine/v1"
)

func TestInjectSystemPolicies(t *testing.T) {
	cfg := &config.Config{
		Analytics: config.AnalyticsConfig{
			Enabled: true,
		},
	}

	existingPolicies := []policyenginev1.PolicyInstance{
		{Name: "UserPolicy", Version: "v1"},
	}

	t.Run("Injection enabled", func(t *testing.T) {
		injected := InjectSystemPolicies(existingPolicies, cfg, nil)
		assert.Len(t, injected, 2)
		assert.Equal(t, constants.ANALYTICS_SYSTEM_POLICY_NAME, injected[0].Name)
		assert.Equal(t, "UserPolicy", injected[1].Name)
	})

	t.Run("Injection disabled", func(t *testing.T) {
		cfgDisabled := &config.Config{
			Analytics: config.AnalyticsConfig{
				Enabled: false,
			},
		}
		injected := InjectSystemPolicies(existingPolicies, cfgDisabled, nil)
		assert.Len(t, injected, 1)
		assert.Equal(t, "UserPolicy", injected[0].Name)
	})

	t.Run("Merge parameters", func(t *testing.T) {
		props := map[string]any{
			constants.ANALYTICS_SYSTEM_POLICY_NAME: map[string]interface{}{
				"key1": "val1",
			},
			SharedParamsKey: map[string]interface{}{
				"shared": "val",
			},
		}
		injected := InjectSystemPolicies(existingPolicies, cfg, props)
		assert.Len(t, injected, 2)
		params := injected[0].Parameters
		assert.Equal(t, "val1", params["key1"])
		assert.Equal(t, "val", params["shared"])
	})

	t.Run("Nil config", func(t *testing.T) {
		injected := InjectSystemPolicies(existingPolicies, nil, nil)
		assert.Equal(t, existingPolicies, injected)
	})
}

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

package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	api "github.com/wso2/api-platform/gateway/gateway-controller/pkg/api/generated"
)

func TestValidator_URLFriendlyName(t *testing.T) {
	validator := NewAPIValidator()

	tests := []struct {
		name        string
		apiName     string
		shouldError bool
		errorMsg    string
	}{
		{
			name:        "valid name with spaces",
			apiName:     "Weather API",
			shouldError: false,
		},
		{
			name:        "valid name with hyphens",
			apiName:     "Weather-API",
			shouldError: false,
		},
		{
			name:        "valid name with underscores",
			apiName:     "Weather_API",
			shouldError: false,
		},
		{
			name:        "valid name with dots",
			apiName:     "Weather.API",
			shouldError: false,
		},
		{
			name:        "valid name alphanumeric",
			apiName:     "WeatherAPI123",
			shouldError: false,
		},
		{
			name:        "invalid name with slash",
			apiName:     "Weather/API",
			shouldError: true,
			errorMsg:    "API display name must be URL-friendly",
		},
		{
			name:        "invalid name with question mark",
			apiName:     "Weather?API",
			shouldError: true,
			errorMsg:    "API display name must be URL-friendly",
		},
		{
			name:        "invalid name with ampersand",
			apiName:     "Weather&API",
			shouldError: true,
			errorMsg:    "API display name must be URL-friendly",
		},
		{
			name:        "invalid name with hash",
			apiName:     "Weather#API",
			shouldError: true,
			errorMsg:    "API display name must be URL-friendly",
		},
		{
			name:        "invalid name with percent",
			apiName:     "Weather%API",
			shouldError: true,
			errorMsg:    "API display name must be URL-friendly",
		},
		{
			name:        "invalid name with brackets",
			apiName:     "Weather[API]",
			shouldError: true,
			errorMsg:    "API display name must be URL-friendly",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			specUnion := api.APIConfiguration_Spec{}
			specUnion.FromAPIConfigData(api.APIConfigData{
				DisplayName: tt.apiName,
				Version:     "v1.0",
				Context:     "/test",
				Upstream: struct {
					Main    api.Upstream  `json:"main" yaml:"main"`
					Sandbox *api.Upstream `json:"sandbox,omitempty" yaml:"sandbox,omitempty"`
				}{
					Main: api.Upstream{
						Url: func() *string { s := "http://example.com"; return &s }(),
					},
				},
				Operations: []api.Operation{
					{Method: "GET", Path: "/test"},
				},
			})
			config := &api.APIConfiguration{
				ApiVersion: api.APIConfigurationApiVersionGatewayApiPlatformWso2Comv1alpha1,
				Kind:       api.RestApi,
				Spec:       specUnion,
			}

			errors := validator.Validate(config)

			// Check if we got errors when we expected them
			hasNameError := false
			for _, err := range errors {
				if err.Field == "spec.displayName" {
					hasNameError = true
					if tt.shouldError && tt.errorMsg != "" {
						if err.Message[:len(tt.errorMsg)] != tt.errorMsg {
							t.Errorf("Expected error message to start with '%s', got '%s'", tt.errorMsg, err.Message)
						}
					}
					break
				}
			}

			if tt.shouldError && !hasNameError {
				t.Errorf("Expected validation error for name '%s', but got none", tt.apiName)
			}

			if !tt.shouldError && hasNameError {
				t.Errorf("Did not expect validation error for name '%s', but got one", tt.apiName)
			}
		})
	}
}

func TestValidateAuthConfig_BothAuthDisabled_AllowsNoAuthMode(t *testing.T) {
	// Test that validation allows no-auth mode when both auth methods are disabled
	config := &Config{
		GatewayController: GatewayController{
			Auth: AuthConfig{
				Basic: BasicAuth{
					Enabled: false,
				},
				IDP: IDPConfig{
					Enabled: false,
				},
			},
		},
	}

	err := config.validateAuthConfig()
	assert.NoError(t, err)
}

func TestValidator_MissingFields(t *testing.T) {
	validator := NewAPIValidator()

	t.Run("Missing display name", func(t *testing.T) {
		specUnion := api.APIConfiguration_Spec{}
		specUnion.FromAPIConfigData(api.APIConfigData{
			Version: "v1",
			Context: "/test",
		})
		cfg := &api.APIConfiguration{Kind: api.RestApi, Spec: specUnion}
		errs := validator.Validate(cfg)
		assert.NotEmpty(t, errs)
	})

	t.Run("Missing version", func(t *testing.T) {
		specUnion := api.APIConfiguration_Spec{}
		specUnion.FromAPIConfigData(api.APIConfigData{
			DisplayName: "Test",
			Context:     "/test",
		})
		cfg := &api.APIConfiguration{Kind: api.RestApi, Spec: specUnion}
		errs := validator.Validate(cfg)
		assert.NotEmpty(t, errs)
	})

	t.Run("Missing context", func(t *testing.T) {
		specUnion := api.APIConfiguration_Spec{}
		specUnion.FromAPIConfigData(api.APIConfigData{
			DisplayName: "Test",
			Version:     "v1",
		})
		cfg := &api.APIConfiguration{Kind: api.RestApi, Spec: specUnion}
		errs := validator.Validate(cfg)
		assert.NotEmpty(t, errs)
	})

	t.Run("Valid Asyncwebsub", func(t *testing.T) {
		specUnion := api.APIConfiguration_Spec{}
		specUnion.FromWebhookAPIData(api.WebhookAPIData{
			Name:    "async-api",
			Version: "v1.0",
			Context: "/events",
			Servers: []api.Server{
				{Url: "http://hub:8080"},
			},
			Channels: []api.Channel{
				{Path: "/topic"},
			},
		})
		cfg := &api.APIConfiguration{
			ApiVersion: "gateway.api-platform.wso2.com/v1alpha1",
			Kind:       api.Asyncwebsub,
			Spec:       specUnion,
		}
		errs := validator.Validate(cfg)
		assert.Empty(t, errs)
	})

	t.Run("Invalid Asyncwebsub - missing channels", func(t *testing.T) {
		specUnion := api.APIConfiguration_Spec{}
		specUnion.FromWebhookAPIData(api.WebhookAPIData{
			Name:    "async-api",
			Version: "v1",
			Context: "/events",
		})
		cfg := &api.APIConfiguration{Kind: api.Asyncwebsub, Spec: specUnion}
		errs := validator.Validate(cfg)
		assert.NotEmpty(t, errs)
	})
}

func TestValidateAuthConfig_BasicAuthEnabled(t *testing.T) {
	// Test that validation passes when basic auth is enabled
	config := &Config{
		GatewayController: GatewayController{
			Auth: AuthConfig{
				Basic: BasicAuth{
					Enabled: true,
					Users: []AuthUser{
						{Username: "admin", Password: "pass", Roles: []string{"admin"}},
					},
				},
				IDP: IDPConfig{
					Enabled: false,
				},
			},
		},
	}

	err := config.validateAuthConfig()
	assert.NoError(t, err)
}

func TestValidateAuthConfig_IDPAuthEnabled(t *testing.T) {
	// Test that validation passes when IDP auth is enabled
	config := &Config{
		GatewayController: GatewayController{
			Auth: AuthConfig{
				Basic: BasicAuth{
					Enabled: false,
				},
				IDP: IDPConfig{
					Enabled: true,
					JWKSURL: "https://idp.example.com/jwks",
				},
			},
		},
	}

	err := config.validateAuthConfig()
	assert.NoError(t, err)
}

func TestValidateAuthConfig_BothAuthEnabled(t *testing.T) {
	// Test that validation passes when both auth methods are enabled
	config := &Config{
		GatewayController: GatewayController{
			Auth: AuthConfig{
				Basic: BasicAuth{
					Enabled: true,
					Users: []AuthUser{
						{Username: "admin", Password: "pass", Roles: []string{"admin"}},
					},
				},
				IDP: IDPConfig{
					Enabled: true,
					JWKSURL: "https://idp.example.com/jwks",
				},
			},
		},
	}

	err := config.validateAuthConfig()
	assert.NoError(t, err)
}

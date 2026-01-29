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
	api "github.com/wso2/api-platform/gateway/gateway-controller/pkg/api/generated"
)

func TestExtractNameVersion(t *testing.T) {
	t.Run("RestApi", func(t *testing.T) {
		specJSON := `{"displayName": "Test API", "version": "1.0.0"}`
		var spec api.APIConfiguration_Spec
		_ = spec.UnmarshalJSON([]byte(specJSON))
		cfg := api.APIConfiguration{Kind: api.RestApi, Spec: spec}

		name, version, err := ExtractNameVersion(cfg)
		assert.NoError(t, err)
		assert.Equal(t, "Test API", name)
		assert.Equal(t, "1.0.0", version)
	})

	t.Run("Asyncwebsub", func(t *testing.T) {
		specJSON := `{"name": "Async API", "version": "2.0.0"}`
		var spec api.APIConfiguration_Spec
		_ = spec.UnmarshalJSON([]byte(specJSON))
		cfg := api.APIConfiguration{Kind: api.Asyncwebsub, Spec: spec}

		name, version, err := ExtractNameVersion(cfg)
		assert.NoError(t, err)
		assert.Equal(t, "Async API", name)
		assert.Equal(t, "2.0.0", version)
	})

	t.Run("Unsupported", func(t *testing.T) {
		cfg := api.APIConfiguration{Kind: "Unknown"}
		_, _, err := ExtractNameVersion(cfg)
		assert.Error(t, err)
	})
}

func TestGetValueFromSourceConfig(t *testing.T) {
	source := map[string]interface{}{
		"kind": "Test",
		"spec": map[string]interface{}{
			"template": "openai",
			"nested": map[string]interface{}{
				"val": 123,
			},
		},
	}

	t.Run("Top level key", func(t *testing.T) {
		val, err := GetValueFromSourceConfig(source, "kind")
		assert.NoError(t, err)
		assert.Equal(t, "Test", val)
	})

	t.Run("Nested path", func(t *testing.T) {
		val, err := GetValueFromSourceConfig(source, "spec.template")
		assert.NoError(t, err)
		assert.Equal(t, "openai", val)

		val, err = GetValueFromSourceConfig(source, "spec.nested.val")
		assert.NoError(t, err)
		assert.Equal(t, float64(123), val) // JSON unmarshals ints to float64 by default in map[string]any
	})

	t.Run("Key not found", func(t *testing.T) {
		_, err := GetValueFromSourceConfig(source, "nonexistent")
		assert.Error(t, err)

		_, err = GetValueFromSourceConfig(source, "spec.missing")
		assert.Error(t, err)
	})

	t.Run("Invalid path", func(t *testing.T) {
		_, err := GetValueFromSourceConfig(source, "kind.something")
		assert.Error(t, err)
	})

	t.Run("Nil source", func(t *testing.T) {
		_, err := GetValueFromSourceConfig(nil, "key")
		assert.Error(t, err)
	})
}

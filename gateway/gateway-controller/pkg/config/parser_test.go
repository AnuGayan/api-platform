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

func TestParser(t *testing.T) {
	parser := NewParser()

	t.Run("ParseJSON", func(t *testing.T) {
		data := []byte(`{"apiVersion": "v1", "kind": "RestApi"}`)
		var cfg api.APIConfiguration
		err := parser.ParseJSON(data, &cfg)
		assert.NoError(t, err)
		assert.Equal(t, api.RestApi, cfg.Kind)
	})

	t.Run("ParseYAML", func(t *testing.T) {
		data := []byte("apiVersion: v1\nkind: RestApi")
		var cfg api.APIConfiguration
		err := parser.ParseYAML(data, &cfg)
		assert.NoError(t, err)
		assert.Equal(t, api.RestApi, cfg.Kind)
	})

	t.Run("ParseAPIConfigYAML", func(t *testing.T) {
		data := []byte("apiVersion: v1\nkind: RestApi\nmetadata:\n  name: test")
		var cfg api.APIConfiguration
		err := parser.ParseAPIConfigYAML(data, &cfg)
		assert.NoError(t, err)
		assert.Equal(t, "test", cfg.Metadata.Name)
	})

	t.Run("Parse with content type", func(t *testing.T) {
		data := []byte(`{"apiVersion": "v1", "kind": "RestApi"}`)
		var cfg api.APIConfiguration
		err := parser.Parse(data, "application/json", &cfg)
		assert.NoError(t, err)

		dataYAML := []byte("apiVersion: v1\nkind: RestApi")
		err = parser.Parse(dataYAML, "application/yaml", &cfg)
		assert.NoError(t, err)

		// Auto-detect
		err = parser.Parse(data, "", &cfg)
		assert.NoError(t, err)
		err = parser.Parse(dataYAML, "", &cfg)
		assert.NoError(t, err)
	})

	t.Run("Parse errors", func(t *testing.T) {
		var cfg api.APIConfiguration
		err := parser.ParseJSON([]byte("{invalid}"), &cfg)
		assert.Error(t, err)

		err = parser.ParseYAML([]byte("invalid: : yaml"), &cfg)
		assert.Error(t, err)

		err = parser.Parse([]byte("!!!"), "text/plain", &cfg)
		assert.Error(t, err)
	})
}

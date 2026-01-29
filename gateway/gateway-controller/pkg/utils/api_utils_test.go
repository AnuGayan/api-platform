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
	"archive/zip"
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestExtractYAMLFromZip(t *testing.T) {
	// Create a zip with a yaml file
	buf := new(bytes.Buffer)
	zw := zip.NewWriter(buf)

	yamlContent := "apiVersion: v1\nkind: RestApi"
	f, err := zw.Create("api.yaml")
	require.NoError(t, err)
	_, err = f.Write([]byte(yamlContent))
	require.NoError(t, err)

	err = zw.Close()
	require.NoError(t, err)

	logger := zaptest.NewLogger(t)
	service := &APIUtilsService{logger: logger}

	t.Run("Valid zip with YAML", func(t *testing.T) {
		got, err := service.ExtractYAMLFromZip(buf.Bytes())
		assert.NoError(t, err)
		assert.Equal(t, yamlContent, string(got))
	})

	t.Run("Zip without YAML", func(t *testing.T) {
		buf2 := new(bytes.Buffer)
		zw2 := zip.NewWriter(buf2)
		f2, _ := zw2.Create("test.txt")
		f2.Write([]byte("not yaml"))
		zw2.Close()

		_, err := service.ExtractYAMLFromZip(buf2.Bytes())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no YAML file found")
	})

	t.Run("Invalid zip data", func(t *testing.T) {
		_, err := service.ExtractYAMLFromZip([]byte("not a zip"))
		assert.Error(t, err)
	})
}

func TestMapToStruct(t *testing.T) {
	type TestStruct struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

	data := map[string]interface{}{
		"name":  "test",
		"value": 123,
	}

	var out TestStruct
	err := MapToStruct(data, &out)
	assert.NoError(t, err)
	assert.Equal(t, "test", out.Name)
	assert.Equal(t, 123, out.Value)

	t.Run("Invalid mapping", func(t *testing.T) {
		data["value"] = "not an int"
		err := MapToStruct(data, &out)
		assert.Error(t, err)
	})
}

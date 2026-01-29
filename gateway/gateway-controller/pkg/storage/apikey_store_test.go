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

package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wso2/api-platform/gateway/gateway-controller/pkg/models"
	"go.uber.org/zap/zaptest"
)

func TestAPIKeyStore(t *testing.T) {
	logger := zaptest.NewLogger(t)
	s := NewAPIKeyStore(logger)

	apiID := "api-1"
	apiKey := &models.APIKey{
		ID:     "key-1",
		Name:   "key-name",
		APIKey: "secret-value",
		APIId:  apiID,
		Status: models.APIKeyStatusActive,
	}

	t.Run("Store", func(t *testing.T) {
		s.Store(apiKey)
		assert.Equal(t, 1, s.Count())

		// Update
		apiKey.Status = models.APIKeyStatusRevoked
		s.Store(apiKey)
		assert.Equal(t, 1, s.Count())

		got, _ := s.GetByID("key-1")
		assert.Equal(t, models.APIKeyStatusRevoked, got.Status)
	})

	t.Run("GetByID", func(t *testing.T) {
		got, exists := s.GetByID("key-1")
		assert.True(t, exists)
		assert.Equal(t, apiKey, got)

		_, exists = s.GetByID("non-existent")
		assert.False(t, exists)
	})

	t.Run("GetByValue", func(t *testing.T) {
		got, exists := s.GetByValue("secret-value")
		assert.True(t, exists)
		assert.Equal(t, apiKey, got)
	})

	t.Run("GetByAPI", func(t *testing.T) {
		keys := s.GetByAPI(apiID)
		assert.Len(t, keys, 1)
		assert.Equal(t, apiKey, keys[0])
	})

	t.Run("GetAll", func(t *testing.T) {
		keys := s.GetAll()
		assert.Len(t, keys, 1)
	})

	t.Run("Revoke", func(t *testing.T) {
		assert.True(t, s.Revoke("secret-value"))
		got, _ := s.GetByID("key-1")
		assert.Equal(t, models.APIKeyStatusRevoked, got.Status)
		assert.False(t, s.Revoke("non-existent"))
	})

	t.Run("Version", func(t *testing.T) {
		v := s.IncrementResourceVersion()
		assert.Equal(t, int64(1), v)
		assert.Equal(t, int64(1), s.GetResourceVersion())
	})

	t.Run("RemoveByID", func(t *testing.T) {
		assert.True(t, s.RemoveByID("key-1"))
		assert.False(t, s.RemoveByID("key-1"))
		assert.Equal(t, 0, s.Count())
	})

	t.Run("RemoveByAPI", func(t *testing.T) {
		s.Store(apiKey)
		assert.Equal(t, 1, s.RemoveByAPI(apiID))
		assert.Equal(t, 0, s.Count())
	})
}

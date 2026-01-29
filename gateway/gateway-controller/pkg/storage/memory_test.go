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
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	api "github.com/wso2/api-platform/gateway/gateway-controller/pkg/api/generated"
	"github.com/wso2/api-platform/gateway/gateway-controller/pkg/models"
)

func TestConfigStoreCRUD(t *testing.T) {
	store := NewConfigStore()

	configID := uuid.New().String()
	handle := "test-api"
	// Create a valid Spec JSON for RestApi
	specJSON := `{"displayName": "Test API", "version": "1.0.0"}`
	var spec api.APIConfiguration_Spec
	_ = spec.UnmarshalJSON([]byte(specJSON))

	cfg := &models.StoredConfig{
		ID:   configID,
		Kind: string(api.RestApi),
		Configuration: api.APIConfiguration{
			Kind: api.RestApi,
			Metadata: api.Metadata{
				Name: handle,
			},
			Spec: spec,
		},
		Status:    models.StatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	t.Run("Add", func(t *testing.T) {
		err := store.Add(cfg)
		assert.NoError(t, err)

		// Duplicate add (handle)
		err = store.Add(cfg)
		assert.Error(t, err)
		assert.True(t, IsConflictError(err))
	})

	t.Run("Get", func(t *testing.T) {
		got, err := store.Get(configID)
		assert.NoError(t, err)
		assert.Equal(t, configID, got.ID)
	})

	t.Run("GetByHandle", func(t *testing.T) {
		got, err := store.GetByHandle(handle)
		assert.NoError(t, err)
		assert.Equal(t, configID, got.ID)
	})

	t.Run("GetByNameVersion", func(t *testing.T) {
		got, err := store.GetByNameVersion("Test API", "1.0.0")
		assert.NoError(t, err)
		assert.Equal(t, configID, got.ID)
	})

	t.Run("GetByKindNameAndVersion", func(t *testing.T) {
		got := store.GetByKindNameAndVersion(string(api.RestApi), "Test API", "1.0.0")
		assert.NotNil(t, got)
		assert.Equal(t, configID, got.ID)
	})

	t.Run("Update", func(t *testing.T) {
		cfg.Status = models.StatusDeployed
		err := store.Update(cfg)
		assert.NoError(t, err)

		got, err := store.Get(configID)
		assert.NoError(t, err)
		assert.Equal(t, models.StatusDeployed, got.Status)

		// Test handle change
		oldHandle := cfg.Configuration.Metadata.Name
		// Copy cfg
		newCfg := *cfg
		newCfg.Configuration.Metadata.Name = "new-handle"
		err = store.Update(&newCfg)
		assert.NoError(t, err)

		_, err = store.GetByHandle("new-handle")
		assert.NoError(t, err)
		_, err = store.GetByHandle(oldHandle)
		assert.Error(t, err)

		// Revert handle for subsequent tests
		newCfg.Configuration.Metadata.Name = oldHandle
		_ = store.Update(&newCfg)
	})

	t.Run("GetAll", func(t *testing.T) {
		configs := store.GetAll()
		assert.Len(t, configs, 1)
	})

	t.Run("GetAllByKind", func(t *testing.T) {
		configs := store.GetAllByKind(string(api.RestApi))
		assert.Len(t, configs, 1)
	})

	t.Run("GetByKindAndHandle", func(t *testing.T) {
		got := store.GetByKindAndHandle(string(api.RestApi), handle)
		assert.NotNil(t, got)
		assert.Equal(t, configID, got.ID)
	})

	t.Run("Delete", func(t *testing.T) {
		err := store.Delete(configID)
		assert.NoError(t, err)

		_, err = store.Get(configID)
		assert.Error(t, err)
	})

	t.Run("Asyncwebsub topics", func(t *testing.T) {
		asyncID := uuid.New().String()
		// Valid WebhookAPIData spec
		asyncSpecJSON := `{"name": "test-webhook", "context": "ctx", "version": "v1", "channels": [{"path": "/events"}]}`
		var asyncSpec api.APIConfiguration_Spec
		_ = asyncSpec.UnmarshalJSON([]byte(asyncSpecJSON))

		asyncCfg := &models.StoredConfig{
			ID:   asyncID,
			Kind: string(api.Asyncwebsub),
			Configuration: api.APIConfiguration{
				Kind: api.Asyncwebsub,
				Metadata: api.Metadata{
					Name: "async-handle",
				},
				Spec: asyncSpec,
			},
		}

		err := store.Add(asyncCfg)
		assert.NoError(t, err)
		assert.True(t, store.TopicManager.Contains(asyncID, "test-webhook_ctx_v1_events"))

		// Update topics
		asyncSpecJSON2 := `{"name": "test-webhook", "context": "ctx", "version": "v1", "channels": [{"path": "/new-events"}]}`
		_ = asyncSpec.UnmarshalJSON([]byte(asyncSpecJSON2))
		asyncCfg.Configuration.Spec = asyncSpec
		err = store.Update(asyncCfg)
		assert.NoError(t, err)
		assert.False(t, store.TopicManager.Contains(asyncID, "test-webhook_ctx_v1_events"))
		assert.True(t, store.TopicManager.Contains(asyncID, "test-webhook_ctx_v1_new-events"))

		// Update with no topic changes
		err = store.Update(asyncCfg)
		assert.NoError(t, err)
		assert.True(t, store.TopicManager.Contains(asyncID, "test-webhook_ctx_v1_new-events"))

		err = store.Delete(asyncID)
		assert.NoError(t, err)
		assert.False(t, store.TopicManager.Contains(asyncID, "test-webhook_ctx_v1_new-events"))
	})
}

func TestConfigStoreSnapshotVersion(t *testing.T) {
	store := NewConfigStore()

	assert.Equal(t, int64(0), store.GetSnapshotVersion())

	v := store.IncrementSnapshotVersion()
	assert.Equal(t, int64(1), v)
	assert.Equal(t, int64(1), store.GetSnapshotVersion())

	store.SetSnapshotVersion(10)
	assert.Equal(t, int64(10), store.GetSnapshotVersion())
}

func TestConfigStoreTemplateCRUD(t *testing.T) {
	store := NewConfigStore()

	templateID := uuid.New().String()
	handle := "openai-tmpl"
	template := &models.StoredLLMProviderTemplate{
		ID: templateID,
		Configuration: api.LLMProviderTemplate{
			Metadata: api.Metadata{
				Name: handle,
			},
		},
	}

	t.Run("AddTemplate", func(t *testing.T) {
		err := store.AddTemplate(template)
		assert.NoError(t, err)

		// Duplicate
		err = store.AddTemplate(template)
		assert.Error(t, err)
	})

	t.Run("GetTemplate", func(t *testing.T) {
		got, err := store.GetTemplate(templateID)
		assert.NoError(t, err)
		assert.Equal(t, templateID, got.ID)
	})

	t.Run("GetTemplateByHandle", func(t *testing.T) {
		got, err := store.GetTemplateByHandle(handle)
		assert.NoError(t, err)
		assert.Equal(t, templateID, got.ID)
	})

	t.Run("UpdateTemplate", func(t *testing.T) {
		err := store.UpdateTemplate(template)
		assert.NoError(t, err)
	})

	t.Run("GetAllTemplates", func(t *testing.T) {
		templates := store.GetAllTemplates()
		assert.Len(t, templates, 1)
	})

	t.Run("DeleteTemplate", func(t *testing.T) {
		err := store.DeleteTemplate(templateID)
		assert.NoError(t, err)

		_, err = store.GetTemplate(templateID)
		assert.Error(t, err)
	})
}

func TestConfigStoreAPIKeyCRUD(t *testing.T) {
	store := NewConfigStore()

	apiID := "api-1"
	apiKey := &models.APIKey{
		ID:     "key-1",
		Name:   "my-key",
		APIKey: "secret-key",
		APIId:  apiID,
	}

	t.Run("StoreAPIKey", func(t *testing.T) {
		err := store.StoreAPIKey(apiKey)
		assert.NoError(t, err)

		// Update existing
		apiKey.Name = "my-key" // same name
		err = store.StoreAPIKey(apiKey)
		assert.NoError(t, err)
	})

	t.Run("GetAPIKeyByKey", func(t *testing.T) {
		got, err := store.GetAPIKeyByKey("secret-key")
		assert.NoError(t, err)
		assert.Equal(t, "key-1", got.ID)
	})

	t.Run("GetAPIKeysByAPI", func(t *testing.T) {
		keys, err := store.GetAPIKeysByAPI(apiID)
		assert.NoError(t, err)
		assert.Len(t, keys, 1)
	})

	t.Run("GetAPIKeyByName", func(t *testing.T) {
		got, err := store.GetAPIKeyByName(apiID, "my-key")
		assert.NoError(t, err)
		assert.Equal(t, "key-1", got.ID)
	})

	t.Run("RemoveAPIKeyByName", func(t *testing.T) {
		err := store.RemoveAPIKeyByName(apiID, "my-key")
		assert.NoError(t, err)

		_, err = store.GetAPIKeyByName(apiID, "my-key")
		assert.Error(t, err)
	})

	// Re-store
	_ = store.StoreAPIKey(apiKey)

	t.Run("RemoveAPIKeysByAPI", func(t *testing.T) {
		err := store.RemoveAPIKeysByAPI(apiID)
		assert.NoError(t, err)

		keys, _ := store.GetAPIKeysByAPI(apiID)
		assert.Len(t, keys, 0)
	})

	// Re-store
	_ = store.StoreAPIKey(apiKey)

	t.Run("RemoveAPIKey", func(t *testing.T) {
		err := store.RemoveAPIKey("secret-key")
		assert.NoError(t, err)

		_, err = store.GetAPIKeyByKey("secret-key")
		assert.Error(t, err)
	})
}

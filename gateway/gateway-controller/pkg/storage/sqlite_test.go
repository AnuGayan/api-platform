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
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	api "github.com/wso2/api-platform/gateway/gateway-controller/pkg/api/generated"
	"github.com/wso2/api-platform/gateway/gateway-controller/pkg/metrics"
	"github.com/wso2/api-platform/gateway/gateway-controller/pkg/models"
	"go.uber.org/zap/zaptest"
)

func setupSQLite(t *testing.T) (*SQLiteStorage, func()) {
	metrics.Init()
	tmpDir, err := os.MkdirTemp("", "sqlite-test-*")
	require.NoError(t, err)

	dbPath := filepath.Join(tmpDir, "test.db")
	logger := zaptest.NewLogger(t)
	storage, err := NewSQLiteStorage(dbPath, logger)
	require.NoError(t, err)

	cleanup := func() {
		storage.Close()
		os.RemoveAll(tmpDir)
	}

	return storage, cleanup
}

func TestSQLiteConfigCRUD(t *testing.T) {
	storage, cleanup := setupSQLite(t)
	defer cleanup()

	configID := uuid.New().String()
	handle := "test-api"
	cfg := &models.StoredConfig{
		ID:   configID,
		Kind: string(api.RestApi),
		Configuration: api.APIConfiguration{
			Metadata: api.Metadata{
				Name: handle,
			},
			Spec: api.APIConfiguration_Spec{},
		},
		Status:    models.StatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	// Manually set displayName and version for testing since GetDisplayName and GetVersion use Spec
	// But let's see how models.StoredConfig is defined.
	// Actually StoredConfig has Configuration field.

	t.Run("SaveConfig", func(t *testing.T) {
		err := storage.SaveConfig(cfg)
		assert.NoError(t, err)

		// Duplicate save should fail
		err = storage.SaveConfig(cfg)
		assert.Error(t, err)
		assert.True(t, IsConflictError(err))
	})

	t.Run("GetConfig", func(t *testing.T) {
		got, err := storage.GetConfig(configID)
		assert.NoError(t, err)
		assert.Equal(t, configID, got.ID)
		assert.Equal(t, string(api.RestApi), got.Kind)
	})

	t.Run("GetConfigByHandle", func(t *testing.T) {
		got, err := storage.GetConfigByHandle(handle)
		assert.NoError(t, err)
		assert.Equal(t, configID, got.ID)
	})

	t.Run("GetConfigByNameVersion", func(t *testing.T) {
		// Note: displayName and version are empty because they are extracted from Spec which is empty in our test cfg
		got, err := storage.GetConfigByNameVersion("", "")
		assert.NoError(t, err)
		assert.Equal(t, configID, got.ID)
	})

	t.Run("UpdateConfig", func(t *testing.T) {
		cfg.Status = models.StatusDeployed
		err := storage.UpdateConfig(cfg)
		assert.NoError(t, err)

		got, err := storage.GetConfig(configID)
		assert.NoError(t, err)
		assert.Equal(t, models.StatusDeployed, got.Status)
	})

	t.Run("GetAllConfigs", func(t *testing.T) {
		configs, err := storage.GetAllConfigs()
		assert.NoError(t, err)
		assert.Len(t, configs, 1)
	})

	t.Run("GetAllConfigsByKind", func(t *testing.T) {
		configs, err := storage.GetAllConfigsByKind(string(api.RestApi))
		assert.NoError(t, err)
		assert.Len(t, configs, 1)

		configs, err = storage.GetAllConfigsByKind("non-existent")
		assert.NoError(t, err)
		assert.Len(t, configs, 0)
	})

	t.Run("DeleteConfig", func(t *testing.T) {
		err := storage.DeleteConfig(configID)
		assert.NoError(t, err)

		_, err = storage.GetConfig(configID)
		assert.Error(t, err)
		assert.True(t, IsNotFoundError(err))
	})
}

func TestSQLiteLLMTemplateCRUD(t *testing.T) {
	storage, cleanup := setupSQLite(t)
	defer cleanup()

	templateID := uuid.New().String()
	handle := "openai-template"
	template := &models.StoredLLMProviderTemplate{
		ID: templateID,
		Configuration: api.LLMProviderTemplate{
			Metadata: api.Metadata{
				Name: handle,
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	t.Run("SaveLLMProviderTemplate", func(t *testing.T) {
		err := storage.SaveLLMProviderTemplate(template)
		assert.NoError(t, err)

		// Duplicate
		err = storage.SaveLLMProviderTemplate(template)
		assert.Error(t, err)
	})

	t.Run("GetLLMProviderTemplate", func(t *testing.T) {
		got, err := storage.GetLLMProviderTemplate(templateID)
		assert.NoError(t, err)
		assert.Equal(t, templateID, got.ID)
	})

	t.Run("UpdateLLMProviderTemplate", func(t *testing.T) {
		template.UpdatedAt = time.Now()
		err := storage.UpdateLLMProviderTemplate(template)
		assert.NoError(t, err)
	})

	t.Run("GetAllLLMProviderTemplates", func(t *testing.T) {
		templates, err := storage.GetAllLLMProviderTemplates()
		assert.NoError(t, err)
		assert.Len(t, templates, 1)
	})

	t.Run("DeleteLLMProviderTemplate", func(t *testing.T) {
		err := storage.DeleteLLMProviderTemplate(templateID)
		assert.NoError(t, err)

		_, err = storage.GetLLMProviderTemplate(templateID)
		assert.Error(t, err)
	})
}

func TestSQLiteCertificateCRUD(t *testing.T) {
	storage, cleanup := setupSQLite(t)
	defer cleanup()

	certID := uuid.New().String()
	name := "test-cert"
	cert := &models.StoredCertificate{
		ID:          certID,
		Name:        name,
		Certificate: []byte("fake-cert-data"),
		Subject:     "CN=test",
		Issuer:      "CN=issuer",
		NotBefore:   time.Now(),
		NotAfter:    time.Now().Add(24 * time.Hour),
		CertCount:   1,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	t.Run("SaveCertificate", func(t *testing.T) {
		err := storage.SaveCertificate(cert)
		assert.NoError(t, err)
	})

	t.Run("GetCertificate", func(t *testing.T) {
		got, err := storage.GetCertificate(certID)
		assert.NoError(t, err)
		assert.Equal(t, certID, got.ID)
	})

	t.Run("GetCertificateByName", func(t *testing.T) {
		got, err := storage.GetCertificateByName(name)
		assert.NoError(t, err)
		assert.NotNil(t, got)
		assert.Equal(t, certID, got.ID)
	})

	t.Run("ListCertificates", func(t *testing.T) {
		certs, err := storage.ListCertificates()
		assert.NoError(t, err)
		assert.Len(t, certs, 1)
	})

	t.Run("DeleteCertificate", func(t *testing.T) {
		err := storage.DeleteCertificate(certID)
		assert.NoError(t, err)

		_, err = storage.GetCertificate(certID)
		assert.Error(t, err)
	})
}

func TestSQLiteAPIKeyCRUD(t *testing.T) {
	storage, cleanup := setupSQLite(t)
	defer cleanup()

	// We need a deployment first because of foreign key constraint
	apiID := uuid.New().String()
	deployment := &models.StoredConfig{
		ID:   apiID,
		Kind: string(api.RestApi),
		Configuration: api.APIConfiguration{
			Metadata: api.Metadata{
				Name: "api-for-key",
			},
		},
		Status:    models.StatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err := storage.SaveConfig(deployment)
	require.NoError(t, err)

	apiKeyID := uuid.New().String()
	keyValue := "test-api-key-value"
	apiKey := &models.APIKey{
		ID:         apiKeyID,
		Name:       "my-key",
		APIKey:     keyValue,
		APIId:      apiID,
		Operations: "*",
		Status:     models.APIKeyStatusActive,
		CreatedAt:  time.Now(),
		CreatedBy:  "user",
		UpdatedAt:  time.Now(),
	}

	t.Run("SaveAPIKey", func(t *testing.T) {
		err := storage.SaveAPIKey(apiKey)
		assert.NoError(t, err)
	})

	t.Run("GetAPIKeyByKey", func(t *testing.T) {
		got, err := storage.GetAPIKeyByKey(keyValue)
		assert.NoError(t, err)
		assert.Equal(t, apiKeyID, got.ID)
	})

	t.Run("GetAPIKeysByAPI", func(t *testing.T) {
		keys, err := storage.GetAPIKeysByAPI(apiID)
		assert.NoError(t, err)
		assert.Len(t, keys, 1)
	})

	t.Run("GetAllAPIKeys", func(t *testing.T) {
		keys, err := storage.GetAllAPIKeys()
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(keys), 1)
	})

	t.Run("UpdateAPIKey", func(t *testing.T) {
		apiKey.Status = models.APIKeyStatusRevoked
		err := storage.UpdateAPIKey(apiKey)
		assert.NoError(t, err)

		got, err := storage.GetAPIKeyByKey(keyValue)
		assert.NoError(t, err)
		assert.Equal(t, models.APIKeyStatusRevoked, got.Status)
	})

	t.Run("GetAPIKeysByAPIAndName", func(t *testing.T) {
		got, err := storage.GetAPIKeysByAPIAndName(apiID, "my-key")
		assert.NoError(t, err)
		assert.Equal(t, apiKeyID, got.ID)
	})


	t.Run("RemoveAPIKeyAPIAndName", func(t *testing.T) {
		err := storage.RemoveAPIKeyAPIAndName(apiID, "my-key")
		assert.NoError(t, err)

		_, err = storage.GetAPIKeysByAPIAndName(apiID, "my-key")
		assert.Error(t, err)
	})

	// Re-add for next test
	err = storage.SaveAPIKey(apiKey)
	require.NoError(t, err)

	t.Run("RemoveAPIKeysAPI", func(t *testing.T) {
		err := storage.RemoveAPIKeysAPI(apiID)
		assert.NoError(t, err)

		keys, err := storage.GetAPIKeysByAPI(apiID)
		assert.NoError(t, err)
		assert.Len(t, keys, 0)
	})

	// Re-add for next test
	err = storage.SaveAPIKey(apiKey)
	require.NoError(t, err)

	t.Run("DeleteAPIKey", func(t *testing.T) {
		err := storage.DeleteAPIKey(keyValue)
		assert.NoError(t, err)

		_, err = storage.GetAPIKeyByKey(keyValue)
		assert.Error(t, err)
	})
}

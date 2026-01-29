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
	"github.com/wso2/api-platform/gateway/gateway-controller/pkg/metrics"
	"github.com/wso2/api-platform/gateway/gateway-controller/pkg/storage"
	"github.com/wso2/api-platform/gateway/gateway-controller/pkg/xds"
	"go.uber.org/zap/zaptest"
)

func TestAPIDeploymentService(t *testing.T) {
	metrics.Init()
	store := storage.NewConfigStore()
	validator := config.NewAPIValidator()
	logger := zaptest.NewLogger(t)
	routerConfig := &config.RouterConfig{
		GatewayHost: "localhost",
		VHosts: config.VHostsConfig{
			Main:    config.VHostEntry{Default: "localhost"},
			Sandbox: config.VHostEntry{Default: "sandbox.localhost"},
		},
	}
	systemConfig := &config.Config{
		GatewayController: config.GatewayController{
			Router: *routerConfig,
		},
	}
	snapshotManager := xds.NewSnapshotManager(store, logger, routerConfig, nil, systemConfig)
	service := NewAPIDeploymentService(store, nil, snapshotManager, validator, routerConfig)

	t.Run("Deploy RestApi", func(t *testing.T) {
		yaml := `
apiVersion: gateway.api-platform.wso2.com/v1alpha1
kind: RestApi
metadata:
  name: test-rest
spec:
  displayName: Test REST
  version: v1
  context: /test
  upstream:
    main:
      url: http://example.com
  operations:
    - method: GET
      path: /test
`
		params := APIDeploymentParams{
			Data:        []byte(yaml),
			ContentType: "application/yaml",
			Logger:      logger,
		}

		// This will panic or fail because snapshotManager is nil and it tries to call it in a goroutine
		// Actually, let's mock snapshotManager if needed, but it's called in a goroutine.
		// Wait, the goroutine might run after the test finishes if not careful.

		res, err := service.DeployAPIConfiguration(params)
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.False(t, res.IsUpdate)
	})

	t.Run("Deploy Duplicate", func(t *testing.T) {
		yaml := `
apiVersion: gateway.api-platform.wso2.com/v1alpha1
kind: RestApi
metadata:
  name: test-rest
spec:
  displayName: Test REST
  version: v1
  context: /test
`
		params := APIDeploymentParams{
			Data:        []byte(yaml),
			ContentType: "application/yaml",
			Logger:      logger,
		}
		_, err := service.DeployAPIConfiguration(params)
		assert.Error(t, err)
	})
}

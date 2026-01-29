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

package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/wso2/api-platform/gateway/gateway-controller/pkg/metrics"
	"go.uber.org/zap/zaptest"
)

func TestMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zaptest.NewLogger(t)
	metrics.Init()

	t.Run("LoggingMiddleware", func(t *testing.T) {
		r := gin.New()
		r.Use(LoggingMiddleware(logger))
		r.GET("/test", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("MetricsMiddleware", func(t *testing.T) {
		r := gin.New()
		r.Use(MetricsMiddleware())
		r.GET("/test", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("ErrorHandlingMiddleware", func(t *testing.T) {
		r := gin.New()
		r.Use(ErrorHandlingMiddleware(logger))
		r.GET("/panic", func(c *gin.Context) {
			panic("test panic")
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/panic", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

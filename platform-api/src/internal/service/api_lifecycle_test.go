/*
 *  Copyright (c) 2025, WSO2 LLC. (http://www.wso2.org) All Rights Reserved.
 *
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *  http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 *
 */

package service

import (
	"platform-api/src/internal/model"
	"platform-api/src/internal/utils"
	"testing"
)

func TestCreateAPI(t *testing.T) {
	apiRepo := &mockAPIRepository{}
	projectRepo := &mockProjectRepo{
		project: &model.Project{ID: "proj-123", OrganizationID: "org-123"},
	}
	backendRepo := &mockBackendRepo{}
	devPortalRepo := &mockDevPortalRepo{
		devPortal: &model.DevPortal{UUID: "dp-uuid"},
	}

	service := &APIService{
		apiRepo:            apiRepo,
		projectRepo:        projectRepo,
		backendServiceRepo: backendRepo,
		devPortalRepo:      devPortalRepo,
		upstreamService:    NewUpstreamService(backendRepo),
		apiUtil:            &utils.APIUtil{},
	}

	req := &CreateAPIRequest{
		Name:      "New API",
		Context:   "/new",
		Version:   "v1",
		ProjectID: "proj-123",
	}

	api, err := service.CreateAPI(req, "org-123")
	if err != nil {
		t.Fatalf("CreateAPI failed: %v", err)
	}

	if api.Name != "New API" {
		t.Errorf("Expected name New API, got %s", api.Name)
	}
}

func TestGetAPIByUUID(t *testing.T) {
	apiRepo := &mockAPIRepository{}
	service := &APIService{
		apiRepo: apiRepo,
		apiUtil: &utils.APIUtil{},
	}

	api, err := service.GetAPIByUUID("existing-uuid", "org-123")
	if err != nil {
		t.Fatalf("GetAPIByUUID failed: %v", err)
	}

	if api.ID != "existing-handle" { // ID in DTO holds the handle
		t.Errorf("Expected ID existing-handle, got %s", api.ID)
	}

	_, err = service.GetAPIByUUID("non-existent", "org-123")
	if err == nil {
		t.Error("Expected error for non-existent API, got nil")
	}
}

func TestDeleteAPI(t *testing.T) {
	apiRepo := &mockAPIRepository{}
	service := &APIService{
		apiRepo: apiRepo,
	}

	err := service.DeleteAPI("existing-uuid", "org-123")
	if err != nil {
		t.Fatalf("DeleteAPI failed: %v", err)
	}

	err = service.DeleteAPI("non-existent", "org-123")
	if err == nil {
		t.Error("Expected error for non-existent API, got nil")
	}
}

func TestAddGatewaysToAPI(t *testing.T) {
	apiUUID := "existing-uuid"
	orgID := "org-123"
	gatewayID := "123e4567-e89b-12d3-a456-426614174001"

	apiRepo := &mockAPIRepository{}
	gatewayRepo := &mockGatewayRepo{
		gateway: &model.Gateway{ID: gatewayID, OrganizationID: orgID},
	}

	service := &APIService{
		apiRepo:     apiRepo,
		gatewayRepo: gatewayRepo,
		apiUtil:     &utils.APIUtil{},
	}

	resp, err := service.AddGatewaysToAPI(apiUUID, []string{gatewayID}, orgID)
	if err != nil {
		t.Fatalf("AddGatewaysToAPI failed: %v", err)
	}

	if resp.Count != 1 {
		t.Errorf("Expected 1 gateway, got %d", resp.Count)
	}
}

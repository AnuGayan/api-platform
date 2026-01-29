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
	"testing"
)

func TestRegisterGateway(t *testing.T) {
	gatewayRepo := &mockGatewayRepo{}
	orgID := "123e4567-e89b-12d3-a456-426614174000"
	orgRepo := &mockOrgRepo{
		org: &model.Organization{ID: orgID},
	}
	apiRepo := &mockAPIRepository{}

	service := NewGatewayService(gatewayRepo, orgRepo, apiRepo)

	resp, err := service.RegisterGateway(orgID, "test-gateway", "Test Gateway", "Desc", "api.example.com", false, "regular")
	if err != nil {
		t.Fatalf("RegisterGateway failed: %v", err)
	}

	if resp.Name != "test-gateway" {
		t.Errorf("Expected name test-gateway, got %s", resp.Name)
	}

	if len(gatewayRepo.tokens) != 1 {
		t.Errorf("Expected 1 token to be created, got %d", len(gatewayRepo.tokens))
	}
}

func TestGetGateway(t *testing.T) {
	gatewayID := "123e4567-e89b-12d3-a456-426614174000"
	orgID := "123e4567-e89b-12d3-a456-426614174000"
	gatewayRepo := &mockGatewayRepo{
		gateway: &model.Gateway{ID: gatewayID, OrganizationID: orgID, Name: "test-gw"},
	}
	service := NewGatewayService(gatewayRepo, nil, nil)

	resp, err := service.GetGateway(gatewayID, orgID)
	if err != nil {
		t.Fatalf("GetGateway failed: %v", err)
	}

	if resp.ID != gatewayID {
		t.Errorf("Expected ID %s, got %s", gatewayID, resp.ID)
	}

	_, err = service.GetGateway(gatewayID, "wrong-org")
	if err == nil {
		t.Error("Expected error for wrong organization, got nil")
	}
}

func TestDeleteGateway(t *testing.T) {
	gatewayID := "123e4567-e89b-12d3-a456-426614174000"
	orgID := "123e4567-e89b-12d3-a456-426614174000"
	gatewayRepo := &mockGatewayRepo{
		gateway: &model.Gateway{ID: gatewayID, OrganizationID: orgID},
	}
	service := NewGatewayService(gatewayRepo, nil, nil)

	err := service.DeleteGateway(gatewayID, orgID)
	if err != nil {
		t.Fatalf("DeleteGateway failed: %v", err)
	}
}

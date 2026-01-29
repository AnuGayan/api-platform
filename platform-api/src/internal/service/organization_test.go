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

func TestRegisterOrganization(t *testing.T) {
	orgRepo := &mockOrgRepo{}
	projectRepo := &mockProjectRepo{}
	service := &OrganizationService{
		orgRepo:     orgRepo,
		projectRepo: projectRepo,
	}

	id := "org-1"
	handle := "acme"
	name := "ACME Corp"
	region := "US"

	org, err := service.RegisterOrganization(id, handle, name, region)
	if err != nil {
		t.Fatalf("RegisterOrganization failed: %v", err)
	}

	if org.ID != id {
		t.Errorf("Expected ID %s, got %s", id, org.ID)
	}
	if org.Handle != handle {
		t.Errorf("Expected handle %s, got %s", handle, org.Handle)
	}

	// Test invalid handle
	_, err = service.RegisterOrganization(id, "Invalid Handle!", name, region)
	if err == nil {
		t.Error("Expected error for invalid handle, got nil")
	}

	// Test duplicate
	orgRepo.org = &model.Organization{ID: id, Handle: handle}
	_, err = service.RegisterOrganization(id, handle, name, region)
	if err == nil {
		t.Error("Expected error for duplicate organization, got nil")
	}
}

func TestGetOrganizationByUUID(t *testing.T) {
	orgID := "org-1"
	orgRepo := &mockOrgRepo{
		org: &model.Organization{ID: orgID, Handle: "acme"},
	}
	service := &OrganizationService{
		orgRepo: orgRepo,
	}

	org, err := service.GetOrganizationByUUID(orgID)
	if err != nil {
		t.Fatalf("GetOrganizationByUUID failed: %v", err)
	}

	if org.ID != orgID {
		t.Errorf("Expected ID %s, got %s", orgID, org.ID)
	}

	_, err = service.GetOrganizationByUUID("non-existent")
	orgRepo.org = nil
	_, err = service.GetOrganizationByUUID("non-existent")
	if err == nil {
		t.Error("Expected error for non-existent organization, got nil")
	}
}

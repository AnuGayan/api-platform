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

func TestOrganizationService_RegisterOrganization_Advanced(t *testing.T) {
	orgRepo := &mockOrgRepo{}
	projectRepo := &mockProjectRepo{}
	service := &OrganizationService{
		orgRepo:     orgRepo,
		projectRepo: projectRepo,
	}

	// Test missing name
	org, err := service.RegisterOrganization("org-1", "handle", "", "US")
	if err != nil {
		t.Fatalf("RegisterOrganization failed: %v", err)
	}
	if org.Name != "handle" {
		t.Errorf("Expected name handle, got %s", org.Name)
	}

	// Test already exists by ID
	orgRepo.org = &model.Organization{ID: "org-1", Handle: "other"}
	_, err = service.RegisterOrganization("org-1", "handle", "Name", "US")
	if err == nil {
		t.Error("Expected error for existing ID, got nil")
	}

	// Test already exists by Handle
	orgRepo.org = &model.Organization{ID: "other", Handle: "handle"}
	_, err = service.RegisterOrganization("org-1", "handle", "Name", "US")
	if err == nil {
		t.Error("Expected error for existing handle, got nil")
	}
}

func TestProjectService_CreateProject_Advanced(t *testing.T) {
	projectRepo := &mockProjectRepo{}
	orgRepo := &mockOrgRepo{
		org: &model.Organization{ID: "org-1"},
	}
	service := NewProjectService(projectRepo, orgRepo, nil)

	// Test duplicate name
	projectRepo.project = &model.Project{ID: "proj-1", Name: "Duplicate", OrganizationID: "org-1"}
	_, err := service.CreateProject("Duplicate", "Desc", "org-1", "")
	if err == nil {
		t.Error("Expected error for duplicate project name, got nil")
	}

	// Test custom ID
	customID := "123e4567-e89b-12d3-a456-426614174000"
	p, err := service.CreateProject("Project", "Desc", "org-1", customID)
	if err != nil {
		t.Fatalf("CreateProject with custom ID failed: %v", err)
	}
	if p.ID != customID {
		t.Errorf("Expected ID %s, got %s", customID, p.ID)
	}

	// Test invalid custom ID
	_, err = service.CreateProject("Project", "Desc", "org-1", "invalid-uuid")
	if err == nil {
		t.Error("Expected error for invalid custom ID, got nil")
	}
}

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

func TestCreateProject(t *testing.T) {
	projectRepo := &mockProjectRepo{}
	orgRepo := &mockOrgRepo{
		org: &model.Organization{ID: "org-1"},
	}
	apiRepo := &mockAPIRepository{}
	service := NewProjectService(projectRepo, orgRepo, apiRepo)

	// Test Success
	project, err := service.CreateProject("My Project", "Description", "org-1", "")
	if err != nil {
		t.Fatalf("CreateProject failed: %v", err)
	}
	if project.Name != "My Project" {
		t.Errorf("Expected name My Project, got %s", project.Name)
	}

	// Test missing name
	_, err = service.CreateProject("", "Description", "org-1", "")
	if err == nil {
		t.Error("Expected error for missing name, got nil")
	}

	// Test missing org
	orgRepo.org = nil
	_, err = service.CreateProject("Project", "Description", "non-existent", "")
	if err == nil {
		t.Error("Expected error for non-existent org, got nil")
	}
}

func TestGetProjectByID(t *testing.T) {
	projectID := "proj-1"
	orgID := "org-1"
	projectRepo := &mockProjectRepo{
		project: &model.Project{ID: projectID, OrganizationID: orgID, Name: "Project"},
	}
	service := NewProjectService(projectRepo, nil, nil)

	project, err := service.GetProjectByID(projectID, orgID)
	if err != nil {
		t.Fatalf("GetProjectByID failed: %v", err)
	}
	if project.ID != projectID {
		t.Errorf("Expected ID %s, got %s", projectID, project.ID)
	}

	// Test wrong org
	_, err = service.GetProjectByID(projectID, "wrong-org")
	if err == nil {
		t.Error("Expected error for wrong organization, got nil")
	}
}

func TestUpdateProject(t *testing.T) {
	projectID := "proj-1"
	orgID := "org-1"
	projectRepo := &mockProjectRepo{
		project: &model.Project{ID: projectID, OrganizationID: orgID, Name: "Old Name"},
	}
	service := NewProjectService(projectRepo, nil, nil)

	updatedProject, err := service.UpdateProject(projectID, "New Name", "New Description", orgID)
	if err != nil {
		t.Fatalf("UpdateProject failed: %v", err)
	}
	if updatedProject.Name != "New Name" {
		t.Errorf("Expected name New Name, got %s", updatedProject.Name)
	}
}

func TestDeleteProject(t *testing.T) {
	projectID := "proj-1"
	orgID := "org-1"
	projectRepo := &mockProjectRepo{
		project: &model.Project{ID: projectID, OrganizationID: orgID},
	}
	apiRepo := &mockAPIRepository{}
	service := NewProjectService(projectRepo, nil, apiRepo)

	// In the real implementation, DeleteProject checks if it's the only project
	// and if there are associated APIs.
	// Our mock mockProjectRepo.GetProjectsByOrganizationID returns []*model.Project{m.project}

	// Test failure - only one project
	err := service.DeleteProject(projectID, orgID)
	if err == nil {
		t.Error("Expected error when deleting the only project, got nil")
	}
}

func TestGetProjectsByOrganization(t *testing.T) {
	orgID := "org-1"
	projectRepo := &mockProjectRepo{
		project: &model.Project{ID: "proj-1", OrganizationID: orgID},
	}
	orgRepo := &mockOrgRepo{
		org: &model.Organization{ID: orgID},
	}
	service := NewProjectService(projectRepo, orgRepo, nil)

	projects, err := service.GetProjectsByOrganization(orgID)
	if err != nil {
		t.Fatalf("GetProjectsByOrganization failed: %v", err)
	}
	if len(projects) != 1 {
		t.Errorf("Expected 1 project, got %d", len(projects))
	}
}

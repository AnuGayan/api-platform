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
	"platform-api/src/internal/dto"
	"platform-api/src/internal/model"
	"platform-api/src/internal/utils"
	"testing"
)

func TestUpdateAPI_Advanced(t *testing.T) {
	apiUUID := "existing-uuid"
	orgID := "org-123"
	apiRepo := &mockAPIRepository{}

	service := &APIService{
		apiRepo: apiRepo,
		apiUtil: &utils.APIUtil{},
		upstreamService: NewUpstreamService(&mockBackendRepo{}),
	}

	description := "New Description"
	status := "PUBLISHED"
	hasThumbnail := true
	req := &UpdateAPIRequest{
		Description:     &description,
		LifeCycleStatus: &status,
		HasThumbnail:    &hasThumbnail,
		BackendServices: &[]dto.BackendService{
			{Name: "be-1", Endpoints: []dto.BackendEndpoint{{URL: "http://be1"}}},
		},
	}

	_, err := service.UpdateAPI(apiUUID, req, orgID)
	if err != nil {
		t.Fatalf("UpdateAPI failed: %v", err)
	}
}

func TestAPIService_HandleExistsCheck(t *testing.T) {
	apiRepo := &mockAPIRepository{handleExistsResult: true}
	service := &APIService{apiRepo: apiRepo}

	check := service.HandleExistsCheck("org-1")
	if !check("handle") {
		t.Error("Expected HandleExistsCheck to return true")
	}
}

func TestAPIService_GetAPIsByOrganization(t *testing.T) {
	orgID := "org-1"
	apiRepo := &mockAPIRepository{}
	projectRepo := &mockProjectRepo{
		project: &model.Project{ID: "proj-1", OrganizationID: orgID},
	}
	service := &APIService{
		apiRepo:     apiRepo,
		projectRepo: projectRepo,
		apiUtil:     &utils.APIUtil{},
	}

	projectID := "proj-1"
	_, err := service.GetAPIsByOrganization(orgID, &projectID)
	if err != nil {
		t.Fatalf("GetAPIsByOrganization failed: %v", err)
	}
}

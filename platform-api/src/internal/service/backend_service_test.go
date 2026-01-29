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
	"testing"
)

func TestUpsertBackendService(t *testing.T) {
	repo := &mockBackendRepo{
		services: make(map[string]*model.BackendService),
	}
	service := NewUpstreamService(repo)

	orgID := "org-1"
	backendDTO := &dto.BackendService{
		Name:        "service-1",
		Description: "Initial description",
	}

	// Test Create
	id, err := service.UpsertBackendService(backendDTO, orgID)
	if err != nil {
		t.Fatalf("UpsertBackendService (Create) failed: %v", err)
	}

	if id == "" {
		t.Fatal("Expected non-empty ID")
	}

	// Test Update
	backendDTO.Description = "Updated description"
	id2, err := service.UpsertBackendService(backendDTO, orgID)
	if err != nil {
		t.Fatalf("UpsertBackendService (Update) failed: %v", err)
	}

	if id != id2 {
		t.Errorf("Expected same ID %s, got %s", id, id2)
	}

	if repo.services[id].Description != "Updated description" {
		t.Errorf("Description not updated. Got %s", repo.services[id].Description)
	}
}

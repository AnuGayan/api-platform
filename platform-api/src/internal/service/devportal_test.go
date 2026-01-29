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
	"platform-api/src/config"
	"platform-api/src/internal/dto"
	"platform-api/src/internal/model"
	"testing"
)

func TestCreateDevPortal(t *testing.T) {
	devPortalRepo := &mockDevPortalRepo{}
	orgRepo := &mockOrgRepo{
		org: &model.Organization{ID: "org-1", Handle: "acme", Name: "ACME Corp"},
	}
	cfg := &config.Server{}
	// Initialize default devportal config to avoid panics
	cfg.DefaultDevPortal.Timeout = 10

	service := &DevPortalService{
		devPortalRepo:      devPortalRepo,
		orgRepo:            orgRepo,
		config:             cfg,
		devPortalClientSvc: &mockDevPortalClientService{},
	}

	req := &dto.CreateDevPortalRequest{
		Name:       "My Portal",
		Identifier: "my-portal",
	}

	resp, err := service.CreateDevPortal("org-1", req)
	if err != nil {
		t.Fatalf("CreateDevPortal failed: %v", err)
	}
	if resp.Name != "My Portal" {
		t.Errorf("Expected name My Portal, got %s", resp.Name)
	}
}

func TestListDevPortals(t *testing.T) {
	devPortalRepo := &mockDevPortalRepo{
		devPortal: &model.DevPortal{UUID: "dp-1", Name: "Portal"},
	}
	service := &DevPortalService{
		devPortalRepo: devPortalRepo,
	}

	resp, err := service.ListDevPortals("org-1", nil, nil, 10, 0)
	if err != nil {
		t.Fatalf("ListDevPortals failed: %v", err)
	}
	if resp.Count != 1 {
		t.Errorf("Expected 1 portal, got %d", resp.Count)
	}
}

func TestDeleteDevPortal(t *testing.T) {
	devPortalRepo := &mockDevPortalRepo{}
	service := &DevPortalService{
		devPortalRepo: devPortalRepo,
	}

	err := service.DeleteDevPortal("dp-1", "org-1")
	if err != nil {
		t.Fatalf("DeleteDevPortal failed: %v", err)
	}
}

func TestEnableDevPortal(t *testing.T) {
	devPortalRepo := &mockDevPortalRepo{
		devPortal: &model.DevPortal{UUID: "dp-1", OrganizationUUID: "org-1", IsActive: false},
	}
	orgRepo := &mockOrgRepo{
		org: &model.Organization{ID: "org-1"},
	}
	service := &DevPortalService{
		devPortalRepo:      devPortalRepo,
		orgRepo:            orgRepo,
		devPortalClientSvc: &mockDevPortalClientService{},
	}

	err := service.EnableDevPortal("dp-1", "org-1")
	if err != nil {
		t.Fatalf("EnableDevPortal failed: %v", err)
	}
}

func TestUpdateDevPortal(t *testing.T) {
	devPortalRepo := &mockDevPortalRepo{
		devPortal: &model.DevPortal{UUID: "dp-1", Name: "Old Name"},
	}
	service := &DevPortalService{
		devPortalRepo: devPortalRepo,
	}

	newName := "New Name"
	req := &dto.UpdateDevPortalRequest{
		Name: &newName,
	}

	resp, err := service.UpdateDevPortal("dp-1", "org-1", req)
	if err != nil {
		t.Fatalf("UpdateDevPortal failed: %v", err)
	}
	if resp.Name != newName {
		t.Errorf("Expected name %s, got %s", newName, resp.Name)
	}
}

func TestCreateDefaultDevPortal(t *testing.T) {
	orgID := "org-1"
	devPortalRepo := &mockDevPortalRepo{}
	orgRepo := &mockOrgRepo{
		org: &model.Organization{ID: orgID, Handle: "acme", Name: "ACME Corp"},
	}
	cfg := &config.Server{}
	cfg.DefaultDevPortal.Enabled = true
	cfg.DefaultDevPortal.Name = "Default Portal"
	cfg.DefaultDevPortal.Timeout = 10

	service := &DevPortalService{
		devPortalRepo:      devPortalRepo,
		orgRepo:            orgRepo,
		config:             cfg,
		devPortalClientSvc: &mockDevPortalClientService{},
	}

	dp, err := service.CreateDefaultDevPortal(orgID)
	if err != nil {
		t.Fatalf("CreateDefaultDevPortal failed: %v", err)
	}
	if dp.Name != cfg.DefaultDevPortal.Name {
		t.Errorf("Expected name %s, got %s", cfg.DefaultDevPortal.Name, dp.Name)
	}
}

func TestGetDevPortal(t *testing.T) {
	devPortalRepo := &mockDevPortalRepo{
		devPortal: &model.DevPortal{UUID: "dp-1", Name: "Portal"},
	}
	service := &DevPortalService{
		devPortalRepo: devPortalRepo,
	}

	resp, err := service.GetDevPortal("dp-1", "org-1")
	if err != nil {
		t.Fatalf("GetDevPortal failed: %v", err)
	}
	if resp.UUID != "dp-1" {
		t.Errorf("Expected UUID dp-1, got %s", resp.UUID)
	}
}

func TestNewDevPortalService(t *testing.T) {
	service := NewDevPortalService(nil, nil, nil, nil, nil, nil)
	if service == nil {
		t.Error("NewDevPortalService returned nil")
	}
}

func TestDisableDevPortal(t *testing.T) {
	devPortalRepo := &mockDevPortalRepo{}
	service := &DevPortalService{
		devPortalRepo: devPortalRepo,
	}

	err := service.DisableDevPortal("dp-1", "org-1")
	if err != nil {
		t.Fatalf("DisableDevPortal failed: %v", err)
	}
}

func TestSetAsDefault(t *testing.T) {
	devPortalRepo := &mockDevPortalRepo{
		devPortal: &model.DevPortal{UUID: "dp-1"},
	}
	service := &DevPortalService{
		devPortalRepo: devPortalRepo,
	}

	err := service.SetAsDefault("dp-1", "org-1")
	if err != nil {
		t.Fatalf("SetAsDefault failed: %v", err)
	}
}

func TestGetDefaultDevPortal(t *testing.T) {
	devPortalRepo := &mockDevPortalRepo{
		devPortal: &model.DevPortal{UUID: "dp-1", IsDefault: true},
	}
	service := &DevPortalService{
		devPortalRepo: devPortalRepo,
	}

	resp, err := service.GetDefaultDevPortal("org-1")
	if err != nil {
		t.Fatalf("GetDefaultDevPortal failed: %v", err)
	}
	if resp.UUID != "dp-1" {
		t.Errorf("Expected UUID dp-1, got %s", resp.UUID)
	}
}

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
	"platform-api/src/internal/repository"
)

// mockAPIRepository is a mock implementation of the APIRepository interface
type mockAPIRepository struct {
	repository.APIRepository // Embed interface for unimplemented methods

	// Mock behavior configuration
	handleExistsResult      bool
	handleExistsError       error
	nameVersionExistsResult bool
	nameVersionExistsError  error

	// Call tracking for verification
	lastExcludeHandle string
}

func (m *mockAPIRepository) CheckAPIExistsByHandleInOrganization(handle, orgUUID string) (bool, error) {
	return m.handleExistsResult, m.handleExistsError
}

func (m *mockAPIRepository) CheckAPIExistsByNameAndVersionInOrganization(name, version, orgUUID, excludeHandle string) (bool, error) {
	m.lastExcludeHandle = excludeHandle // Track for verification
	return m.nameVersionExistsResult, m.nameVersionExistsError
}

func (m *mockAPIRepository) CreateAPI(api *model.API) error {
	api.ID = "generated-uuid"
	return nil
}

func (m *mockAPIRepository) CreateAPIAssociation(assoc *model.APIAssociation) error {
	return nil
}

func (m *mockAPIRepository) GetAPIByUUID(apiUUID, orgUUID string) (*model.API, error) {
	if apiUUID == "existing-uuid" {
		return &model.API{ID: "existing-uuid", OrganizationID: orgUUID, Handle: "existing-handle"}, nil
	}
	return nil, nil
}

func (m *mockAPIRepository) DeleteAPI(apiUUID, orgUUID string) error {
	return nil
}

func (m *mockAPIRepository) GetAPIAssociations(apiUUID, associationType, orgUUID string) ([]*model.APIAssociation, error) {
	return nil, nil
}

func (m *mockAPIRepository) GetAPIGatewaysWithDetails(apiUUID, orgUUID string) ([]*model.APIGatewayWithDetails, error) {
	if apiUUID == "existing-uuid" {
		return []*model.APIGatewayWithDetails{
			{
				ID:             "gw-1",
				OrganizationID: orgUUID,
				Name:           "gw-1",
			},
		}, nil
	}
	return nil, nil
}

func (m *mockAPIRepository) GetDeployedAPIsByGatewayUUID(gatewayUUID, orgUUID string) ([]*model.API, error) {
	return nil, nil
}

// mockProjectRepo is a mock implementation of the ProjectRepository interface
type mockProjectRepo struct {
	repository.ProjectRepository
	project *model.Project
	err     error
}

func (m *mockProjectRepo) GetProjectByUUID(id string) (*model.Project, error) {
	return m.project, m.err
}

// mockGatewayRepo is a mock implementation of the GatewayRepository interface
type mockGatewayRepo struct {
	repository.GatewayRepository
	gateway *model.Gateway
	tokens  []*model.GatewayToken
}

func (m *mockGatewayRepo) GetByNameAndOrgID(name, orgID string) (*model.Gateway, error) {
	if m.gateway != nil && m.gateway.Name == name && m.gateway.OrganizationID == orgID {
		return m.gateway, nil
	}
	return nil, nil
}

func (m *mockGatewayRepo) Create(gw *model.Gateway) error {
	m.gateway = gw
	return nil
}

func (m *mockGatewayRepo) CreateToken(t *model.GatewayToken) error {
	m.tokens = append(m.tokens, t)
	return nil
}

func (m *mockGatewayRepo) GetByUUID(id string) (*model.Gateway, error) {
	if m.gateway != nil && m.gateway.ID == id {
		return m.gateway, nil
	}
	return nil, nil
}

func (m *mockGatewayRepo) Delete(id, orgID string) error {
	return nil
}

func (m *mockGatewayRepo) HasGatewayAssociations(id, orgID string) (bool, error) {
	return false, nil
}

func (m *mockGatewayRepo) List() ([]*model.Gateway, error) {
	return nil, nil
}

func (m *mockGatewayRepo) GetActiveTokensByGatewayUUID(gatewayId string) ([]*model.GatewayToken, error) {
	return nil, nil
}

// mockOrgRepo is a mock implementation of the OrganizationRepository interface
type mockOrgRepo struct {
	repository.OrganizationRepository
	org *model.Organization
}

func (m *mockOrgRepo) GetOrganizationByUUID(id string) (*model.Organization, error) {
	return m.org, nil
}

// mockBackendRepo is a mock implementation of the BackendServiceRepository interface
type mockBackendRepo struct {
	repository.BackendServiceRepository
	services     map[string]*model.BackendService
	associateErr error
}

func (m *mockBackendRepo) GetBackendServiceByNameAndOrgID(name, orgID string) (*model.BackendService, error) {
	for _, s := range m.services {
		if s.Name == name && s.OrganizationID == orgID {
			return s, nil
		}
	}
	return nil, nil
}

func (m *mockBackendRepo) CreateBackendService(service *model.BackendService) error {
	if m.services == nil {
		m.services = make(map[string]*model.BackendService)
	}
	m.services[service.ID] = service
	return nil
}

func (m *mockBackendRepo) UpdateBackendService(service *model.BackendService) error {
	if m.services == nil {
		m.services = make(map[string]*model.BackendService)
	}
	m.services[service.ID] = service
	return nil
}

func (m *mockBackendRepo) AssociateBackendServiceWithAPI(apiId, backendServiceId string, isDefault bool) error {
	return m.associateErr
}

func (m *mockBackendRepo) GetBackendServicesByAPIID(apiId string) ([]*model.BackendService, error) {
	return nil, nil
}

// mockDevPortalRepo is a mock implementation of the DevPortalRepository interface
type mockDevPortalRepo struct {
	repository.DevPortalRepository
	devPortal *model.DevPortal
	err       error
}

func (m *mockDevPortalRepo) GetDefaultByOrganizationUUID(orgId string) (*model.DevPortal, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.devPortal, nil
}

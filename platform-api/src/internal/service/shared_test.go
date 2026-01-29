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
	"platform-api/src/internal/client/devportal_client"
	"platform-api/src/internal/dto"
	"platform-api/src/internal/model"
	"platform-api/src/internal/repository"
)

// mockDevPortalClientService is a mock implementation
type mockDevPortalClientService struct {
	syncErr error
}

func (m *mockDevPortalClientService) SyncOrganizationToDevPortal(devPortal *model.DevPortal, organization *model.Organization) error {
	return m.syncErr
}

func (m *mockDevPortalClientService) CreateDefaultSubscriptionPolicy(devPortal *model.DevPortal) error {
	return nil
}

func (m *mockDevPortalClientService) CreateDevPortalClient(devPortal *model.DevPortal) *devportal_client.DevPortalClient {
	return nil
}

func (m *mockDevPortalClientService) PublishAPIToDevPortal(client *devportal_client.DevPortalClient, orgID string, apiMetadata devportal_client.APIMetadataRequest, apiDefinition []byte) (*devportal_client.APIResponse, error) {
	return &devportal_client.APIResponse{ID: "ref-id"}, nil
}

func (m *mockDevPortalClientService) CheckAPIExists(client *devportal_client.DevPortalClient, orgID string, apiID string) (bool, error) {
	return false, nil
}

func (m *mockDevPortalClientService) UnpublishAPIFromDevPortal(client *devportal_client.DevPortalClient, orgID string, apiID string) error {
	return nil
}

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
		return &model.API{
			ID:             "existing-uuid",
			OrganizationID: orgUUID,
			Handle:         "existing-handle",
			Name:           "Existing API",
			Version:        "v1",
			Context:        "/existing",
		}, nil
	}
	return nil, nil
}

type mockGitService struct {
	ValidateAPIProjectResult *dto.APIProjectConfig
	ValidateAPIProjectErr    error
	FetchWSO2ArtifactResult  *dto.APIDeploymentYAML
	FetchWSO2ArtifactErr     error
	FetchFileContentResult   []byte
	FetchFileContentErr      error
}

func (m *mockGitService) FetchRepoBranches(repoURL string) (*dto.GitRepoBranchesResponse, error) {
	return &dto.GitRepoBranchesResponse{
		Branches: []dto.GitRepoBranch{{Name: "main"}},
	}, nil
}
func (m *mockGitService) FetchRepoContent(repoURL, branch string) (*dto.GitRepoContentResponse, error) {
	return &dto.GitRepoContentResponse{
		Items: []dto.GitRepoItem{{Path: "file.txt"}},
	}, nil
}
func (m *mockGitService) GetSupportedProviders() []string {
	return []string{"github"}
}
func (m *mockGitService) FetchFileContent(repoURL, branch, path string) ([]byte, error) {
	return m.FetchFileContentResult, m.FetchFileContentErr
}
func (m *mockGitService) ValidateAPIProject(repoURL, branch, path string) (*dto.APIProjectConfig, error) {
	return m.ValidateAPIProjectResult, m.ValidateAPIProjectErr
}
func (m *mockGitService) FetchWSO2Artifact(repoURL, branch, path string) (*dto.APIDeploymentYAML, error) {
	return m.FetchWSO2ArtifactResult, m.FetchWSO2ArtifactErr
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
	if gatewayUUID == "gw-1" {
		return []*model.API{{ID: "api-1", Handle: "handle-1", Name: "API 1"}}, nil
	}
	return nil, nil
}

func (m *mockAPIRepository) GetAPIsByProjectUUID(projectUUID, orgUUID string) ([]*model.API, error) {
	return nil, nil
}

func (m *mockAPIRepository) GetAPIsByOrganizationUUID(orgUUID string, projectUUID *string) ([]*model.API, error) {
	return nil, nil
}

func (m *mockAPIRepository) GetAPIMetadataByHandle(handle, orgUUID string) (*model.APIMetadata, error) {
	if handle == "existing-handle" {
		return &model.APIMetadata{ID: "existing-uuid", OrganizationID: orgUUID}, nil
	}
	return nil, nil
}

func (m *mockAPIRepository) UpdateAPI(api *model.API) error {
	return nil
}

func (m *mockAPIRepository) CreateDeployment(deployment *model.APIDeployment) error {
	return nil
}

func (m *mockAPIRepository) GetDeploymentsByAPIUUID(apiUUID, orgUUID string) ([]*model.APIDeployment, error) {
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

func (m *mockProjectRepo) CreateProject(project *model.Project) error {
	return m.err
}

func (m *mockProjectRepo) GetProjectsByOrganizationID(orgID string) ([]*model.Project, error) {
	if m.project == nil {
		return nil, m.err
	}
	return []*model.Project{m.project}, m.err
}

func (m *mockProjectRepo) UpdateProject(project *model.Project) error {
	return m.err
}

func (m *mockProjectRepo) DeleteProject(projectId string) error {
	return m.err
}

func (m *mockProjectRepo) GetProjectByNameAndOrgID(name, orgID string) (*model.Project, error) {
	if m.project != nil && m.project.Name == name {
		return m.project, nil
	}
	return nil, nil
}

func (m *mockProjectRepo) ListProjects(orgID string, limit, offset int) ([]*model.Project, error) {
	return []*model.Project{m.project}, m.err
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
	if m.gateway != nil {
		return []*model.Gateway{m.gateway}, nil
	}
	return nil, nil
}

func (m *mockGatewayRepo) GetActiveTokensByGatewayUUID(gatewayId string) ([]*model.GatewayToken, error) {
	return m.tokens, nil
}

func (m *mockGatewayRepo) GetByOrganizationID(orgID string) ([]*model.Gateway, error) {
	if m.gateway != nil && m.gateway.OrganizationID == orgID {
		return []*model.Gateway{m.gateway}, nil
	}
	return nil, nil
}

func (m *mockGatewayRepo) UpdateGateway(gateway *model.Gateway) error {
	return nil
}

func (m *mockGatewayRepo) UpdateActiveStatus(gatewayId string, isActive bool) error {
	return nil
}

func (m *mockGatewayRepo) CountActiveTokens(gatewayId string) (int, error) {
	return len(m.tokens), nil
}

// mockOrgRepo is a mock implementation of the OrganizationRepository interface
type mockOrgRepo struct {
	repository.OrganizationRepository
	org *model.Organization
}

func (m *mockOrgRepo) GetOrganizationByUUID(id string) (*model.Organization, error) {
	return m.org, nil
}

func (m *mockOrgRepo) GetOrganizationByIdOrHandle(id, handle string) (*model.Organization, error) {
	return m.org, nil
}

func (m *mockOrgRepo) CreateOrganization(org *model.Organization) error {
	return nil
}

func (m *mockOrgRepo) ListOrganizations(limit, offset int) ([]*model.Organization, error) {
	return []*model.Organization{m.org}, nil
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
	if apiId == "existing-uuid" {
		return []*model.BackendService{{ID: "be-1"}}, nil
	}
	return nil, nil
}

func (m *mockBackendRepo) DisassociateBackendServiceFromAPI(apiId, backendServiceId string) error {
	return nil
}

func (m *mockBackendRepo) GetBackendServicesByOrganizationID(orgID string) ([]*model.BackendService, error) {
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

func (m *mockDevPortalRepo) Create(devPortal *model.DevPortal) error {
	return nil
}

func (m *mockDevPortalRepo) GetByUUID(uuid, orgUUID string) (*model.DevPortal, error) {
	return m.devPortal, m.err
}

func (m *mockDevPortalRepo) GetByOrganizationUUID(orgUUID string, isDefault, isActive *bool, limit, offset int) ([]*model.DevPortal, error) {
	return []*model.DevPortal{m.devPortal}, m.err
}

func (m *mockDevPortalRepo) CountByOrganizationUUID(orgUUID string, isDefault, isActive *bool) (int, error) {
	return 1, nil
}

func (m *mockDevPortalRepo) Update(devPortal *model.DevPortal, orgUUID string) error {
	return nil
}

func (m *mockDevPortalRepo) Delete(uuid, orgUUID string) error {
	return nil
}

func (m *mockDevPortalRepo) UpdateEnabledStatus(uuid, orgUUID string, isEnabled bool) error {
	return nil
}

func (m *mockDevPortalRepo) SetAsDefault(uuid, orgUUID string) error {
	return nil
}

type mockPublicationRepo struct {
	repository.APIPublicationRepository
}

func (m *mockPublicationRepo) GetByAPIAndDevPortal(apiUUID, devPortalUUID, orgUUID string) (*model.APIPublication, error) {
	if apiUUID == "already-published-uuid" {
		return &model.APIPublication{APIUUID: apiUUID}, nil
	}
	return nil, nil
}

func (m *mockPublicationRepo) Create(publication *model.APIPublication) error {
	return nil
}

func (m *mockPublicationRepo) Delete(apiUUID, devPortalUUID, orgUUID string) error {
	return nil
}

func (m *mockPublicationRepo) GetAPIDevPortalsWithDetails(apiUUID, orgUUID string) ([]*model.APIDevPortalWithDetails, error) {
	if apiUUID == "existing-uuid" {
		return []*model.APIDevPortalWithDetails{
			{
				UUID: "dp-1",
				Name: "Portal 1",
			},
		}, nil
	}
	return nil, nil
}

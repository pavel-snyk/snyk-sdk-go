package snyk

import (
	"context"
	"fmt"
	"net/http"
)

const integrationBasePath = "org/%v/integrations"

// IntegrationsService handles communication with the integration related methods of the Snyk API.
type IntegrationsService service

const (
	ACRIntegrationType                 = "acr"
	ArtifactoryCRIntegrationType       = "artifactory-cr"
	AzureReposIntegrationType          = "azure-repos"
	BitBucketCloudIntegrationType      = "bitbucket-cloud"
	BitBucketConnectAppIntegrationType = "bitbucket-connect-app"
	BitBucketServerIntegrationType     = "bitbucket-server"
	DigitalOceanCRIntegrationType      = "digitalocean-cr"
	DockerHubIntegrationType           = "docker-hub"
	ECRIntegrationType                 = "ecr"
	GCRIntegrationType                 = "gcr"
	GitHubIntegrationType              = "github"
	GitHubCRIntegrationType            = "github-cr"
	GitHubEnterpriseIntegrationType    = "github-enterprise"
	GitLabIntegrationType              = "gitlab"
	GitLabCRIntegrationType            = "gitlab-cr"
	GoogleArtifactCRIntegrationType    = "google-artifact-cr"
	HarborCRIntegrationType            = "harbor-cr"
	KubernetesIntegrationType          = "kubernetes"
	NexusCRIntegrationType             = "nexus-cr"
	QuayCRIntegrationType              = "quay-cr"
)

// IntegrationType defines an integration type, e.g. "github" or "gitlab".
type IntegrationType string

type Integrations map[IntegrationType]string

// Integration represents a Snyk integration. Integrations are connections to places where code lives.
type Integration struct {
	Credentials *IntegrationCredentials `json:"credentials,omitempty"`
	ID          string                  `json:"id,omitempty"`
	Type        IntegrationType         `json:"type,omitempty"`
}

// IntegrationCredentials represents a credentials for the specific integration.
type IntegrationCredentials struct {
	Password     string `json:"password,omitempty"`
	Region       string `json:"region,omitempty"`
	RegistryBase string `json:"registryBase,omitempty"`
	RoleARN      string `json:"roleArn,omitempty"`
	Token        string `json:"token,omitempty"`
	URL          string `json:"url,omitempty"`
	Username     string `json:"username,omitempty"`
}

// IntegrationCreateRequest represents a request to create an integration.
type IntegrationCreateRequest struct {
	*Integration
}

// IntegrationUpdateRequest represents a request to update an integration.
type IntegrationUpdateRequest struct {
	*Integration
}

// List provides a list of all integrations for the given organization.
func (s *IntegrationsService) List(ctx context.Context, organizationID string) (Integrations, *Response, error) {
	if organizationID == "" {
		return nil, nil, ErrEmptyArgument
	}

	path := fmt.Sprintf(integrationBasePath, organizationID)

	req, err := s.client.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}

	integrations := new(Integrations)
	resp, err := s.client.Do(ctx, req, integrations)
	if err != nil {
		return nil, resp, err
	}

	return *integrations, resp, nil
}

// GetByType retrieves information about an integration identified by type.
func (s *IntegrationsService) GetByType(ctx context.Context, organizationID string, integrationType IntegrationType) (*Integration, *Response, error) {
	if organizationID == "" {
		return nil, nil, ErrEmptyArgument
	}
	if integrationType == "" {
		return nil, nil, ErrEmptyArgument
	}

	path := fmt.Sprintf(integrationBasePath+"/%v", organizationID, integrationType)

	req, err := s.client.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}

	integration := new(Integration)
	resp, err := s.client.Do(ctx, req, integration)
	if err != nil {
		return nil, resp, err
	}

	return integration, resp, nil
}

// Create makes a new integration with given payload.
func (s *IntegrationsService) Create(ctx context.Context, organizationID string, createRequest *IntegrationCreateRequest) (*Integration, *Response, error) {
	if organizationID == "" {
		return nil, nil, ErrEmptyArgument
	}
	if createRequest == nil {
		return nil, nil, ErrEmptyPayloadNotAllowed
	}

	path := fmt.Sprintf(integrationBasePath, organizationID)

	req, err := s.client.NewRequest(http.MethodPost, path, createRequest)
	if err != nil {
		return nil, nil, err
	}

	integration := new(Integration)
	resp, err := s.client.Do(ctx, req, integration)
	if err != nil {
		return nil, resp, err
	}

	return integration, resp, nil
}

// Update edits an integration identified by id.
func (s *IntegrationsService) Update(ctx context.Context, organizationID, integrationID string, updateRequest *IntegrationUpdateRequest) (*Integration, *Response, error) {
	if organizationID == "" {
		return nil, nil, ErrEmptyArgument
	}
	if integrationID == "" {
		return nil, nil, ErrEmptyArgument
	}
	if updateRequest == nil {
		return nil, nil, ErrEmptyPayloadNotAllowed
	}

	path := fmt.Sprintf(integrationBasePath+"/%v", organizationID, integrationID)

	req, err := s.client.NewRequest(http.MethodPut, path, updateRequest)
	if err != nil {
		return nil, nil, err
	}

	integration := new(Integration)
	resp, err := s.client.Do(ctx, req, integration)
	if err != nil {
		return nil, resp, err
	}

	return integration, resp, err
}

// DeleteCredentials removes any credentials set for the given integration.
// If this is a brokered connection the operation will have no effect.
func (s *IntegrationsService) DeleteCredentials(ctx context.Context, organizationID, integrationID string) (*Response, error) {
	if organizationID == "" {
		return nil, ErrEmptyArgument
	}
	if integrationID == "" {
		return nil, ErrEmptyArgument
	}

	path := fmt.Sprintf(integrationBasePath+"/%v/authentication", organizationID, integrationID)

	req, err := s.client.NewRequest(http.MethodDelete, path, nil)
	if err != nil {
		return nil, err
	}

	return s.client.Do(ctx, req, nil)
}

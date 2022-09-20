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

// IntegrationCredentials represents a credentials object for the specific integration.
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

// IntegrationSettings represents settings for the specific integration.
type IntegrationSettings struct {
	// DependencyAutoUpgradeEnabled can automatically raise pull requests to update out-of-date dependencies.
	DependencyAutoUpgradeEnabled *bool `json:"autoDepUpgradeEnabled,omitempty"`

	// DependencyAutoUpgradeIgnoredDependencies list of dependencies should be ignored.
	DependencyAutoUpgradeIgnoredDependencies []string `json:"autoDepUpgradeIgnoredDependencies,omitempty"`

	// DependencyAutoUpgradePullRequestLimit how many automatic dependency upgrade PRs can be opened simultaneously.
	DependencyAutoUpgradePullRequestLimit int64 `json:"autoDepUpgradeLimit,omitempty"`

	// DependencyAutoUpgradeIncludeMajorVersion includes major version in upgrade recommendation, otherwise it will be
	// minor and patch versions only.
	DependencyAutoUpgradeIncludeMajorVersion *bool `json:"isMajorUpgradeEnabled,omitempty"`

	// DockerfileDetectionEnabled will automatically detect and scan Dockerfiles in your Git repositories.
	DockerfileDetectionEnabled *bool `json:"dockerfileSCMEnabled,omitempty"`

	// PullRequestFailOnAnyIssue fails an opened pull request if any vulnerable dependencies have been detected,
	// otherwise the pull request should only fail when a dependency with issues is added.
	PullRequestFailOnAnyIssue *bool `json:"pullRequestFailOnAnyVulns,omitempty"`

	// PullRequestFailOnlyForIssuesWithFix fails an opened pull request only when issues found have a fix available.
	PullRequestFailOnlyForIssuesWithFix *bool `json:"pullRequestFailOnlyForIssuesWithFix,omitempty"`

	// PullRequestFailOnlyForHighAndCriticalSeverity fails an opened pull request if any dependencies are marked
	// as being of high or critical severity.
	PullRequestFailOnlyForHighAndCriticalSeverity *bool `json:"pullRequestFailOnlyForHighSeverity,omitempty"`

	// PullRequestTestEnabled tests any newly created pull request in your repositories for security vulnerabilities
	// and sends a status check to GitHub.
	//
	// Snyk docs: https://docs.snyk.io/integrations/git-repository-scm-integrations/github-integration#pull-request-testing
	PullRequestTestEnabled *bool `json:"pullRequestTestEnabled,omitempty"`
}

// IntegrationSettingsUpdateRequest represents a request to update an integration settings.
type IntegrationSettingsUpdateRequest struct {
	*IntegrationSettings
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

// Update edits an integration.
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

// GetSettings retrieves information about a settings for the given integration.
func (s *IntegrationsService) GetSettings(ctx context.Context, organizationID, integrationID string) (*IntegrationSettings, *Response, error) {
	if organizationID == "" {
		return nil, nil, ErrEmptyArgument
	}
	if integrationID == "" {
		return nil, nil, ErrEmptyArgument
	}

	path := fmt.Sprintf(integrationBasePath+"/%v/settings", organizationID, integrationID)

	req, err := s.client.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}

	settings := new(IntegrationSettings)
	resp, err := s.client.Do(ctx, req, settings)
	if err != nil {
		return nil, resp, err
	}

	return settings, resp, nil
}

// UpdateSettings edits an integration settings.
func (s *IntegrationsService) UpdateSettings(ctx context.Context, organizationID, integrationID string, updateRequest *IntegrationSettingsUpdateRequest) (*IntegrationSettings, *Response, error) {
	if organizationID == "" {
		return nil, nil, ErrEmptyArgument
	}
	if integrationID == "" {
		return nil, nil, ErrEmptyArgument
	}
	if updateRequest == nil {
		return nil, nil, ErrEmptyPayloadNotAllowed
	}

	path := fmt.Sprintf(integrationBasePath+"/%v/settings", organizationID, integrationID)

	req, err := s.client.NewRequest(http.MethodPut, path, updateRequest)
	if err != nil {
		return nil, nil, err
	}

	settings := new(IntegrationSettings)
	resp, err := s.client.Do(ctx, req, settings)
	if err != nil {
		return nil, resp, err
	}

	return settings, resp, nil
}

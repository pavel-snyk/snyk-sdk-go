package snyk

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

const (
	brokersBasePath           = "brokers"
	brokerDeploymentsBasePath = tenantsBasePath + "/%v/brokers/installs/%v/deployments"
	brokerDeploymentBasePath  = brokerDeploymentsBasePath + "/%v"
	brokerConnectionsBasePath = brokerDeploymentBasePath + "/connections"
	brokerConnectionBasePath  = brokerConnectionsBasePath + "/%v"
	brokersAPIVersion         = "2025-11-05"
)

// BrokersServiceAPI is an interface for interacting with the brokers endpoints of the Snyk API.
//
// See: https://docs.snyk.io/snyk-api/reference/universal-broker
type BrokersServiceAPI interface {
	// ListDeployments provides a list of broker deployments.
	//
	// See: https://docs.snyk.io/snyk-api/reference/universal-broker#get-tenants-tenant_id-brokers-installs-install_id-deployments
	ListDeployments(ctx context.Context, tenantID, appInstallID string) ([]BrokerDeployment, *Response, error)

	// CreateDeployment makes a new broker deployment.
	// "orgID" parameter in createRequest is the ID of organization where Universal Broker Snyk App is installed.
	//
	// See: https://docs.snyk.io/snyk-api/reference/universal-broker#post-tenants-tenant_id-brokers-installs-install_id-deployments
	CreateDeployment(ctx context.Context, tenantID, appInstallID string, createRequest *BrokerDeploymentCreateOrUpdateRequest) (*BrokerDeployment, *Response, error)

	// UpdateDeployment changes a broker deployment.
	//
	// See: https://docs.snyk.io/snyk-api/reference/universal-broker#patch-tenants-tenant_id-brokers-installs-install_id-deployments-deployment_id
	UpdateDeployment(ctx context.Context, tenantID, appInstallID, deploymentID string, updateRequest *BrokerDeploymentCreateOrUpdateRequest) (*BrokerDeployment, *Response, error)

	// DeleteDeployment removes a broker deployment.
	//
	// See: https://docs.snyk.io/snyk-api/reference/universal-broker#delete-tenants-tenant_id-brokers-installs-install_id-deployments-deployment_id
	DeleteDeployment(ctx context.Context, tenantID, appInstallID, deploymentID string) (*Response, error)

	// ListDeploymentCredentials provides a list of broker deployment credentials for a given deployment.
	//
	// See: https://docs.snyk.io/snyk-api/reference/universal-broker#get-tenants-tenant_id-brokers-installs-install_id-deployments-deployment_id-credentials
	ListDeploymentCredentials(ctx context.Context, tenantID, appInstallID, deploymentID string) ([]BrokerDeploymentCredential, *Response, error)

	// GetDeploymentCredential provides the full details of a broker deployment credential for a given deployment.
	//
	// See: https://docs.snyk.io/snyk-api/reference/universal-broker#get-tenants-tenant_id-brokers-installs-install_id-deployments-deployment_id-credentials-credential_i
	GetDeploymentCredential(ctx context.Context, tenantID, appInstallID, deploymentID, credentialID string) (*BrokerDeploymentCredential, *Response, error)

	// CreateDeploymentCredential makes a new broker deployment credential.
	//
	// See: https://docs.snyk.io/snyk-api/reference/universal-broker#post-tenants-tenant_id-brokers-installs-install_id-deployments-deployment_id-credentials
	CreateDeploymentCredential(ctx context.Context, tenantID, appInstallID, deploymentID string, createRequest *BrokerDeploymentCredentialCreateOrUpdateRequest) (*BrokerDeploymentCredential, *Response, error)

	// UpdateDeploymentCredential changes a broker deployment credential.
	//
	// See: https://docs.snyk.io/snyk-api/reference/universal-broker#patch-tenants-tenant_id-brokers-installs-install_id-deployments-deployment_id-credentials-credential
	UpdateDeploymentCredential(ctx context.Context, tenantID, appInstallID, deploymentID, credentialID string, updateRequest *BrokerDeploymentCredentialCreateOrUpdateRequest) (*BrokerDeploymentCredential, *Response, error)

	// DeleteDeploymentCredential removes a broker deployment credential.
	//
	// See: https://docs.snyk.io/snyk-api/reference/universal-broker#delete-tenants-tenant_id-brokers-installs-install_id-deployments-deployment_id-credentials-credentia
	DeleteDeploymentCredential(ctx context.Context, tenantID, appInstallID, deploymentID, credentialID string) (*Response, error)

	// ListConnections provides a list of broker connections for a given deployment.
	//
	// See: https://docs.snyk.io/snyk-api/reference/universal-broker#get-tenants-tenant_id-brokers-installs-install_id-deployments-deployment_id-connections
	ListConnections(ctx context.Context, tenantID, appInstallID, deploymentID string) ([]BrokerConnection, *Response, error)

	// GetConnection provides the full details of a broker connection for a given deployment.
	//
	// See: https://docs.snyk.io/snyk-api/reference/universal-broker#get-tenants-tenant_id-brokers-installs-install_id-deployments-deployment_id-connections-connection_i
	GetConnection(ctx context.Context, tenantID, appInstallID, deploymentID, connectionID string) (*BrokerConnection, *Response, error)

	// CreateConnection
	//
	// See: https://docs.snyk.io/snyk-api/reference/universal-broker#post-tenants-tenant_id-brokers-installs-install_id-deployments-deployment_id-connections
	CreateConnection(ctx context.Context, tenantID, appInstallID, deploymentID string, createRequest *BrokerConnectionCreateOrUpdateRequest) (*BrokerConnection, *Response, error)

	// UpdateConnection
	//
	// See: https://docs.snyk.io/snyk-api/reference/universal-broker#patch-tenants-tenant_id-brokers-installs-install_id-deployments-deployment_id-connections-connection
	UpdateConnection(ctx context.Context, tenantID, appInstallID, deploymentID, connectionID string, updateRequest *BrokerConnectionCreateOrUpdateRequest) (*BrokerConnection, *Response, error)

	// DeleteConnection
	//
	// See: https://docs.snyk.io/snyk-api/reference/universal-broker#delete-tenants-tenant_id-brokers-installs-install_id-deployments-deployment_id-connections-connectio
	DeleteConnection(ctx context.Context, tenantID, appInstallID, deploymentID, connectionID string) (*Response, error)
}

// BrokersService handles communication with the broker related methods of the Snyk API.
type BrokersService service

var _ BrokersServiceAPI = (*BrokersService)(nil)

// BrokerDeployment represents a Snyk broker deployment.
//
// See: https://docs.snyk.io/implementation-and-setup/enterprise-setup/snyk-broker/universal-broker/setting-up-and-integrating-your-universal-broker-connections#create-deployments-and-connections
type BrokerDeployment struct {
	ID         string                      `json:"id"`                   // The BrokerDeployment identifier.
	Type       string                      `json:"type"`                 // The resource type `broker_deployment`.
	Attributes *BrokerDeploymentAttributes `json:"attributes,omitempty"` // The BrokerDeployment resource data.
}

type BrokerDeploymentAttributes struct {
	AppInstallID string            `json:"install_id,omitempty"`                     // The ID of the Universal Broker AppInstall.
	OrgID        string            `json:"broker_app_installed_in_org_id,omitempty"` // The ID of the organization containing the AppInstall.
	Metadata     map[string]string `json:"metadata,omitempty"`                       // Metadata information as key/value.
}

type BrokerDeploymentCreateOrUpdateRequest struct {
	OrgID    string // The ID of the organization containing the Universal Broker Snyk AppInstall.
	Metadata map[string]string
}

type brokerDeploymentRoot struct {
	BrokerDeployment *BrokerDeployment `json:"data"`
}

type brokerDeploymentsRoot struct {
	BrokerDeployments []BrokerDeployment `json:"data"`
	Links             *PaginatedLinks    `json:"links,omitempty"`
}

func (d BrokerDeployment) String() string { return Stringify(d) }

func (s *BrokersService) ListDeployments(ctx context.Context, tenantID, appInstallID string) ([]BrokerDeployment, *Response, error) {
	if tenantID == "" {
		return nil, nil, errors.New("failed to list broker deployments: tenant id must be supplied")
	}
	if appInstallID == "" {
		return nil, nil, errors.New("failed to list broker deployments: app install id must be supplied")
	}

	opts := BaseOptions{Version: brokersAPIVersion}

	path, err := addOptions(fmt.Sprintf("%v/%v/%v/installs/%v/deployments", tenantsBasePath, tenantID, brokersBasePath, appInstallID), opts)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.prepareRequest(ctx, http.MethodGet, s.client.restBaseURL, path, nil)
	if err != nil {
		return nil, nil, err
	}

	root := new(brokerDeploymentsRoot)
	resp, err := s.client.do(ctx, req, &root)
	if err != nil {
		return nil, resp, err
	}
	if l := root.Links; l != nil {
		resp.Links = l
	}

	return root.BrokerDeployments, resp, nil
}

func (s *BrokersService) CreateDeployment(ctx context.Context, tenantID, appInstallID string, createRequest *BrokerDeploymentCreateOrUpdateRequest) (*BrokerDeployment, *Response, error) {
	if tenantID == "" {
		return nil, nil, errors.New("failed to create broker deployment: tenant id must be supplied")
	}
	if appInstallID == "" {
		return nil, nil, errors.New("failed to create broker deployment: app install id must be supplied")
	}
	if createRequest == nil {
		return nil, nil, errors.New("failed to create broker deployment: payload must be supplied")
	}

	opts := BaseOptions{Version: brokersAPIVersion}

	path, err := addOptions(fmt.Sprintf("%v/%v/%v/installs/%v/deployments", tenantsBasePath, tenantID, brokersBasePath, appInstallID), opts)
	if err != nil {
		return nil, nil, err
	}

	// inline jsonapi create payload to keep create function simple
	var createRequestJSON struct {
		Data struct {
			Attributes struct {
				OrgID    string       `json:"broker_app_installed_in_org_id"`
				Metadata *KeyValueMap `json:"metadata,omitempty"`
			} `json:"attributes"`
			Type string `json:"type"`
		} `json:"data"`
	}
	createRequestJSON.Data.Attributes.OrgID = createRequest.OrgID
	metadata := KeyValueMap{}
	if createRequest.Metadata != nil {
		metadata = createRequest.Metadata
	}
	createRequestJSON.Data.Attributes.Metadata = &metadata
	createRequestJSON.Data.Type = "broker_deployment"

	req, err := s.client.prepareRequest(ctx, http.MethodPost, s.client.restBaseURL, path, createRequestJSON)
	if err != nil {
		return nil, nil, err
	}

	root := new(brokerDeploymentRoot)
	resp, err := s.client.do(ctx, req, &root)
	if err != nil {
		return nil, resp, err
	}

	return root.BrokerDeployment, resp, nil
}

func (s *BrokersService) UpdateDeployment(ctx context.Context, tenantID, appInstallID, deploymentID string, updateRequest *BrokerDeploymentCreateOrUpdateRequest) (*BrokerDeployment, *Response, error) {
	if tenantID == "" {
		return nil, nil, errors.New("failed to update broker deployment: tenant id must be supplied")
	}
	if appInstallID == "" {
		return nil, nil, errors.New("failed to update broker deployment: app install id must be supplied")
	}
	if deploymentID == "" {
		return nil, nil, errors.New("failed to update broker deployment: id must be supplied")
	}
	if updateRequest == nil {
		return nil, nil, errors.New("failed to update broker deployment: payload must be supplied")
	}

	opts := BaseOptions{Version: brokersAPIVersion}

	path, err := addOptions(fmt.Sprintf("%v/%v/%v/installs/%v/deployments/%v", tenantsBasePath, tenantID, brokersBasePath, appInstallID, deploymentID), opts)
	if err != nil {
		return nil, nil, err
	}

	// inline jsonapi update payload to keep update function simple
	var updateRequestJSON struct {
		Data struct {
			Attributes struct {
				InstallID string       `json:"install_id"`
				OrgID     string       `json:"broker_app_installed_in_org_id"`
				Metadata  *KeyValueMap `json:"metadata,omitempty"`
			} `json:"attributes"`
			Type string `json:"type"`
		} `json:"data"`
	}
	updateRequestJSON.Data.Attributes.InstallID = appInstallID
	updateRequestJSON.Data.Attributes.OrgID = updateRequest.OrgID
	metadata := KeyValueMap{}
	if updateRequest.Metadata != nil {
		metadata = updateRequest.Metadata
	}
	updateRequestJSON.Data.Attributes.Metadata = &metadata
	updateRequestJSON.Data.Type = "broker_deployment"

	req, err := s.client.prepareRequest(ctx, http.MethodPatch, s.client.restBaseURL, path, updateRequestJSON)
	if err != nil {
		return nil, nil, err
	}

	root := new(brokerDeploymentRoot)
	resp, err := s.client.do(ctx, req, &root)
	if err != nil {
		return nil, resp, err
	}

	return root.BrokerDeployment, resp, nil
}

func (s *BrokersService) DeleteDeployment(ctx context.Context, tenantID, appInstallID, deploymentID string) (*Response, error) {
	if tenantID == "" {
		return nil, errors.New("failed to delete broker deployment: tenant id must be supplied")
	}
	if appInstallID == "" {
		return nil, errors.New("failed to delete broker deployment: app install id must be supplied")
	}
	if deploymentID == "" {
		return nil, errors.New("failed to delete broker deployment: id must be supplied")
	}

	opts := BaseOptions{Version: brokersAPIVersion}

	path, err := addOptions(fmt.Sprintf("%v/%v/%v/installs/%v/deployments/%v", tenantsBasePath, tenantID, brokersBasePath, appInstallID, deploymentID), opts)
	if err != nil {
		return nil, err
	}

	req, err := s.client.prepareRequest(ctx, http.MethodDelete, s.client.restBaseURL, path, nil)
	if err != nil {
		return nil, err
	}

	return s.client.do(ctx, req, nil)
}

// BrokerDeploymentCredential represents a Snyk broker deployment credential.
//
// See: https://docs.snyk.io/implementation-and-setup/enterprise-setup/snyk-broker/universal-broker/setting-up-and-integrating-your-universal-broker-connections#create-deployments-and-connections
type BrokerDeploymentCredential struct {
	ID         string                                `json:"id"`                   // The BrokerDeploymentCredential identifier.
	Type       string                                `json:"type"`                 // The resource type `deployment_credential`.
	Attributes *BrokerDeploymentCredentialAttributes `json:"attributes,omitempty"` // The BrokerDeploymentCredential resource data.
}

type BrokerDeploymentCredentialAttributes struct {
	Comment            string `json:"comment,omitempty"`                   // The comment information.
	BrokerDeploymentID string `json:"deployment_id,omitempty"`             // The ID of the associated BrokerDeployment.
	EnvVarName         string `json:"environment_variable_name,omitempty"` // The name of the environment variable.
	Type               string `json:"type,omitempty"`                      // The type of the associated BrokerConnection.
}

type BrokerDeploymentCredentialCreateOrUpdateRequest struct {
	Comment    string
	EnvVarName string
	Type       string
}

type brokerDeploymentCredentialRoot struct {
	BrokerDeploymentCredential *BrokerDeploymentCredential `json:"data"`
}

type brokerDeploymentCredentialsRoot struct {
	BrokerDeploymentCredentials []BrokerDeploymentCredential `json:"data"`
	Links                       *PaginatedLinks              `json:"links,omitempty"`
}

func (dc BrokerDeploymentCredential) String() string { return Stringify(dc) }

func (s *BrokersService) ListDeploymentCredentials(ctx context.Context, tenantID, appInstallID, deploymentID string) ([]BrokerDeploymentCredential, *Response, error) {
	if tenantID == "" {
		return nil, nil, errors.New("failed to list broker deployment credentials: tenant id must be supplied")
	}
	if appInstallID == "" {
		return nil, nil, errors.New("failed to list broker deployment credentials: app install id must be supplied")
	}
	if deploymentID == "" {
		return nil, nil, errors.New("failed to list broker deployment credentials: deployment id must be supplied")
	}

	opts := BaseOptions{Version: brokersAPIVersion}

	path, err := addOptions(fmt.Sprintf("%v/%v/%v/installs/%v/deployments/%v/credentials", tenantsBasePath, tenantID, brokersBasePath, appInstallID, deploymentID), opts)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.prepareRequest(ctx, http.MethodGet, s.client.restBaseURL, path, nil)
	if err != nil {
		return nil, nil, err
	}

	root := new(brokerDeploymentCredentialsRoot)
	resp, err := s.client.do(ctx, req, &root)
	if err != nil {
		return nil, resp, err
	}
	if l := root.Links; l != nil {
		resp.Links = l
	}

	return root.BrokerDeploymentCredentials, resp, nil
}

func (s *BrokersService) GetDeploymentCredential(ctx context.Context, tenantID, appInstallID, deploymentID, credentialID string) (*BrokerDeploymentCredential, *Response, error) {
	if tenantID == "" {
		return nil, nil, errors.New("failed to get broker deployment credential: tenant id must be supplied")
	}
	if appInstallID == "" {
		return nil, nil, errors.New("failed to get broker deployment credential: app install id must be supplied")
	}
	if deploymentID == "" {
		return nil, nil, errors.New("failed to get broker deployment credential: deployment id must be supplied")
	}
	if credentialID == "" {
		return nil, nil, errors.New("failed to get broker deployment credential: credential id must be supplied")
	}

	opts := BaseOptions{Version: brokersAPIVersion}

	path, err := addOptions(fmt.Sprintf("%v/%v/%v/installs/%v/deployments/%v/credentials/%v", tenantsBasePath, tenantID, brokersBasePath, appInstallID, deploymentID, credentialID), opts)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.prepareRequest(ctx, http.MethodGet, s.client.restBaseURL, path, nil)
	if err != nil {
		return nil, nil, err
	}

	root := new(brokerDeploymentCredentialRoot)
	resp, err := s.client.do(ctx, req, &root)
	if err != nil {
		return nil, resp, err
	}

	return root.BrokerDeploymentCredential, resp, nil
}

func (s *BrokersService) CreateDeploymentCredential(ctx context.Context, tenantID, appInstallID, deploymentID string, createRequest *BrokerDeploymentCredentialCreateOrUpdateRequest) (*BrokerDeploymentCredential, *Response, error) {
	if tenantID == "" {
		return nil, nil, errors.New("failed to create broker deployment credentials: tenant id must be supplied")
	}
	if appInstallID == "" {
		return nil, nil, errors.New("failed to create broker deployment credentials: app install id must be supplied")
	}
	if deploymentID == "" {
		return nil, nil, errors.New("failed to create broker deployment credentials: deployment id must be supplied")
	}
	if createRequest == nil {
		return nil, nil, errors.New("failed to create broker deployment credentials: payload must be supplied")
	}

	opts := BaseOptions{Version: brokersAPIVersion}

	path, err := addOptions(fmt.Sprintf("%v/%v/%v/installs/%v/deployments/%v/credentials", tenantsBasePath, tenantID, brokersBasePath, appInstallID, deploymentID), opts)
	if err != nil {
		return nil, nil, err
	}

	// inline jsonapi create payload to keep create function simple
	var createRequestJSON struct {
		Data struct {
			Attributes []struct {
				Comment    string `json:"comment,omitempty"`
				EnvVarName string `json:"environment_variable_name"`
				Type       string `json:"type"`
			} `json:"attributes"`
			Type string `json:"type"`
		} `json:"data"`
	}
	createRequestJSON.Data.Attributes = append(createRequestJSON.Data.Attributes, struct {
		Comment    string `json:"comment,omitempty"`
		EnvVarName string `json:"environment_variable_name"`
		Type       string `json:"type"`
	}{Comment: createRequest.Comment, EnvVarName: createRequest.EnvVarName, Type: createRequest.Type})
	createRequestJSON.Data.Type = "deployment_credential"

	req, err := s.client.prepareRequest(ctx, http.MethodPost, s.client.restBaseURL, path, createRequestJSON)
	if err != nil {
		return nil, nil, err
	}

	root := new(brokerDeploymentCredentialsRoot)
	resp, err := s.client.do(ctx, req, &root)
	if err != nil {
		return nil, resp, err
	}

	return &root.BrokerDeploymentCredentials[0], resp, nil

}

func (s *BrokersService) UpdateDeploymentCredential(ctx context.Context, tenantID, appInstallID, deploymentID, credentialID string, updateRequest *BrokerDeploymentCredentialCreateOrUpdateRequest) (*BrokerDeploymentCredential, *Response, error) {
	if tenantID == "" {
		return nil, nil, errors.New("failed to update broker deployment credential: tenant id must be supplied")
	}
	if appInstallID == "" {
		return nil, nil, errors.New("failed to update broker deployment credential: app install id must be supplied")
	}
	if deploymentID == "" {
		return nil, nil, errors.New("failed to update broker deployment credential: deployment id must be supplied")
	}
	if credentialID == "" {
		return nil, nil, errors.New("failed to update broker deployment credential: credential id must be supplied")
	}
	if updateRequest == nil {
		return nil, nil, errors.New("failed to update broker deployment credential: payload must be supplied")
	}

	opts := BaseOptions{Version: brokersAPIVersion}

	path, err := addOptions(fmt.Sprintf("%v/%v/%v/installs/%v/deployments/%v/credentials/%v", tenantsBasePath, tenantID, brokersBasePath, appInstallID, deploymentID, credentialID), opts)
	if err != nil {
		return nil, nil, err
	}

	// inline jsonapi update payload to keep update function simple
	var updateRequestJSON struct {
		Data struct {
			Attributes struct {
				Comment    string `json:"comment,omitempty"`
				EnvVarName string `json:"environment_variable_name"`
				Type       string `json:"type"`
			} `json:"attributes"`
			Type string `json:"type"`
		} `json:"data"`
	}
	updateRequestJSON.Data.Attributes.Comment = updateRequest.Comment
	updateRequestJSON.Data.Attributes.EnvVarName = updateRequest.EnvVarName
	updateRequestJSON.Data.Attributes.Type = updateRequest.Type
	updateRequestJSON.Data.Type = "deployment_credential"

	req, err := s.client.prepareRequest(ctx, http.MethodPatch, s.client.restBaseURL, path, updateRequestJSON)
	if err != nil {
		return nil, nil, err
	}

	root := new(brokerDeploymentCredentialRoot)
	resp, err := s.client.do(ctx, req, &root)
	if err != nil {
		return nil, resp, err
	}

	return root.BrokerDeploymentCredential, resp, nil
}

func (s *BrokersService) DeleteDeploymentCredential(ctx context.Context, tenantID, appInstallID, deploymentID, credentialID string) (*Response, error) {
	if tenantID == "" {
		return nil, errors.New("failed to delete broker deployment credential: tenant id must be supplied")
	}
	if appInstallID == "" {
		return nil, errors.New("failed to delete broker deployment credential: app install id must be supplied")
	}
	if deploymentID == "" {
		return nil, errors.New("failed to delete broker deployment credential: deployment id must be supplied")
	}
	if credentialID == "" {
		return nil, errors.New("failed to delete broker deployment credential: credential id must be supplied")
	}

	opts := BaseOptions{Version: brokersAPIVersion}

	path, err := addOptions(fmt.Sprintf("%v/%v/%v/installs/%v/deployments/%v/credentials/%v", tenantsBasePath, tenantID, brokersBasePath, appInstallID, deploymentID, credentialID), opts)
	if err != nil {
		return nil, err
	}

	req, err := s.client.prepareRequest(ctx, http.MethodDelete, s.client.restBaseURL, path, nil)
	if err != nil {
		return nil, err
	}

	return s.client.do(ctx, req, nil)
}

// BrokerConnection represents a Snyk broker connection.
//
// See: https://docs.snyk.io/implementation-and-setup/enterprise-setup/snyk-broker/universal-broker/setting-up-and-integrating-your-universal-broker-connections#create-deployments-and-connections
type BrokerConnection struct {
	ID         string                      `json:"id"`                   // The BrokerConnection identifier.
	Type       string                      `json:"type"`                 // The resource type `broker_connection`.
	Attributes *BrokerConnectionAttributes `json:"attributes,omitempty"` // The BrokerConnection resource data.
}

type BrokerConnectionAttributes struct {
	BrokerDeploymentID string                                   `json:"deployment_id,omitempty"` // The ID of the associated BrokerDeployment.
	Identifier         string                                   `json:"identifier"`              // The BrokerConnection identifier (known as broker token in Classic Broker).
	Name               string                                   `json:"name,omitempty"`          // The name of the BrokerConnection.
	Configuration      *BrokerConnectionAttributesConfiguration `json:"configuration"`           // The configuration of the BrokerConnection.
}

// BrokerConnectionType represents the type of BrokerConnection.
type BrokerConnectionType string

const (
	BrokerConnectionTypeArtifactory BrokerConnectionType = "artifactory"
	BrokerConnectionTypeGitLab      BrokerConnectionType = "gitlab"
	BrokerConnectionTypeGitHub      BrokerConnectionType = "github"
)

type BrokerConnectionAttributesConfiguration struct {
	Type        BrokerConnectionType `json:"type"` // The type of the BrokerConnection, e.g. 'gitlab' or 'github'.
	Artifactory *BrokerConnectionArtifactoryConfiguration
	GitHub      *BrokerConnectionGitHubConfiguration
	GitLab      *BrokerConnectionGitLabConfiguration
}

type BrokerConnectionArtifactoryConfiguration struct {
	ArtifactoryURL string `json:"artifactory_url"`
}

type BrokerConnectionCreateOrUpdateRequest struct {
	ArtifactoryURL  string
	BrokerClientURL string
	GitHubToken     string
	GitLabHostname  string
	GitLabToken     string
	Name            string
	Type            BrokerConnectionType
}

type BrokerConnectionGitHubConfiguration struct {
	BrokerClientURL string `json:"broker_client_url"`
	GitHubToken     string `json:"github_token"`
}

type BrokerConnectionGitLabConfiguration struct {
	BrokerClientURL string `json:"broker_client_url"`
	GitLabHostname  string `json:"gitlab"`
	GitLabToken     string `json:"gitlab_token"`
}

// UnmarshalJSON is a custom unmarshaller that handles the discriminator logic for BrokerConnectionAttributesConfiguration.
// It inspects the "type" field and unmarshals the "required" data into the corresponding typed struct.
func (c *BrokerConnectionAttributesConfiguration) UnmarshalJSON(data []byte) error {
	var raw struct {
		Type     BrokerConnectionType `json:"type"`
		Required json.RawMessage      `json:"required"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	c.Type = raw.Type

	switch c.Type {
	case BrokerConnectionTypeArtifactory:
		var artifactoryConfig BrokerConnectionArtifactoryConfiguration
		if err := json.Unmarshal(raw.Required, &artifactoryConfig); err != nil {
			return err
		}
		c.Artifactory = &artifactoryConfig
	case BrokerConnectionTypeGitHub:
		var githubConfig BrokerConnectionGitHubConfiguration
		if err := json.Unmarshal(raw.Required, &githubConfig); err != nil {
			return err
		}
		c.GitHub = &githubConfig
	case BrokerConnectionTypeGitLab:
		var gitlabConfig BrokerConnectionGitLabConfiguration
		if err := json.Unmarshal(raw.Required, &gitlabConfig); err != nil {
			return err
		}
		c.GitLab = &gitlabConfig
	}

	return nil
}

type brokerConnectionRoot struct {
	BrokerConnection *BrokerConnection `json:"data"`
}

type brokerConnectionsRoot struct {
	BrokerConnections []BrokerConnection `json:"data"`
	Links             *PaginatedLinks    `json:"links,omitempty"`
}

func (c BrokerConnection) String() string { return Stringify(c) }

func (s *BrokersService) ListConnections(ctx context.Context, tenantID, appInstallID, deploymentID string) ([]BrokerConnection, *Response, error) {
	if tenantID == "" {
		return nil, nil, errors.New("failed to list broker connections: tenant id must be supplied")
	}
	if appInstallID == "" {
		return nil, nil, errors.New("failed to list broker connections: app install id must be supplied")
	}
	if deploymentID == "" {
		return nil, nil, errors.New("failed to list broker connections: deployment id must be supplied")
	}

	opts := BaseOptions{Version: brokersAPIVersion}

	path, err := addOptions(fmt.Sprintf(brokerConnectionsBasePath, tenantID, appInstallID, deploymentID), opts)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.prepareRequest(ctx, http.MethodGet, s.client.restBaseURL, path, nil)
	if err != nil {
		return nil, nil, err
	}

	root := new(brokerConnectionsRoot)
	resp, err := s.client.do(ctx, req, &root)
	if err != nil {
		return nil, resp, err
	}
	if l := root.Links; l != nil {
		resp.Links = l
	}

	return root.BrokerConnections, resp, nil
}

func (s *BrokersService) GetConnection(ctx context.Context, tenantID, appInstallID, deploymentID, connectionID string) (*BrokerConnection, *Response, error) {
	if tenantID == "" {
		return nil, nil, errors.New("failed to get broker connection: tenant id must be supplied")
	}
	if appInstallID == "" {
		return nil, nil, errors.New("failed to get broker connection: app install id must be supplied")
	}
	if deploymentID == "" {
		return nil, nil, errors.New("failed to get broker connection: deployment id must be supplied")
	}
	if connectionID == "" {
		return nil, nil, errors.New("failed to get broker connection: connection id must be supplied")
	}

	opts := BaseOptions{Version: brokersAPIVersion}

	path, err := addOptions(fmt.Sprintf(brokerConnectionBasePath, tenantID, appInstallID, deploymentID, connectionID), opts)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.prepareRequest(ctx, http.MethodGet, s.client.restBaseURL, path, nil)
	if err != nil {
		return nil, nil, err
	}

	root := new(brokerConnectionRoot)
	resp, err := s.client.do(ctx, req, &root)
	if err != nil {
		return nil, resp, err
	}

	return root.BrokerConnection, resp, nil
}

func (s *BrokersService) CreateConnection(ctx context.Context, tenantID, appInstallID, deploymentID string, createRequest *BrokerConnectionCreateOrUpdateRequest) (*BrokerConnection, *Response, error) {
	if tenantID == "" {
		return nil, nil, errors.New("failed to create broker connection: tenant id must be supplied")
	}
	if appInstallID == "" {
		return nil, nil, errors.New("failed to create broker connection: app install id must be supplied")
	}
	if deploymentID == "" {
		return nil, nil, errors.New("failed to create broker connection: deployment id must be supplied")
	}
	if createRequest == nil {
		return nil, nil, errors.New("failed to create broker connection: payload must be supplied")
	}

	opts := BaseOptions{Version: brokersAPIVersion}

	path, err := addOptions(fmt.Sprintf(brokerConnectionsBasePath, tenantID, appInstallID, deploymentID), opts)
	if err != nil {
		return nil, nil, err
	}

	createPayload, err := buildBrokerConnectionRequestPayload(deploymentID, createRequest)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to build create broker connection request: %w", err)
	}

	req, err := s.client.prepareRequest(ctx, http.MethodPost, s.client.restBaseURL, path, createPayload)
	if err != nil {
		return nil, nil, err
	}

	root := new(brokerConnectionRoot)
	resp, err := s.client.do(ctx, req, &root)
	if err != nil {
		return nil, resp, err
	}

	return root.BrokerConnection, resp, nil
}

func (s *BrokersService) UpdateConnection(ctx context.Context, tenantID, appInstallID, deploymentID, connectionID string, updateRequest *BrokerConnectionCreateOrUpdateRequest) (*BrokerConnection, *Response, error) {
	if tenantID == "" {
		return nil, nil, errors.New("failed to update broker connection: tenant id must be supplied")
	}
	if appInstallID == "" {
		return nil, nil, errors.New("failed to update broker connection: app install id must be supplied")
	}
	if deploymentID == "" {
		return nil, nil, errors.New("failed to update broker connection: deployment id must be supplied")
	}
	if connectionID == "" {
		return nil, nil, errors.New("failed to update broker connection: connection id must be supplied")
	}
	if updateRequest == nil {
		return nil, nil, errors.New("failed to update broker connection: payload must be supplied")
	}

	opts := BaseOptions{Version: brokersAPIVersion}

	path, err := addOptions(fmt.Sprintf(brokerConnectionBasePath, tenantID, appInstallID, deploymentID, connectionID), opts)
	if err != nil {
		return nil, nil, err
	}

	updatePayload, err := buildBrokerConnectionRequestPayload(deploymentID, updateRequest)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to build update broker connection request: %w", err)
	}

	req, err := s.client.prepareRequest(ctx, http.MethodPatch, s.client.restBaseURL, path, updatePayload)
	if err != nil {
		return nil, nil, err
	}

	root := new(brokerConnectionRoot)
	resp, err := s.client.do(ctx, req, &root)
	if err != nil {
		return nil, resp, err
	}

	return root.BrokerConnection, resp, nil
}

func (s *BrokersService) DeleteConnection(ctx context.Context, tenantID, appInstallID, deploymentID, connectionID string) (*Response, error) {
	if tenantID == "" {
		return nil, errors.New("failed to delete broker connection: tenant id must be supplied")
	}
	if appInstallID == "" {
		return nil, errors.New("failed to delete broker connection: app install id must be supplied")
	}
	if deploymentID == "" {
		return nil, errors.New("failed to delete broker connection: deployment id must be supplied")
	}
	if connectionID == "" {
		return nil, errors.New("failed to delete broker connection: connection id must be supplied")
	}

	opts := BaseOptions{Version: brokersAPIVersion}

	path, err := addOptions(fmt.Sprintf(brokerConnectionBasePath, tenantID, appInstallID, deploymentID, connectionID), opts)
	if err != nil {
		return nil, err
	}

	req, err := s.client.prepareRequest(ctx, http.MethodDelete, s.client.restBaseURL, path, nil)
	if err != nil {
		return nil, err
	}

	return s.client.do(ctx, req, nil)
}

// buildBrokerConnectionRequestPayload converts, validates and prepares request payloads
// for BrokersService.CreateConnection and BrokersService.UpdateConnection() functions.
func buildBrokerConnectionRequestPayload(deploymentID string, request *BrokerConnectionCreateOrUpdateRequest) (any, error) {
	if request.Type == "" {
		return nil, errors.New("request.Type must be supplied")
	}

	type configurationJSON struct {
		Required struct {
			ArtifactoryURL  string `json:"artifactory_url,omitempty"`
			BrokerClientURL string `json:"broker_client_url,omitempty"`
			GitLabHostname  string `json:"gitlab,omitempty"`
			GitLabToken     string `json:"gitlab_token,omitempty"`
			GitHubToken     string `json:"github_token,omitempty"`
		} `json:"required"`
		Type BrokerConnectionType `json:"type"`
	}
	var requestJSON struct {
		Data struct {
			Attributes struct {
				Configuration configurationJSON `json:"configuration"`
				DeploymentID  string            `json:"deployment_id"`
				Name          string            `json:"name"`
			} `json:"attributes"`
			Type string `json:"type"`
		} `json:"data"`
	}

	var configuration configurationJSON
	configuration.Type = request.Type

	switch request.Type {
	case BrokerConnectionTypeArtifactory:
		if request.ArtifactoryURL == "" {
			return nil, errors.New("ArtifactoryURL must be supplied for artifactory connection type")
		}
		configuration.Required.ArtifactoryURL = request.ArtifactoryURL
	case BrokerConnectionTypeGitHub:
		if request.BrokerClientURL == "" {
			return nil, errors.New("BrokerClientURL must be supplied for github connection type")
		}
		if request.GitHubToken == "" {
			return nil, errors.New("GitHubToken must be supplied for github connection type")
		}
		configuration.Required.BrokerClientURL = request.BrokerClientURL
		configuration.Required.GitHubToken = request.GitHubToken
	case BrokerConnectionTypeGitLab:
		if request.BrokerClientURL == "" {
			return nil, errors.New("BrokerClientURL must be supplied for gitlab connection type")
		}
		if request.GitLabHostname == "" {
			return nil, errors.New("GitLabHostname must be supplied for gitlab connection type")
		}
		if request.GitLabToken == "" {
			return nil, errors.New("GitLabToken must be supplied for gitlab connection type")
		}
		configuration.Required.BrokerClientURL = request.BrokerClientURL
		configuration.Required.GitLabHostname = request.GitLabHostname
		configuration.Required.GitLabToken = request.GitLabToken
	default:
		return nil, fmt.Errorf("unsupported broker connection type: %s", request.Type)
	}

	requestJSON.Data.Attributes.Configuration = configuration
	requestJSON.Data.Attributes.DeploymentID = deploymentID
	requestJSON.Data.Attributes.Name = request.Name
	requestJSON.Data.Type = "broker_connection"

	return requestJSON, nil
}

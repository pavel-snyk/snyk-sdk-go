package snyk

import (
	"context"
	"errors"
	"fmt"
	"net/http"
)

const (
	brokersBasePath   = "brokers"
	brokersAPIVersion = "2025-11-05"
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

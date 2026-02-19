package snyk

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBrokers_ListDeployments(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/tenants/tenant-id/brokers/installs/install-id/deployments", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		_, _ = fmt.Fprint(w, `
{
  "jsonapi": { "version": "1.0" },
  "data": [
    {
      "id": "0779acb9-968e-4bff-abd7-94193e589028",
      "type": "broker_deployment",
      "attributes": {
        "install_id": "216b7774-9198-4fcd-a525-bace50228e18",
        "broker_app_installed_in_org_id": "6c6b3b6d-24e5-4f70-9896-4a49609cd61a",
        "metadata": {
          "region": "us-east-1"
        }
      }
    }
  ],
  "links": {}
}
`)
	})
	expectedDeployments := []BrokerDeployment{{
		ID:   "0779acb9-968e-4bff-abd7-94193e589028",
		Type: "broker_deployment",
		Attributes: &BrokerDeploymentAttributes{
			AppInstallID: "216b7774-9198-4fcd-a525-bace50228e18",
			OrgID:        "6c6b3b6d-24e5-4f70-9896-4a49609cd61a",
			Metadata:     map[string]string{"region": "us-east-1"},
		},
	}}

	actualDeployments, _, err := client.Brokers.ListDeployments(ctx, "tenant-id", "install-id")

	assert.NoError(t, err)
	assert.Equal(t, expectedDeployments, actualDeployments)
}

func TestBrokers_ListDeployments_emptyTenantID(t *testing.T) {
	_, _, err := client.Brokers.ListDeployments(ctx, "", "install-id")

	assert.Error(t, err)
	assert.ErrorContains(t, err, "tenant id must be supplied")
}

func TestBrokers_ListDeployments_emptyAppInstallID(t *testing.T) {
	_, _, err := client.Brokers.ListDeployments(ctx, "tenant-id", "")

	assert.Error(t, err)
	assert.ErrorContains(t, err, "install id must be supplied")
}

func TestBrokers_CreateDeployment(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/tenants/tenant-id/brokers/installs/install-id/deployments", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		_, _ = fmt.Fprint(w, `
{
  "jsonapi": { "version": "1.0" },
  "data": {
    "id": "0779acb9-968e-4bff-abd7-94193e589028",
    "type": "broker_deployment",
    "attributes": {
      "install_id": "216b7774-9198-4fcd-a525-bace50228e18",
      "broker_app_installed_in_org_id": "6c6b3b6d-24e5-4f70-9896-4a49609cd61a",
      "metadata": {}
    }
  },
  "links": {}
}
`)
	})
	expectedDeployment := &BrokerDeployment{
		ID:   "0779acb9-968e-4bff-abd7-94193e589028",
		Type: "broker_deployment",
		Attributes: &BrokerDeploymentAttributes{
			AppInstallID: "216b7774-9198-4fcd-a525-bace50228e18",
			OrgID:        "6c6b3b6d-24e5-4f70-9896-4a49609cd61a",
			Metadata:     map[string]string{},
		},
	}

	actualDeployment, _, err := client.Brokers.CreateDeployment(ctx, "tenant-id", "install-id",
		&BrokerDeploymentCreateOrUpdateRequest{OrgID: "6c6b3b6d-24e5-4f70-9896-4a49609cd61a"},
	)

	assert.NoError(t, err)
	assert.Equal(t, expectedDeployment, actualDeployment)
}

func TestBrokers_CreateDeployment_emptyTenantID(t *testing.T) {
	_, _, err := client.Brokers.CreateDeployment(ctx, "", "install-id", &BrokerDeploymentCreateOrUpdateRequest{})

	assert.Error(t, err)
	assert.ErrorContains(t, err, "tenant id must be supplied")
}

func TestBrokers_CreateDeployment_emptyAppInstallID(t *testing.T) {
	_, _, err := client.Brokers.CreateDeployment(ctx, "tenant-id", "", &BrokerDeploymentCreateOrUpdateRequest{})

	assert.Error(t, err)
	assert.ErrorContains(t, err, "install id must be supplied")
}

func TestBrokers_CreateDeployment_emptyPayload(t *testing.T) {
	_, _, err := client.Brokers.CreateDeployment(ctx, "tenant-id", "install-id", nil)

	assert.Error(t, err)
	assert.ErrorContains(t, err, "payload must be supplied")
}

func TestBrokers_UpdateDeployment(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/tenants/tenant-id/brokers/installs/install-id/deployments/deployment-id", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPatch, r.Method)
		_, _ = fmt.Fprint(w, `
{
  "jsonapi": { "version": "1.0" },
  "data": {
    "id": "0779acb9-968e-4bff-abd7-94193e589028",
    "type": "broker_deployment",
    "attributes": {
      "install_id": "216b7774-9198-4fcd-a525-bace50228e18",
      "broker_app_installed_in_org_id": "6c6b3b6d-24e5-4f70-9896-4a49609cd61a",
      "metadata": { "region": "us-east-1" }
    }
  },
  "links": {}
}
`)
	})
	expectedDeployment := &BrokerDeployment{
		ID:   "0779acb9-968e-4bff-abd7-94193e589028",
		Type: "broker_deployment",
		Attributes: &BrokerDeploymentAttributes{
			AppInstallID: "216b7774-9198-4fcd-a525-bace50228e18",
			OrgID:        "6c6b3b6d-24e5-4f70-9896-4a49609cd61a",
			Metadata:     map[string]string{"region": "us-east-1"},
		},
	}

	actualDeployment, _, err := client.Brokers.UpdateDeployment(ctx, "tenant-id", "install-id", "deployment-id",
		&BrokerDeploymentCreateOrUpdateRequest{OrgID: "6c6b3b6d-24e5-4f70-9896-4a49609cd61a", Metadata: map[string]string{"region": "us-east-1"}},
	)

	assert.NoError(t, err)
	assert.Equal(t, expectedDeployment, actualDeployment)
}

func TestBrokers_UpdateDeployment_emptyTenantID(t *testing.T) {
	_, _, err := client.Brokers.UpdateDeployment(ctx, "", "install-id", "deployment-id", nil)

	assert.Error(t, err)
	assert.ErrorContains(t, err, "tenant id must be supplied")
}

func TestBrokers_UpdateDeployment_emptyAppInstallID(t *testing.T) {
	_, _, err := client.Brokers.UpdateDeployment(ctx, "tenant-id", "", "deployment-id", nil)

	assert.Error(t, err)
	assert.ErrorContains(t, err, "install id must be supplied")
}

func TestBrokers_UpdateDeployment_emptyDeploymentID(t *testing.T) {
	_, _, err := client.Brokers.UpdateDeployment(ctx, "tenant-id", "install-id", "", nil)

	assert.Error(t, err)
	assert.ErrorContains(t, err, "id must be supplied")
}

func TestBrokers_UpdateDeployment_emptyPayload(t *testing.T) {
	_, _, err := client.Brokers.UpdateDeployment(ctx, "tenant-id", "install-id", "deployment-id", nil)

	assert.Error(t, err)
	assert.ErrorContains(t, err, "payload must be supplied")
}

func TestBrokers_DeleteDeployment(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/tenants/tenant-id/brokers/installs/install-id/deployments/deployment-id", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
	})

	_, err := client.Brokers.DeleteDeployment(ctx, "tenant-id", "install-id", "deployment-id")

	assert.NoError(t, err)
}

func TestBrokers_DeleteDeployment_emptyTenantID(t *testing.T) {
	_, err := client.Brokers.DeleteDeployment(ctx, "", "install-id", "deployment-id")

	assert.Error(t, err)
	assert.ErrorContains(t, err, "tenant id must be supplied")
}

func TestBrokers_DeleteDeployment_emptyAppInstallID(t *testing.T) {
	_, err := client.Brokers.DeleteDeployment(ctx, "tenant-id", "", "deployment-id")

	assert.Error(t, err)
	assert.ErrorContains(t, err, "install id must be supplied")
}

func TestBrokers_DeleteDeployment_emptyDeploymentID(t *testing.T) {
	_, err := client.Brokers.DeleteDeployment(ctx, "tenant-id", "install-id", "")

	assert.Error(t, err)
	assert.ErrorContains(t, err, "id must be supplied")
}

func TestBrokers_ListDeploymentCredentials(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/tenants/tenant-id/brokers/installs/install-id/deployments/deployment-id/credentials", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		_, _ = fmt.Fprint(w, `
{
  "jsonapi": { "version": "1.0" },
  "data": [
    {
      "id": "7fba7667-2ca3-4534-ab3f-bd1b61b0bd7b",
      "type": "deployment_credential",
      "attributes": {
        "comment": "test comment for gitlab",
        "deployment_id": "1793ad4f-f506-45a7-8c8c-d14f25fff941",
        "environment_variable_name": "MY_GITLAB_TEST_TOKEN",
        "type": "gitlab"
      }
    },
    {
      "id": "354e0e11-d3c8-405d-8fec-33683276a98b",
      "type": "deployment_credential",
      "attributes": {
        "comment": "test comment for github",
        "deployment_id": "1793ad4f-f506-45a7-8c8c-d14f25fff941",
        "environment_variable_name": "MY_GITHUB_TEST_TOKEN",
        "type": "github"
      }
    }
  ],
  "links": {}
}
`)
	})
	expectedDeploymentCredentials := []BrokerDeploymentCredential{
		{
			ID:   "7fba7667-2ca3-4534-ab3f-bd1b61b0bd7b",
			Type: "deployment_credential",
			Attributes: &BrokerDeploymentCredentialAttributes{
				Comment:            "test comment for gitlab",
				BrokerDeploymentID: "1793ad4f-f506-45a7-8c8c-d14f25fff941",
				EnvVarName:         "MY_GITLAB_TEST_TOKEN",
				Type:               "gitlab",
			},
		},
		{
			ID:   "354e0e11-d3c8-405d-8fec-33683276a98b",
			Type: "deployment_credential",
			Attributes: &BrokerDeploymentCredentialAttributes{
				Comment:            "test comment for github",
				BrokerDeploymentID: "1793ad4f-f506-45a7-8c8c-d14f25fff941",
				EnvVarName:         "MY_GITHUB_TEST_TOKEN",
				Type:               "github",
			},
		},
	}

	actualDeploymentCredentials, _, err := client.Brokers.ListDeploymentCredentials(ctx, "tenant-id", "install-id", "deployment-id")

	assert.NoError(t, err)
	assert.Equal(t, 2, len(actualDeploymentCredentials), "expect 2 deployment credentials")
	assert.Equal(t, expectedDeploymentCredentials, actualDeploymentCredentials)
}

func TestBrokers_ListDeploymentCredentials_emptyTenantID(t *testing.T) {
	_, _, err := client.Brokers.ListDeploymentCredentials(ctx, "", "install-id", "deployment-id")

	assert.Error(t, err)
	assert.ErrorContains(t, err, "tenant id must be supplied")
}

func TestBrokers_ListDeploymentCredentials_emptyAppInstallID(t *testing.T) {
	_, _, err := client.Brokers.ListDeploymentCredentials(ctx, "tenant-id", "", "deployment-id")

	assert.Error(t, err)
	assert.ErrorContains(t, err, "app install id must be supplied")
}

func TestBrokers_ListDeploymentCredentials_emptyDeploymentID(t *testing.T) {
	_, _, err := client.Brokers.ListDeploymentCredentials(ctx, "tenant-id", "install-id", "")

	assert.Error(t, err)
	assert.ErrorContains(t, err, "deployment id must be supplied")
}

func TestBrokers_GetDeploymentCredential(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/tenants/tenant-id/brokers/installs/install-id/deployments/deployment-id/credentials/credential-id", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		_, _ = fmt.Fprint(w, `
{
  "jsonapi": { "version": "1.0" },
  "data": {
    "id": "7fba7667-2ca3-4534-ab3f-bd1b61b0bd7b",
    "type": "deployment_credential",
    "attributes": {
      "comment": "test comment for gitlab",
      "deployment_id": "1793ad4f-f506-45a7-8c8c-d14f25fff941",
      "environment_variable_name": "MY_GITLAB_TEST_TOKEN",
      "type": "gitlab"
    },
    "relationships": {
      "broker_connections": []
    }
  },
  "links": {}
}
`)
	})
	expectedDeploymentCredential := &BrokerDeploymentCredential{
		ID:   "7fba7667-2ca3-4534-ab3f-bd1b61b0bd7b",
		Type: "deployment_credential",
		Attributes: &BrokerDeploymentCredentialAttributes{
			Comment:            "test comment for gitlab",
			BrokerDeploymentID: "1793ad4f-f506-45a7-8c8c-d14f25fff941",
			EnvVarName:         "MY_GITLAB_TEST_TOKEN",
			Type:               "gitlab",
		},
	}

	actualDeploymentCredential, _, err := client.Brokers.GetDeploymentCredential(ctx, "tenant-id", "install-id", "deployment-id", "credential-id")

	assert.NoError(t, err)
	assert.Equal(t, expectedDeploymentCredential, actualDeploymentCredential)
}

func TestBrokers_GetDeploymentCredential_emptyTenantID(t *testing.T) {
	_, _, err := client.Brokers.GetDeploymentCredential(ctx, "", "install-id", "deployment-id", "credential-id")

	assert.Error(t, err)
	assert.ErrorContains(t, err, "tenant id must be supplied")
}

func TestBrokers_GetDeploymentCredential_emptyAppInstallID(t *testing.T) {
	_, _, err := client.Brokers.GetDeploymentCredential(ctx, "tenant-id", "", "deployment-id", "credential-id")

	assert.Error(t, err)
	assert.ErrorContains(t, err, "app install id must be supplied")
}

func TestBrokers_GetDeploymentCredential_emptyDeploymentID(t *testing.T) {
	_, _, err := client.Brokers.GetDeploymentCredential(ctx, "tenant-id", "install-id", "", "credential-id")

	assert.Error(t, err)
	assert.ErrorContains(t, err, "deployment id must be supplied")
}

func TestBrokers_GetDeploymentCredential_emptyCredentialID(t *testing.T) {
	_, _, err := client.Brokers.GetDeploymentCredential(ctx, "tenant-id", "install-id", "deployment-id", "")

	assert.Error(t, err)
	assert.ErrorContains(t, err, "credential id must be supplied")
}

func TestBrokers_CreateDeploymentCredential(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/tenants/tenant-id/brokers/installs/install-id/deployments/deployment-id/credentials", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		_, _ = fmt.Fprint(w, `
{
  "jsonapi": { "version": "1.0" },
  "data": [
    {
      "id": "7fba7667-2ca3-4534-ab3f-bd1b61b0bd7b",
      "type": "deployment_credential",
      "attributes": {
        "comment": "test comment for gitlab",
        "deployment_id": "1793ad4f-f506-45a7-8c8c-d14f25fff941",
        "environment_variable_name": "MY_GITLAB_TEST_TOKEN",
        "type": "gitlab"
      }
    }
  ],
  "links": {}
}
`)
	})
	expectedDeploymentCredential := &BrokerDeploymentCredential{
		ID:   "7fba7667-2ca3-4534-ab3f-bd1b61b0bd7b",
		Type: "deployment_credential",
		Attributes: &BrokerDeploymentCredentialAttributes{
			Comment:            "test comment for gitlab",
			BrokerDeploymentID: "1793ad4f-f506-45a7-8c8c-d14f25fff941",
			EnvVarName:         "MY_GITLAB_TEST_TOKEN",
			Type:               "gitlab",
		},
	}

	actualDeploymentCredential, _, err := client.Brokers.CreateDeploymentCredential(ctx, "tenant-id", "install-id", "deployment-id",
		&BrokerDeploymentCredentialCreateOrUpdateRequest{
			Comment:    "test comment for gitlab",
			EnvVarName: "MY_GITLAB_TEST_TOKEN",
			Type:       "gitlab",
		},
	)

	assert.NoError(t, err)
	assert.Equal(t, expectedDeploymentCredential, actualDeploymentCredential)
}

func TestBrokers_CreateDeploymentCredential_emptyTenantID(t *testing.T) {
	_, _, err := client.Brokers.CreateDeploymentCredential(ctx, "", "install-id", "deployment-id", nil)

	assert.Error(t, err)
	assert.ErrorContains(t, err, "tenant id must be supplied")
}

func TestBrokers_CreateDeploymentCredential_emptyAppInstallID(t *testing.T) {
	_, _, err := client.Brokers.CreateDeploymentCredential(ctx, "tenant-id", "", "deployment-id", nil)

	assert.Error(t, err)
	assert.ErrorContains(t, err, "app install id must be supplied")
}

func TestBrokers_CreateDeploymentCredential_emptyDeploymentID(t *testing.T) {
	_, _, err := client.Brokers.CreateDeploymentCredential(ctx, "tenant-id", "install-id", "", nil)

	assert.Error(t, err)
	assert.ErrorContains(t, err, "deployment id must be supplied")
}

func TestBrokers_CreateDeploymentCredential_emptyPayload(t *testing.T) {
	_, _, err := client.Brokers.CreateDeploymentCredential(ctx, "tenant-id", "install-id", "deployment-id", nil)

	assert.Error(t, err)
	assert.ErrorContains(t, err, "payload must be supplied")
}

func TestBrokers_UpdateDeploymentCredential(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/tenants/tenant-id/brokers/installs/install-id/deployments/deployment-id/credentials/credential-id", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPatch, r.Method)
		_, _ = fmt.Fprint(w, `
{
  "jsonapi": { "version": "1.0" },
  "data": {
    "id": "7fba7667-2ca3-4534-ab3f-bd1b61b0bd7b",
    "type": "deployment_credential",
    "attributes": {
      "comment": "test comment for gitlab (updated)",
      "deployment_id": "1793ad4f-f506-45a7-8c8c-d14f25fff941",
      "environment_variable_name": "MY_GITLAB_TEST_TOKEN_UPDATED",
      "type": "gitlab"
    }
  },
  "links": {}
}
`)
	})
	expectedDeploymentCredential := &BrokerDeploymentCredential{
		ID:   "7fba7667-2ca3-4534-ab3f-bd1b61b0bd7b",
		Type: "deployment_credential",
		Attributes: &BrokerDeploymentCredentialAttributes{
			Comment:            "test comment for gitlab (updated)",
			BrokerDeploymentID: "1793ad4f-f506-45a7-8c8c-d14f25fff941",
			EnvVarName:         "MY_GITLAB_TEST_TOKEN_UPDATED",
			Type:               "gitlab",
		},
	}

	actualDeploymentCredential, _, err := client.Brokers.UpdateDeploymentCredential(ctx, "tenant-id", "install-id", "deployment-id", "credential-id",
		&BrokerDeploymentCredentialCreateOrUpdateRequest{
			Comment:    "test comment for gitlab (updated)",
			EnvVarName: "MY_GITLAB_TEST_TOKEN_UPDATED",
			Type:       "gitlab",
		},
	)

	assert.NoError(t, err)
	assert.Equal(t, expectedDeploymentCredential, actualDeploymentCredential)
}

func TestBrokers_UpdateDeploymentCredential_emptyTenantID(t *testing.T) {
	_, _, err := client.Brokers.UpdateDeploymentCredential(ctx, "", "install-id", "deployment-id", "credential-id", nil)

	assert.Error(t, err)
	assert.ErrorContains(t, err, "tenant id must be supplied")
}

func TestBrokers_UpdateDeploymentCredential_emptyAppInstallID(t *testing.T) {
	_, _, err := client.Brokers.UpdateDeploymentCredential(ctx, "tenant-id", "", "deployment-id", "credential-id", nil)

	assert.Error(t, err)
	assert.ErrorContains(t, err, "app install id must be supplied")
}

func TestBrokers_UpdateDeploymentCredential_emptyDeploymentID(t *testing.T) {
	_, _, err := client.Brokers.UpdateDeploymentCredential(ctx, "tenant-id", "install-id", "", "credential-id", nil)

	assert.Error(t, err)
	assert.ErrorContains(t, err, "deployment id must be supplied")
}

func TestBrokers_UpdateDeploymentCredential_emptyCredentialID(t *testing.T) {
	_, _, err := client.Brokers.UpdateDeploymentCredential(ctx, "tenant-id", "install-id", "deployment-id", "", nil)

	assert.Error(t, err)
	assert.ErrorContains(t, err, "credential id must be supplied")
}

func TestBrokers_UpdateDeploymentCredential_emptyPayload(t *testing.T) {
	_, _, err := client.Brokers.UpdateDeploymentCredential(ctx, "tenant-id", "install-id", "deployment-id", "credential-id", nil)

	assert.Error(t, err)
	assert.ErrorContains(t, err, "payload must be supplied")
}

func TestBrokers_DeleteDeploymentCredential(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/tenants/tenant-id/brokers/installs/install-id/deployments/deployment-id/credentials/credential-id", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
	})

	_, err := client.Brokers.DeleteDeploymentCredential(ctx, "tenant-id", "install-id", "deployment-id", "credential-id")

	assert.NoError(t, err)
}

func TestBrokers_DeleteDeploymentCredential_emptyTenantID(t *testing.T) {
	_, err := client.Brokers.DeleteDeploymentCredential(ctx, "", "install-id", "deployment-id", "credential-id")

	assert.Error(t, err)
	assert.ErrorContains(t, err, "tenant id must be supplied")
}

func TestBrokers_DeleteDeploymentCredential_emptyAppInstallID(t *testing.T) {
	_, err := client.Brokers.DeleteDeploymentCredential(ctx, "tenant-id", "", "deployment-id", "credential-id")

	assert.Error(t, err)
	assert.ErrorContains(t, err, "app install id must be supplied")
}

func TestBrokers_DeleteDeploymentCredential_emptyDeploymentID(t *testing.T) {
	_, err := client.Brokers.DeleteDeploymentCredential(ctx, "tenant-id", "install-id", "", "credential-id")

	assert.Error(t, err)
	assert.ErrorContains(t, err, "deployment id must be supplied")
}

func TestBrokers_DeleteDeploymentCredential_emptyCredentialID(t *testing.T) {
	_, err := client.Brokers.DeleteDeploymentCredential(ctx, "tenant-id", "install-id", "deployment-id", "")

	assert.Error(t, err)
	assert.ErrorContains(t, err, "credential id must be supplied")
}

func TestBrokers_ListConnections(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/tenants/tenant-id/brokers/installs/install-id/deployments/deployment-id/connections", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		_, _ = fmt.Fprint(w, `
{
  "jsonapi": { "version": "1.0" },
  "data": [
    {
      "id": "a9d79dc9-63c5-4b5d-ae5c-5c42bc2f3d38",
      "type": "broker_connection",
      "attributes": {
        "deployment_id": "1793ad4f-f506-45a7-8c8c-d14f25fff941",
        "identifier": "9dd6c62e-6541-4ff5-8e2c-e69e5183f2cc",
        "name": "test-github-connection",
        "configuration": {
          "required": {
            "github_token": "${MY_GITHUB_TEST_TOKEN}"
          },
          "type": "github"
        }
      }
    }
  ],
  "links": {}
}
`)
	})
	expectedConnections := []BrokerConnection{
		{
			ID:   "a9d79dc9-63c5-4b5d-ae5c-5c42bc2f3d38",
			Type: "broker_connection",
			Attributes: &BrokerConnectionAttributes{
				BrokerDeploymentID: "1793ad4f-f506-45a7-8c8c-d14f25fff941",
				Configuration: &BrokerConnectionAttributesConfiguration{
					GitHub: &BrokerConnectionGitHubConfiguration{GitHubToken: "${MY_GITHUB_TEST_TOKEN}"},
					Type:   "github",
				},
				Identifier: "9dd6c62e-6541-4ff5-8e2c-e69e5183f2cc",
				Name:       "test-github-connection",
			},
		},
	}

	actualConnections, _, err := client.Brokers.ListConnections(ctx, "tenant-id", "install-id", "deployment-id")

	assert.NoError(t, err)
	assert.Equal(t, 1, len(actualConnections), "expect 1 broker connection")
	assert.Equal(t, expectedConnections, actualConnections)
}

func TestBrokers_ListConnections_emptyTenantID(t *testing.T) {
	_, _, err := client.Brokers.ListConnections(ctx, "", "install-id", "deployment-id")

	assert.Error(t, err)
	assert.ErrorContains(t, err, "tenant id must be supplied")
}

func TestBrokers_ListConnections_emptyAppInstallID(t *testing.T) {
	_, _, err := client.Brokers.ListConnections(ctx, "tenant-id", "", "deployment-id")

	assert.Error(t, err)
	assert.ErrorContains(t, err, "app install id must be supplied")
}

func TestBrokers_ListConnections_emptyDeploymentID(t *testing.T) {
	_, _, err := client.Brokers.ListConnections(ctx, "tenant-id", "install-id", "")

	assert.Error(t, err)
	assert.ErrorContains(t, err, "deployment id must be supplied")
}

func TestBrokers_GetConnection(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/tenants/tenant-id/brokers/installs/install-id/deployments/deployment-id/connections/connection-id", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		_, _ = fmt.Fprint(w, `
{
  "jsonapi": { "version": "1.0" },
  "data": {
    "id": "a9d79dc9-63c5-4b5d-ae5c-5c42bc2f3d38",
    "type": "broker_connection",
    "attributes": {
      "deployment_id": "1793ad4f-f506-45a7-8c8c-d14f25fff941",
      "identifier": "9dd6c62e-6541-4ff5-8e2c-e69e5183f2cc",
      "name": "test-github-connection",
      "configuration": {
        "required": { "github_token": "${MY_GITHUB_TEST_TOKEN}" },
        "type": "github"
      }
    }
  },
  "links": {}
}
`)
	})
	expectedConnection := &BrokerConnection{
		ID:   "a9d79dc9-63c5-4b5d-ae5c-5c42bc2f3d38",
		Type: "broker_connection",
		Attributes: &BrokerConnectionAttributes{
			BrokerDeploymentID: "1793ad4f-f506-45a7-8c8c-d14f25fff941",
			Configuration: &BrokerConnectionAttributesConfiguration{
				GitHub: &BrokerConnectionGitHubConfiguration{GitHubToken: "${MY_GITHUB_TEST_TOKEN}"},
				Type:   "github",
			},
			Identifier: "9dd6c62e-6541-4ff5-8e2c-e69e5183f2cc",
			Name:       "test-github-connection",
		},
	}

	actualConnection, _, err := client.Brokers.GetConnection(ctx, "tenant-id", "install-id", "deployment-id", "connection-id")

	assert.NoError(t, err)
	assert.Equal(t, expectedConnection, actualConnection)
}

func TestBrokers_GetConnection_emptyTenantID(t *testing.T) {
	_, _, err := client.Brokers.GetConnection(ctx, "", "install-id", "deployment-id", "connection-id")

	assert.Error(t, err)
	assert.ErrorContains(t, err, "tenant id must be supplied")
}

func TestBrokers_GetConnection_emptyAppInstallID(t *testing.T) {
	_, _, err := client.Brokers.GetConnection(ctx, "tenant-id", "", "deployment-id", "connection-id")

	assert.Error(t, err)
	assert.ErrorContains(t, err, "app install id must be supplied")
}

func TestBrokers_GetConnection_emptyDeploymentID(t *testing.T) {
	_, _, err := client.Brokers.GetConnection(ctx, "tenant-id", "install-id", "", "connection-id")

	assert.Error(t, err)
	assert.ErrorContains(t, err, "deployment id must be supplied")
}

func TestBrokers_GetConnection_emptyConnectionID(t *testing.T) {
	_, _, err := client.Brokers.GetConnection(ctx, "tenant-id", "install-id", "deployment-id", "")

	assert.Error(t, err)
	assert.ErrorContains(t, err, "connection id must be supplied")
}

func TestBrokers_CreateConnection(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/tenants/tenant-id/brokers/installs/install-id/deployments/deployment-id/connections", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		_, _ = fmt.Fprint(w, `
{
  "jsonapi": { "version": "1.0" },
  "data": {
    "id": "a9d79dc9-63c5-4b5d-ae5c-5c42bc2f3d38",
    "type": "broker_connection",
    "attributes": {
      "deployment_id": "1793ad4f-f506-45a7-8c8c-d14f25fff941",
      "identifier": "9dd6c62e-6541-4ff5-8e2c-e69e5183f2cc",
      "name": "test-github-connection",
      "configuration": {
        "required": {
          "broker_client_url": "http://locahost:8080",
          "github_token": "${MY_GITHUB_TEST_TOKEN}"
        },
        "type": "github"
      }
    }
  },
  "links": {}
}
`)
	})
	expectedConnection := &BrokerConnection{
		ID:   "a9d79dc9-63c5-4b5d-ae5c-5c42bc2f3d38",
		Type: "broker_connection",
		Attributes: &BrokerConnectionAttributes{
			BrokerDeploymentID: "1793ad4f-f506-45a7-8c8c-d14f25fff941",
			Configuration: &BrokerConnectionAttributesConfiguration{
				GitHub: &BrokerConnectionGitHubConfiguration{
					BrokerClientURL: "http://locahost:8080",
					GitHubToken:     "${MY_GITHUB_TEST_TOKEN}"},
				Type: "github",
			},
			Identifier: "9dd6c62e-6541-4ff5-8e2c-e69e5183f2cc",
			Name:       "test-github-connection",
		},
	}

	actualConnection, _, err := client.Brokers.CreateConnection(
		ctx, "tenant-id", "install-id", "deployment-id",
		&BrokerConnectionCreateOrUpdateRequest{
			BrokerClientURL: "http://locahost:8080",
			GitHubToken:     "1a519722-816d-4fdf-b501-2528e91bcda4",
			Type:            BrokerConnectionTypeGitHub,
		},
	)

	assert.NoError(t, err)
	assert.Equal(t, expectedConnection, actualConnection)
}

func TestBrokers_CreateConnection_emptyTenantID(t *testing.T) {
	_, _, err := client.Brokers.CreateConnection(ctx, "", "install-id", "deployment-id", nil)

	assert.Error(t, err)
	assert.ErrorContains(t, err, "tenant id must be supplied")
}

func TestBrokers_CreateConnection_emptyAppInstallID(t *testing.T) {
	_, _, err := client.Brokers.CreateConnection(ctx, "tenant-id", "", "deployment-id", nil)

	assert.Error(t, err)
	assert.ErrorContains(t, err, "app install id must be supplied")
}

func TestBrokers_CreateConnection_emptyDeploymentID(t *testing.T) {
	_, _, err := client.Brokers.CreateConnection(ctx, "tenant-id", "install-id", "", nil)

	assert.Error(t, err)
	assert.ErrorContains(t, err, "deployment id must be supplied")
}

func TestBrokers_CreateConnection_emptyPayload(t *testing.T) {
	_, _, err := client.Brokers.CreateConnection(ctx, "tenant-id", "install-id", "deployment-id", nil)

	assert.Error(t, err)
	assert.ErrorContains(t, err, "payload must be supplied")
}

func TestBrokers_UpdateConnection(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/tenants/tenant-id/brokers/installs/install-id/deployments/deployment-id/connections/connection-id", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPatch, r.Method)
		_, _ = fmt.Fprint(w, `
{
  "jsonapi": { "version": "1.0" },
  "data": {
    "id": "a9d79dc9-63c5-4b5d-ae5c-5c42bc2f3d38",
    "type": "broker_connection",
    "attributes": {
      "deployment_id": "1793ad4f-f506-45a7-8c8c-d14f25fff941",
      "identifier": "9dd6c62e-6541-4ff5-8e2c-e69e5183f2cc",
      "name": "test-github-connection-updated",
      "configuration": {
        "required": {
          "broker_client_url": "http://locahost:8080",
          "github_token": "${MY_GITHUB_TEST_TOKEN}"
        },
        "type": "github"
      }
    }
  },
  "links": {}
}
`)
	})
	expectedConnection := &BrokerConnection{
		ID:   "a9d79dc9-63c5-4b5d-ae5c-5c42bc2f3d38",
		Type: "broker_connection",
		Attributes: &BrokerConnectionAttributes{
			BrokerDeploymentID: "1793ad4f-f506-45a7-8c8c-d14f25fff941",
			Configuration: &BrokerConnectionAttributesConfiguration{
				GitHub: &BrokerConnectionGitHubConfiguration{
					BrokerClientURL: "http://locahost:8080",
					GitHubToken:     "${MY_GITHUB_TEST_TOKEN}"},
				Type: "github",
			},
			Identifier: "9dd6c62e-6541-4ff5-8e2c-e69e5183f2cc",
			Name:       "test-github-connection-updated",
		},
	}

	actualConnection, _, err := client.Brokers.UpdateConnection(
		ctx, "tenant-id", "install-id", "deployment-id", "connection-id",
		&BrokerConnectionCreateOrUpdateRequest{
			BrokerClientURL: "http://locahost:8080",
			GitHubToken:     "1a519722-816d-4fdf-b501-2528e91bcda4",
			Type:            BrokerConnectionTypeGitHub,
		},
	)

	assert.NoError(t, err)
	assert.Equal(t, expectedConnection, actualConnection)
}

func TestBrokers_UpdateConnection_emptyTenantID(t *testing.T) {
	_, _, err := client.Brokers.UpdateConnection(ctx, "", "install-id", "deployment-id", "connection-id", nil)

	assert.Error(t, err)
	assert.ErrorContains(t, err, "tenant id must be supplied")
}

func TestBrokers_UpdateConnection_emptyAppInstallID(t *testing.T) {
	_, _, err := client.Brokers.UpdateConnection(ctx, "tenant-id", "", "deployment-id", "connection-id", nil)

	assert.Error(t, err)
	assert.ErrorContains(t, err, "app install id must be supplied")
}

func TestBrokers_UpdateConnection_emptyDeploymentID(t *testing.T) {
	_, _, err := client.Brokers.UpdateConnection(ctx, "tenant-id", "install-id", "", "connection-id", nil)

	assert.Error(t, err)
	assert.ErrorContains(t, err, "deployment id must be supplied")
}

func TestBrokers_UpdateConnection_emptyConnectionID(t *testing.T) {
	_, _, err := client.Brokers.UpdateConnection(ctx, "tenant-id", "install-id", "deployment-id", "", nil)

	assert.Error(t, err)
	assert.ErrorContains(t, err, "connection id must be supplied")
}

func TestBrokers_UpdateConnection_emptyPayload(t *testing.T) {
	_, _, err := client.Brokers.UpdateConnection(ctx, "tenant-id", "install-id", "deployment-id", "connection-id", nil)

	assert.Error(t, err)
	assert.ErrorContains(t, err, "payload must be supplied")
}

func TestBrokers_DeleteConnection(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/tenants/tenant-id/brokers/installs/install-id/deployments/deployment-id/connections/connection-id", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
	})

	_, err := client.Brokers.DeleteConnection(ctx, "tenant-id", "install-id", "deployment-id", "connection-id")

	assert.NoError(t, err)
}

func TestBrokers_DeleteConnection_emptyTenantID(t *testing.T) {
	_, err := client.Brokers.DeleteConnection(ctx, "", "install-id", "deployment-id", "connection-id")

	assert.Error(t, err)
	assert.ErrorContains(t, err, "tenant id must be supplied")
}

func TestBrokers_DeleteConnection_emptyAppInstallID(t *testing.T) {
	_, err := client.Brokers.DeleteConnection(ctx, "tenant-id", "", "deployment-id", "connection-id")

	assert.Error(t, err)
	assert.ErrorContains(t, err, "app install id must be supplied")
}

func TestBrokers_DeleteConnection_emptyDeploymentID(t *testing.T) {
	_, err := client.Brokers.DeleteConnection(ctx, "tenant-id", "install-id", "", "connection-id")

	assert.Error(t, err)
	assert.ErrorContains(t, err, "deployment id must be supplied")
}

func TestBrokers_DeleteConnection_emptyConnectionID(t *testing.T) {
	_, err := client.Brokers.DeleteConnection(ctx, "tenant-id", "install-id", "deployment-id", "")

	assert.Error(t, err)
	assert.ErrorContains(t, err, "connection id must be supplied")
}

func TestBrokers_buildBrokerConnectionRequestPayload(t *testing.T) {
	_, err := buildBrokerConnectionRequestPayload("", nil)

	assert.Error(t, err)
}

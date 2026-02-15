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

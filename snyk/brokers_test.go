package snyk

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var _ func(*BrokersService, context.Context, string) ([]BrokerDeployment, *Response, error) = (*BrokersService).ListDeploymentsForTenant

func TestBrokers_ListDeployments(t *testing.T) {
	setup(t)
	defer teardown()

	mux.HandleFunc("/tenants/tenant-id/brokers/installs/install-id/deployments", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assertRequestAPIVersion(t, r, brokersAPIVersion)
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

	require.ErrorIs(t, err, ErrEmptyArgument)
	assert.EqualError(t, err, "tenant ID: argument is empty")
}

func TestBrokers_ListDeployments_emptyAppInstallID(t *testing.T) {
	_, _, err := client.Brokers.ListDeployments(ctx, "tenant-id", "")

	assert.Error(t, err)
	assert.ErrorContains(t, err, "install id must be supplied")
}

func TestBrokers_ListDeploymentsForTenant(t *testing.T) {
	setup(t)
	defer teardown()

	mux.HandleFunc("/tenants/tenant-id/brokers/deployments", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assertRequestAPIVersion(t, r, brokersAPIVersion)
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

	actualDeployments, _, err := client.Brokers.ListDeploymentsForTenant(ctx, "tenant-id")

	assert.NoError(t, err)
	assert.Equal(t, expectedDeployments, actualDeployments)
}

func TestBrokers_ListDeploymentsForTenant_emptyTenantID(t *testing.T) {
	_, _, err := client.Brokers.ListDeploymentsForTenant(ctx, "")

	require.ErrorIs(t, err, ErrEmptyArgument)
	assert.EqualError(t, err, "tenant ID: argument is empty")
}

func TestBrokers_AllDeploymentsForTenant_multiplePages(t *testing.T) {
	setup(t)
	defer teardown()

	options := &ListOptions{StartingAfter: "initial-cursor", Limit: 2}
	requestCount := 0
	mux.HandleFunc("/tenants/tenant-id/brokers/deployments", func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		assert.Equal(t, http.MethodGet, r.Method)
		assertRequestAPIVersion(t, r, brokersAPIVersion)

		switch requestCount {
		case 1:
			assert.Equal(t, url.Values{
				"limit":          {"2"},
				"starting_after": {"initial-cursor"},
				"version":        {brokersAPIVersion},
			}, r.URL.Query())
			w.Header().Set(headerSnykRequestID, "page-1-request")
			_, _ = fmt.Fprint(w, `{
				"data": [
					{"id":"11111111-1111-1111-1111-111111111111","type":"broker_deployment"},
					{"id":"22222222-2222-2222-2222-222222222222","type":"broker_deployment"}
				],
				"links": {"next":"/rest/tenants/tenant-id/brokers/deployments?limit=2&starting_after=next-cursor"}
			}`)
		case 2:
			assert.Equal(t, url.Values{
				"limit":          {"2"},
				"starting_after": {"next-cursor"},
				"version":        {brokersAPIVersion},
			}, r.URL.Query())
			w.Header().Set(headerSnykRequestID, "page-2-request")
			_, _ = fmt.Fprint(w, `{
				"data": [
					{"id":"33333333-3333-3333-3333-333333333333","type":"broker_deployment"}
				],
				"links": {}
			}`)
		default:
			http.Error(w, "unexpected pagination request", http.StatusInternalServerError)
		}
	})

	seq, iterErr := client.Brokers.AllDeploymentsForTenant(ctx, "tenant-id", options)
	var deployments []BrokerDeployment
	var responses []*Response
	for deployment, response := range seq {
		deployments = append(deployments, deployment)
		responses = append(responses, response)
	}

	require.NoError(t, iterErr())
	require.Len(t, deployments, 3)
	assert.Equal(t, []string{
		"11111111-1111-1111-1111-111111111111",
		"22222222-2222-2222-2222-222222222222",
		"33333333-3333-3333-3333-333333333333",
	}, []string{deployments[0].ID, deployments[1].ID, deployments[2].ID})
	require.Len(t, responses, 3)
	assert.Same(t, responses[0], responses[1])
	assert.NotSame(t, responses[0], responses[2])
	assert.Equal(t, []string{"page-1-request", "page-1-request", "page-2-request"}, []string{
		responses[0].SnykRequestID,
		responses[1].SnykRequestID,
		responses[2].SnykRequestID,
	})
	assert.Equal(t, 2, requestCount)
	assert.Equal(t, &ListOptions{StartingAfter: "initial-cursor", Limit: 2}, options)
}

func TestBrokers_AllDeploymentsForTenant_preservesEarlierPagesOnLaterFailure(t *testing.T) {
	setup(t)
	defer teardown()

	requestCount := 0
	mux.HandleFunc("/tenants/tenant-id/brokers/deployments", func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		if requestCount == 1 {
			_, _ = fmt.Fprint(w, `{
				"data": [{"id":"11111111-1111-1111-1111-111111111111","type":"broker_deployment"}],
				"links": {"next":"/rest/tenants/tenant-id/brokers/deployments?starting_after=next-cursor"}
			}`)
			return
		}

		assert.Equal(t, "next-cursor", r.URL.Query().Get("starting_after"))
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = fmt.Fprint(w, `{"errors":[{"status":"500","title":"synthetic later-page failure"}]}`)
	})

	seq, iterErr := client.Brokers.AllDeploymentsForTenant(ctx, "tenant-id", nil)
	var deploymentIDs []string
	for deployment := range seq {
		deploymentIDs = append(deploymentIDs, deployment.ID)
	}

	assert.Equal(t, []string{"11111111-1111-1111-1111-111111111111"}, deploymentIDs)
	require.Equal(t, 2, requestCount)
	var responseErr *ErrorResponse
	require.ErrorAs(t, iterErr(), &responseErr)
	assert.Equal(t, http.StatusInternalServerError, responseErr.Response.StatusCode)
	assert.ErrorContains(t, iterErr(), "synthetic later-page failure")
}

func TestBrokers_AllDeploymentsForTenant_snapshotsOptionsAtConstruction(t *testing.T) {
	setup(t)
	defer teardown()

	options := &ListOptions{StartingAfter: "original-cursor", Limit: 20}
	seq, iterErr := client.Brokers.AllDeploymentsForTenant(ctx, "tenant-id", options)

	options.StartingAfter = "changed-cursor"
	options.EndingBefore = "changed-ending-cursor"
	options.Limit = 99

	mux.HandleFunc("/tenants/tenant-id/brokers/deployments", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, url.Values{
			"limit":          {"20"},
			"starting_after": {"original-cursor"},
			"version":        {brokersAPIVersion},
		}, r.URL.Query())
		_, _ = fmt.Fprint(w, `{"data":[],"links":{}}`)
	})

	for range seq {
		t.Fatal("unexpected broker deployment")
	}

	require.NoError(t, iterErr())
}

func TestBrokers_AllDeploymentsForTenant_restartsSequentialIterationFromInitialCursor(t *testing.T) {
	setup(t)
	defer teardown()

	requestCount := 0
	mux.HandleFunc("/tenants/tenant-id/brokers/deployments", func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		switch requestCount {
		case 1, 3:
			assert.Equal(t, "initial-cursor", r.URL.Query().Get("starting_after"))
			_, _ = fmt.Fprint(w, `{
				"data": [{"id":"11111111-1111-1111-1111-111111111111","type":"broker_deployment"}],
				"links": {"next":"/rest/tenants/tenant-id/brokers/deployments?starting_after=next-cursor"}
			}`)
		case 2, 4:
			assert.Equal(t, "next-cursor", r.URL.Query().Get("starting_after"))
			_, _ = fmt.Fprint(w, `{
				"data": [{"id":"22222222-2222-2222-2222-222222222222","type":"broker_deployment"}],
				"links": {}
			}`)
		default:
			http.Error(w, "unexpected pagination request", http.StatusInternalServerError)
		}
	})

	seq, iterErr := client.Brokers.AllDeploymentsForTenant(ctx, "tenant-id", &ListOptions{StartingAfter: "initial-cursor"})

	var firstIteration []string
	for deployment := range seq {
		firstIteration = append(firstIteration, deployment.ID)
	}
	require.NoError(t, iterErr())

	var secondIteration []string
	for deployment := range seq {
		secondIteration = append(secondIteration, deployment.ID)
	}
	require.NoError(t, iterErr())

	wantDeploymentIDs := []string{
		"11111111-1111-1111-1111-111111111111",
		"22222222-2222-2222-2222-222222222222",
	}
	assert.Equal(t, wantDeploymentIDs, firstIteration)
	assert.Equal(t, wantDeploymentIDs, secondIteration)
	assert.Equal(t, 4, requestCount)
}

func TestBrokers_AllDeploymentsForTenant_rejectsEndingBefore(t *testing.T) {
	setup(t)
	defer teardown()

	seq, iterErr := client.Brokers.AllDeploymentsForTenant(ctx, "tenant-id", &ListOptions{
		EndingBefore: "previous-cursor",
	})
	for range seq {
		t.Fatal("unexpected broker deployment")
	}

	assert.EqualError(t, iterErr(), "ending-before pagination is not supported when iterating all broker deployments")
}

func TestBrokers_AllDeploymentsForTenant_emptyTenantID(t *testing.T) {
	c := newTestClient(t)
	seq, iterErr := c.Brokers.AllDeploymentsForTenant(ctx, "", nil)
	for range seq {
		t.Fatal("unexpected broker deployment")
	}

	require.ErrorIs(t, iterErr(), ErrEmptyArgument)
	assert.EqualError(t, iterErr(), "tenant ID: argument is empty")
}

func TestBrokers_CreateDeployment(t *testing.T) {
	setup(t)
	defer teardown()

	mux.HandleFunc("/tenants/tenant-id/brokers/installs/install-id/deployments", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assertRequestAPIVersion(t, r, brokersAPIVersion)
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
	setup(t)
	defer teardown()

	mux.HandleFunc("/tenants/tenant-id/brokers/installs/install-id/deployments/deployment-id", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPatch, r.Method)
		assertRequestAPIVersion(t, r, brokersAPIVersion)
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
	setup(t)
	defer teardown()

	mux.HandleFunc("/tenants/tenant-id/brokers/installs/install-id/deployments/deployment-id", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		assertRequestAPIVersion(t, r, brokersAPIVersion)
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
	setup(t)
	defer teardown()

	mux.HandleFunc("/tenants/tenant-id/brokers/installs/install-id/deployments/deployment-id/credentials", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assertRequestAPIVersion(t, r, brokersAPIVersion)
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
	setup(t)
	defer teardown()

	mux.HandleFunc("/tenants/tenant-id/brokers/installs/install-id/deployments/deployment-id/credentials/credential-id", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assertRequestAPIVersion(t, r, brokersAPIVersion)
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
	setup(t)
	defer teardown()

	mux.HandleFunc("/tenants/tenant-id/brokers/installs/install-id/deployments/deployment-id/credentials", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assertRequestAPIVersion(t, r, brokersAPIVersion)
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
	setup(t)
	defer teardown()

	mux.HandleFunc("/tenants/tenant-id/brokers/installs/install-id/deployments/deployment-id/credentials/credential-id", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPatch, r.Method)
		assertRequestAPIVersion(t, r, brokersAPIVersion)
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
	setup(t)
	defer teardown()

	mux.HandleFunc("/tenants/tenant-id/brokers/installs/install-id/deployments/deployment-id/credentials/credential-id", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		assertRequestAPIVersion(t, r, brokersAPIVersion)
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
	setup(t)
	defer teardown()

	mux.HandleFunc("/tenants/tenant-id/brokers/installs/install-id/deployments/deployment-id/connections", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assertRequestAPIVersion(t, r, brokersAPIVersion)
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
	setup(t)
	defer teardown()

	mux.HandleFunc("/tenants/tenant-id/brokers/installs/install-id/deployments/deployment-id/connections/connection-id", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assertRequestAPIVersion(t, r, brokersAPIVersion)
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
	setup(t)
	defer teardown()

	mux.HandleFunc("/tenants/tenant-id/brokers/installs/install-id/deployments/deployment-id/connections", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assertRequestAPIVersion(t, r, brokersAPIVersion)
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
	setup(t)
	defer teardown()

	mux.HandleFunc("/tenants/tenant-id/brokers/installs/install-id/deployments/deployment-id/connections/connection-id", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPatch, r.Method)
		assertRequestAPIVersion(t, r, brokersAPIVersion)
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
	setup(t)
	defer teardown()

	mux.HandleFunc("/tenants/tenant-id/brokers/installs/install-id/deployments/deployment-id/connections/connection-id", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		assertRequestAPIVersion(t, r, brokersAPIVersion)
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

func TestBrokers_buildBrokerConnectionRequestPayload_emptyPayload(t *testing.T) {
	_, err := buildBrokerConnectionRequestPayload("", nil)

	assert.Error(t, err)
}

func TestBrokers_ListIntegrations(t *testing.T) {
	setup(t)
	defer teardown()

	mux.HandleFunc("/tenants/tenant-id/brokers/connections/connection-id/integrations", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assertRequestAPIVersion(t, r, brokersAPIVersion)
		_, _ = fmt.Fprint(w, `
{
  "jsonapi": { "version": "1.0" },
  "data": [
    {
      "org_id": "6c6b3b6d-24e5-4f70-9896-4a49609cd61a",
      "id": "4ef2ac82-ebcf-4b6c-a39b-6fd27fe09506",
      "integration_type": "github",
      "type": "broker_integration"
    }
  ],
  "links": {}
}
`)
	})
	expectedIntegrations := []BrokerIntegration{
		{
			ID:              "4ef2ac82-ebcf-4b6c-a39b-6fd27fe09506",
			OrgID:           "6c6b3b6d-24e5-4f70-9896-4a49609cd61a",
			Type:            "broker_integration",
			IntegrationType: "github",
		},
	}

	actualIntegrations, _, err := client.Brokers.ListIntegrations(ctx, "tenant-id", "connection-id")

	assert.NoError(t, err)
	assert.Equal(t, 1, len(actualIntegrations), "expect 1 broker integration")
	assert.Equal(t, expectedIntegrations, actualIntegrations)
}

func TestBrokers_ListIntegrations_emptyTenantID(t *testing.T) {
	_, _, err := client.Brokers.ListIntegrations(ctx, "", "connection-id")

	assert.Error(t, err)
	assert.ErrorContains(t, err, "tenant id must be supplied")
}

func TestBrokers_ListIntegrations_emptyConnectionID(t *testing.T) {
	_, _, err := client.Brokers.ListIntegrations(ctx, "tenant-id", "")

	assert.Error(t, err)
	assert.ErrorContains(t, err, "connection id must be supplied")
}

func TestBrokers_CreateIntegration(t *testing.T) {
	setup(t)
	defer teardown()

	mux.HandleFunc("/tenants/tenant-id/brokers/connections/connection-id/orgs/org-id/integration", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assertRequestAPIVersion(t, r, brokersAPIVersion)
		_, _ = fmt.Fprint(w, `
{
  "jsonapi": { "version": "1.0" },
  "data": {
    "id": "a9d79dc9-63c5-4b5d-ae5c-5c42bc2f3d38",
    "type": "broker_connection",
    "org_id": "6c6b3b6d-24e5-4f70-9896-4a49609cd61a"
  },
  "links": {}
}
`)
	})
	expectedIntegration := &BrokerIntegration{
		ID:    "a9d79dc9-63c5-4b5d-ae5c-5c42bc2f3d38",
		Type:  "broker_connection",
		OrgID: "6c6b3b6d-24e5-4f70-9896-4a49609cd61a",
	}

	actualIntegration, _, err := client.Brokers.CreateIntegration(
		ctx, "tenant-id", "connection-id", "org-id",
		&BrokerIntegrationCreateRequest{
			Type: "github",
		},
	)

	assert.NoError(t, err)
	assert.Equal(t, expectedIntegration, actualIntegration)
}

func TestBrokers_CreateIntegration_tenantID(t *testing.T) {
	_, _, err := client.Brokers.CreateIntegration(ctx, "", "connection-id", "org-id", nil)

	assert.Error(t, err)
	assert.ErrorContains(t, err, "tenant id must be supplied")
}

func TestBrokers_CreateIntegration_emptyConnectionID(t *testing.T) {
	_, _, err := client.Brokers.CreateIntegration(ctx, "tenant-id", "", "org-id", nil)

	assert.Error(t, err)
	assert.ErrorContains(t, err, "connection id must be supplied")
}

func TestBrokers_CreateIntegration_emptyOrgID(t *testing.T) {
	_, _, err := client.Brokers.CreateIntegration(ctx, "tenant-id", "connection-id", "", nil)

	assert.Error(t, err)
	assert.ErrorContains(t, err, "org id must be supplied")
}

func TestBrokers_CreateIntegration_emptyPayload(t *testing.T) {
	_, _, err := client.Brokers.CreateIntegration(ctx, "tenant-id", "connection-id", "org-id", nil)

	assert.Error(t, err)
	assert.ErrorContains(t, err, "payload must be supplied")
}

func TestBrokers_DeleteIntegration(t *testing.T) {
	setup(t)
	defer teardown()

	mux.HandleFunc("/tenants/tenant-id/brokers/connections/connection-id/orgs/org-id/integrations/integration-id", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		assertRequestAPIVersion(t, r, brokersAPIVersion)
	})

	_, err := client.Brokers.DeleteIntegration(ctx, "tenant-id", "connection-id", "org-id", "integration-id")

	assert.NoError(t, err)
}

func TestBrokers_DeleteIntegration_tenantID(t *testing.T) {
	_, err := client.Brokers.DeleteIntegration(ctx, "", "connection-id", "org-id", "integration-id")

	assert.Error(t, err)
	assert.ErrorContains(t, err, "tenant id must be supplied")
}

func TestBrokers_DeleteIntegration_emptyConnectionID(t *testing.T) {
	_, err := client.Brokers.DeleteIntegration(ctx, "tenant-id", "", "org-id", "integration-id")

	assert.Error(t, err)
	assert.ErrorContains(t, err, "connection id must be supplied")
}

func TestBrokers_DeleteIntegration_emptyOrgID(t *testing.T) {
	_, err := client.Brokers.DeleteIntegration(ctx, "tenant-id", "connection-id", "", "integration-id")

	assert.Error(t, err)
	assert.ErrorContains(t, err, "org id must be supplied")
}

func TestBrokers_DeleteIntegration_emptyIntegrationID(t *testing.T) {
	_, err := client.Brokers.DeleteIntegration(ctx, "tenant-id", "connection-id", "org-id", "")

	assert.Error(t, err)
	assert.ErrorContains(t, err, "integration id must be supplied")
}

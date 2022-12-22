package snyk

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIntegrations_List(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/org/long-uuid/integrations", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		_, _ = fmt.Fprint(w, `
{
  "acr": "3aaa59f8-4a29-41ff-a074-911365dbb400",
  "github": "fef79ea8-3ad4-4598-ae11-d8730ede2382"
}
`)
	})
	expectedIntegrations := Integrations{
		ACRIntegrationType:    "3aaa59f8-4a29-41ff-a074-911365dbb400",
		GitHubIntegrationType: "fef79ea8-3ad4-4598-ae11-d8730ede2382",
	}

	actualIntegrations, _, err := client.Integrations.List(ctx, "long-uuid")

	assert.NoError(t, err)
	assert.Equal(t, expectedIntegrations, actualIntegrations)
}

func TestIntegrations_List_emptyOrganizationID(t *testing.T) {
	setup()
	defer teardown()

	_, _, err := client.Integrations.List(ctx, "")

	assert.Error(t, err)
	assert.Equal(t, ErrEmptyArgument, err)
}

func TestIntegrations_GetByType(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/org/long-uuid/integrations/github", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		_, _ = fmt.Fprint(w, `
{
  "id": "fef79ea8-3ad4-4598-ae11-d8730ede2382"
}
`)
	})
	expectedIntegration := &Integration{ID: "fef79ea8-3ad4-4598-ae11-d8730ede2382"}

	actualIntegration, _, err := client.Integrations.GetByType(ctx, "long-uuid", GitHubIntegrationType)

	assert.NoError(t, err)
	assert.Equal(t, expectedIntegration, actualIntegration)
}

func TestIntegrations_GetByType_emptyOrganizationID(t *testing.T) {
	setup()
	defer teardown()

	_, _, err := client.Integrations.GetByType(ctx, "", GitHubIntegrationType)

	assert.Error(t, err)
	assert.Equal(t, ErrEmptyArgument, err)
}

func TestIntegrations_GetByType_emptyIntegrationType(t *testing.T) {
	setup()
	defer teardown()

	_, _, err := client.Integrations.GetByType(ctx, "long-uuid", "")

	assert.Error(t, err)
	assert.Equal(t, ErrEmptyArgument, err)
}

func TestIntegrations_Create(t *testing.T) {
	setup()
	defer teardown()

	input := &IntegrationCreateRequest{
		Integration: &Integration{
			Type: DockerHubIntegrationType,
			Credentials: &IntegrationCredentials{
				Username: "test-user",
				Password: "secret-password",
			},
		},
	}
	mux.HandleFunc("/org/long-uuid/integrations", func(w http.ResponseWriter, r *http.Request) {
		v := new(IntegrationCreateRequest)
		_ = json.NewDecoder(r.Body).Decode(v)
		assert.Equal(t, input, v)
		assert.Equal(t, http.MethodPost, r.Method)
		_, _ = fmt.Fprint(w, `
{
  "id": "bd3cd15a-0b2d-4ca0-aa2e-1f7ae5a071ee"
}
`)
	})
	expectedIntegration := &Integration{ID: "bd3cd15a-0b2d-4ca0-aa2e-1f7ae5a071ee"}

	actualIntegration, _, err := client.Integrations.Create(ctx, "long-uuid", input)

	assert.NoError(t, err)
	assert.Equal(t, expectedIntegration, actualIntegration)
}

func TestIntegrations_Create_emptyOrganizationID(t *testing.T) {
	setup()
	defer teardown()

	_, _, err := client.Integrations.Create(ctx, "", &IntegrationCreateRequest{})

	assert.Error(t, err)
	assert.Equal(t, ErrEmptyArgument, err)
}

func TestIntegrations_Create_emptyPayload(t *testing.T) {
	setup()
	defer teardown()

	_, _, err := client.Integrations.Create(ctx, "long-uuid", nil)

	assert.Error(t, err)
	assert.Equal(t, ErrEmptyPayloadNotAllowed, err)
}

func TestIntegrations_Update(t *testing.T) {
	setup()
	defer teardown()

	input := &IntegrationUpdateRequest{
		Integration: &Integration{
			Type: GitHubIntegrationType,
			Credentials: &IntegrationCredentials{
				Token: "updated-token",
			},
		},
	}
	mux.HandleFunc("/org/long-uuid/integrations/fef79ea8-3ad4-4598-ae11-d8730ede2382", func(w http.ResponseWriter, r *http.Request) {
		v := new(IntegrationUpdateRequest)
		_ = json.NewDecoder(r.Body).Decode(v)
		assert.Equal(t, input, v)
		assert.Equal(t, http.MethodPut, r.Method)
		_, _ = fmt.Fprint(w, `
{
  "id": "fef79ea8-3ad4-4598-ae11-d8730ede2382"
}
`)
	})
	expectedIntegration := &Integration{ID: "fef79ea8-3ad4-4598-ae11-d8730ede2382"}

	actualIntegration, _, err := client.Integrations.Update(ctx, "long-uuid", "fef79ea8-3ad4-4598-ae11-d8730ede2382", input)

	assert.NoError(t, err)
	assert.Equal(t, expectedIntegration, actualIntegration)
}

func TestIntegrations_Update_emptyOrganizationID(t *testing.T) {
	setup()
	defer teardown()

	_, _, err := client.Integrations.Update(ctx, "", "integration-id", &IntegrationUpdateRequest{})

	assert.Error(t, err)
	assert.Equal(t, ErrEmptyArgument, err)
}

func TestIntegrations_Update_emptyIntegrationID(t *testing.T) {
	setup()
	defer teardown()

	_, _, err := client.Integrations.Update(ctx, "long-uuid", "", &IntegrationUpdateRequest{})

	assert.Error(t, err)
	assert.Equal(t, ErrEmptyArgument, err)
}

func TestIntegrations_Update_emptyPayload(t *testing.T) {
	setup()
	defer teardown()

	_, _, err := client.Integrations.Update(ctx, "long-uuid", "integration-id", nil)

	assert.Error(t, err)
	assert.Equal(t, ErrEmptyPayloadNotAllowed, err)
}

func TestIntegrations_DeleteCredentials(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/org/long-uuid/integrations/fef79ea8-3ad4-4598-ae11-d8730ede2382/authentication", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
	})

	_, err := client.Integrations.DeleteCredentials(ctx, "long-uuid", "fef79ea8-3ad4-4598-ae11-d8730ede2382")

	assert.NoError(t, err)
}

func TestIntegrations_DeleteCredentials_emptyOrganizationID(t *testing.T) {
	setup()
	defer teardown()

	_, err := client.Integrations.DeleteCredentials(ctx, "", "fef79ea8-3ad4-4598-ae11-d8730ede2382")

	assert.Error(t, err)
	assert.Equal(t, ErrEmptyArgument, err)
}

func TestIntegrations_DeleteCredentials_emptyIntegrationID(t *testing.T) {
	setup()
	defer teardown()

	_, err := client.Integrations.DeleteCredentials(ctx, "long-uuid", "")

	assert.Error(t, err)
	assert.Equal(t, ErrEmptyArgument, err)
}

func TestIntegrations_GetSettings(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/org/long-uuid/integrations/fef79ea8-3ad4-4598-ae11-d8730ede2382/settings", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		_, _ = fmt.Fprint(w, `
{
  "autoDepUpgradeEnabled": true,
  "autoDepUpgradeIgnoredDependencies": ["lodash"],
  "autoDepUpgradeLimit": 3,
  "dockerfileSCMEnabled": false,
  "isMajorUpgradeEnabled": false,
  "pullRequestFailOnAnyVulns": false,
  "pullRequestFailOnlyForIssuesWithFix": true,
  "pullRequestFailOnlyForHighSeverity": false,
  "pullRequestTestEnabled": true
}
`)
	})
	expectedSettings := &IntegrationSettings{
		DependencyAutoUpgradeEnabled:                  ptr(true),
		DependencyAutoUpgradeIgnoredDependencies:      []string{"lodash"},
		DependencyAutoUpgradePullRequestLimit:         3,
		DependencyAutoUpgradeIncludeMajorVersion:      ptr(false),
		DockerfileDetectionEnabled:                    ptr(false),
		PullRequestFailOnAnyIssue:                     ptr(false),
		PullRequestFailOnlyForIssuesWithFix:           ptr(true),
		PullRequestFailOnlyForHighAndCriticalSeverity: ptr(false),
		PullRequestTestEnabled:                        ptr(true),
	}

	actualSettings, _, err := client.Integrations.GetSettings(ctx, "long-uuid", "fef79ea8-3ad4-4598-ae11-d8730ede2382")

	assert.NoError(t, err)
	assert.Equal(t, expectedSettings, actualSettings)
}

func TestIntegrations_GetSettings_emptyOrganizationID(t *testing.T) {
	setup()
	defer teardown()

	_, _, err := client.Integrations.GetSettings(ctx, "", "integration-id")

	assert.Error(t, err)
	assert.Equal(t, ErrEmptyArgument, err)
}

func TestIntegrations_UpdateSettings(t *testing.T) {
	setup()
	defer teardown()

	input := &IntegrationSettingsUpdateRequest{
		IntegrationSettings: &IntegrationSettings{
			DependencyAutoUpgradeEnabled: ptr(true),
			DockerfileDetectionEnabled:   ptr(false),
			PullRequestTestEnabled:       ptr(true),
		},
	}
	mux.HandleFunc("/org/long-uuid/integrations/fef79ea8-3ad4-4598-ae11-d8730ede2382/settings", func(w http.ResponseWriter, r *http.Request) {
		v := new(IntegrationSettingsUpdateRequest)
		_ = json.NewDecoder(r.Body).Decode(v)
		assert.Equal(t, input, v)
		assert.Equal(t, http.MethodPut, r.Method)
		_, _ = fmt.Fprint(w, `
{
  "autoDepUpgradeEnabled": true,
  "dockerfileSCMEnabled": false,
  "pullRequestTestEnabled": true
}
`)
	})
	expectedSettings := &IntegrationSettings{
		DependencyAutoUpgradeEnabled: ptr(true),
		DockerfileDetectionEnabled:   ptr(false),
		PullRequestTestEnabled:       ptr(true),
	}

	actualSettings, _, err := client.Integrations.UpdateSettings(ctx, "long-uuid", "fef79ea8-3ad4-4598-ae11-d8730ede2382", input)

	assert.NoError(t, err)
	assert.Equal(t, expectedSettings, actualSettings)
}

func TestIntegrations_UpdateSettings_emptyOrganizationID(t *testing.T) {
	setup()
	defer teardown()

	_, _, err := client.Integrations.UpdateSettings(ctx, "", "integration-id", &IntegrationSettingsUpdateRequest{})

	assert.Error(t, err)
	assert.Equal(t, ErrEmptyArgument, err)
}

func TestIntegrations_UpdateSettings_emptyIntegrationID(t *testing.T) {
	setup()
	defer teardown()

	_, _, err := client.Integrations.UpdateSettings(ctx, "long-uuid", "", &IntegrationSettingsUpdateRequest{})

	assert.Error(t, err)
	assert.Equal(t, ErrEmptyArgument, err)
}

func TestIntegrations_UpdateSettings_emptyPayload(t *testing.T) {
	setup()
	defer teardown()

	_, _, err := client.Integrations.UpdateSettings(ctx, "long-uuid", "integration-id", nil)

	assert.Error(t, err)
	assert.Equal(t, ErrEmptyPayloadNotAllowed, err)
}

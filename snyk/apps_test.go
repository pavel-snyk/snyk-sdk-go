package snyk

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestApps_DeleteAppInstallFromOrg(t *testing.T) {
	setup(t)
	defer teardown()

	mux.HandleFunc("/orgs/org-id/apps/installs/install-id", func(w http.ResponseWriter, r *http.Request) {
		assertRequestAPIVersion(t, r, appsAPIVersion)
		w.WriteHeader(http.StatusNoContent)
	})

	response, err := client.Apps.DeleteAppInstallFromOrg(ctx, "org-id", "install-id")

	assert.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, response.StatusCode)
}

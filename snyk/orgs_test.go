package snyk

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOrgs_ListAccessibleOrgs(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/orgs", func(w http.ResponseWriter, r *http.Request) {
		assertRequestAPIVersion(t, r, orgsAPIVersion)
		assert.Equal(t, "20", r.URL.Query().Get("limit"))
		w.Header().Set(headerSnykVersionServed, "2024-02-28")
		_, _ = fmt.Fprint(w, `{"data":[],"links":{}}`)
	})

	_, response, err := client.Orgs.ListAccessibleOrgs(ctx, &ListOrganizationOptions{
		ListOptions: ListOptions{Limit: 20},
	})

	assert.NoError(t, err)
	assert.Equal(t, "2024-02-28", response.ServedAPIVersion)
}

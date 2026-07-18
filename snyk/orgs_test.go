package snyk

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOrgs_ListAccessibleOrgs(t *testing.T) {
	setup(t)
	defer teardown()

	mux.HandleFunc("/orgs", func(w http.ResponseWriter, r *http.Request) {
		assertRequestAPIVersion(t, r, orgsAPIVersion)
		assert.Equal(t, "cursor", r.URL.Query().Get("starting_after"))
		assert.Equal(t, "20", r.URL.Query().Get("limit"))
		assert.Equal(t, "group-id", r.URL.Query().Get("group_id"))
		assert.Equal(t, "tenant", r.URL.Query().Get("expand"))
		w.Header().Set(headerSnykVersionServed, "2024-02-28")
		_, _ = fmt.Fprint(w, `{"data":[],"links":{}}`)
	})

	_, response, err := client.Orgs.ListAccessibleOrgs(ctx, &ListOrganizationOptions{
		ListOptions: ListOptions{StartingAfter: "cursor", Limit: 20},
		GroupID:     "group-id",
		Expand:      "tenant",
	})

	assert.NoError(t, err)
	assert.Equal(t, "2024-02-28", response.ServedAPIVersion)
}

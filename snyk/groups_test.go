package snyk

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGroups_All(t *testing.T) {
	setup()
	defer teardown()

	requestCount := 0
	mux.HandleFunc("/groups", func(w http.ResponseWriter, r *http.Request) {
		assertRequestAPIVersion(t, r, groupsAPIVersion)
		requestCount++
		switch requestCount {
		case 1:
			assert.Empty(t, r.URL.Query().Get("starting_after"))
			_, _ = fmt.Fprint(w, `{
  "data":[{"id":"group-1","type":"group","attributes":{"name":"First"}}],
  "links":{"next":"/groups?version=server-value&starting_after=cursor-1"}
}`)
		case 2:
			assert.Equal(t, "cursor-1", r.URL.Query().Get("starting_after"))
			_, _ = fmt.Fprint(w, `{
  "data":[{"id":"group-2","type":"group","attributes":{"name":"Second"}}],
  "links":{}
}`)
		default:
			http.Error(w, "unexpected pagination request", http.StatusInternalServerError)
		}
	})

	seq, iterErr := client.Groups.All(ctx, nil)
	var groupIDs []string
	for group := range seq {
		groupIDs = append(groupIDs, group.ID)
	}

	assert.NoError(t, iterErr())
	assert.Equal(t, 2, requestCount)
	assert.Equal(t, []string{"group-1", "group-2"}, groupIDs)
}

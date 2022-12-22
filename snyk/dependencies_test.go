package snyk

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDependencies_List(t *testing.T) {
	setup()
	defer teardown()

	expected := []Dependency{
		{
			ID:                         "dep1",
			Name:                       "dep-name",
			Type:                       "abc",
			Version:                    "2",
			LatestVersion:              "3",
			LatestVersionPublishedDate: ptr(time.Now().UTC()),
			FirstPublishedDate:         ptr(time.Now().UTC()),
			IsDeprecated:               true,
			DeprecatedVersions:         []string{"1", "2"},
			DependenciesWithIssues:     []string{"a", "b"},
			IssuesCritical:             1,
			IssuesHigh:                 2,
			IssuesMedium:               3,
			IssuesLow:                  4,
			Licenses:                   nil,
			Projects: []Project{
				{
					ID:   "proj1",
					Name: "proj-name",
				},
			},
			Copyright: []string{
				"some",
				"license",
			},
		},
	}
	mux.HandleFunc("/org/long-uuid/dependencies", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		b, err := json.Marshal(
			dependenciesRoot{
				Total:   1,
				Results: expected,
			},
		)
		if err != nil {
			http.Error(w, "unable to marshal response: "+err.Error(), http.StatusBadRequest)
			return
		}
		if _, err := w.Write(b); err != nil {
			http.Error(w, "failed to write", http.StatusBadRequest)
			return
		}
	})

	actual, _, err := client.Dependencies.List(ctx, "long-uuid")

	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestDependencies_List_emptyOrganizationID(t *testing.T) {
	setup()
	defer teardown()

	_, _, err := client.Dependencies.List(ctx, "")

	assert.Error(t, err)
	assert.Equal(t, ErrEmptyArgument, err)
}

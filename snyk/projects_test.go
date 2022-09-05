package snyk

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProject_List(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/org/long-uuid/projects", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		_, _ = fmt.Fprint(w, `
{
  "org": {
    "id": "long-uuid",
    "name": "test-org"
  },
  "projects": [
    {
      "id": "e8feca4a-4ebc-494f-80d9-f8b0532188da",
      "name": "test-org/test-project",
      "origin": "github"
    }
  ]
}
`)
	})
	expectedProjects := []Project{
		{
			ID:     "e8feca4a-4ebc-494f-80d9-f8b0532188da",
			Name:   "test-org/test-project",
			Origin: "github",
		},
	}

	actualProjects, _, err := client.Projects.List(ctx, "long-uuid")

	assert.NoError(t, err)
	assert.Equal(t, expectedProjects, actualProjects)
}

func TestProject_List_emptyOrganizationID(t *testing.T) {
	_, _, err := client.Projects.List(ctx, "")

	assert.Error(t, err)
	assert.Equal(t, ErrEmptyArgument, err)
}

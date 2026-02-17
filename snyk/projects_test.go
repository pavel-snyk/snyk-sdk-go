package snyk

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestProject_List(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/orgs/org-id/projects", func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, `
{
  "jsonapi": { "version": "1.0" },
  "data": [
    {
      "type": "project",
      "id": "7f76f013-a3fd-4151-95e2-b3a347dc8546",
      "meta": {},
      "attributes": {
        "name": "test-user/test-project-1:package.json",
        "type": "npm",
        "target_file": "package.json",
        "target_reference": "main",
        "origin": "gitlab",
        "created": "2012-12-12T12:12:12.127Z",
        "status": "active",
        "business_criticality": [ "medium", "low" ],
        "environment": [ "external" ],
        "lifecycle": [ "development" ],
        "tags": [],
        "read_only": false,
        "settings": { "recurring_tests": { "frequency": "weekly" }, "pull_requests": {} }
      },
      "relationships": {
        "organization": {
          "data": { "type": "org", "id": "2bd5babe-884b-49c5-9e82-0ce9f2a3a147" },
          "links": {}
        },
        "target": {
          "data": { "type": "target", "id": "ccd00feb-45c1-4e03-80c3-8eba50fac80a" },
          "links": { "related": "/rest/orgs/2bd5babe-884b-49c5-9e82-0ce9f2a3a147/targets/ccd00feb-45c1-4e03-80c3-8eba50fac80a" }
        },
        "owner": {
          "data": { "type": "user", "id": "f210d0ed-61b2-4588-a997-56346f61a7a7" },
          "links": { "related": "/rest/orgs/2bd5babe-884b-49c5-9e82-0ce9f2a3a147/users/f210d0ed-61b2-4588-a997-56346f61a7a7" }
        },
        "importer": {
          "data": { "type": "user", "id": "f210d0ed-61b2-4588-a997-56346f61a7a7" },
          "links": { "related": "/rest/orgs/2bd5babe-884b-49c5-9e82-0ce9f2a3a147/users/f210d0ed-61b2-4588-a997-56346f61a7a7" }
        }
      }
    }
  ],
  "links": {}
}
`)
	})
	createdAtTime, _ := time.Parse(time.RFC3339, "2012-12-12T12:12:12.127Z")
	expectedProjects := []Project{{
		ID:   "7f76f013-a3fd-4151-95e2-b3a347dc8546",
		Type: "project",
		Attributes: &ProjectAttributes{
			CreatedAt:       createdAtTime,
			Name:            "test-user/test-project-1:package.json",
			Origin:          "gitlab",
			Status:          "active",
			TargetFile:      "package.json",
			TargetReference: "main",
			Type:            "npm",
		},
	}}

	actualProjects, _, err := client.Projects.List(ctx, "org-id", nil)

	assert.NoError(t, err)
	assert.Equal(t, 1, len(actualProjects))
	assert.Equal(t, expectedProjects, actualProjects)
}

func TestProject_List_emptyOrgID(t *testing.T) {
	_, _, err := client.Projects.List(ctx, "", nil)

	assert.Error(t, err)
	assert.ErrorContains(t, err, "orgID must be supplied")
}

func TestProject_Get(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/orgs/org-id/projects/project-id", func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, `
{
  "jsonapi": { "version": "1.0" },
  "data": {
    "type": "project",
    "id": "7f76f013-a3fd-4151-95e2-b3a347dc8546",
    "meta": {},
    "attributes": {
      "name": "test-user/test-project-1:package.json",
      "type": "npm",
      "target_file": "package.json",
      "target_reference": "main",
      "origin": "gitlab",
      "created": "2012-12-12T12:12:12.127Z",
      "status": "active",
      "business_criticality": [ "medium", "low" ],
      "environment": [ "external" ],
      "lifecycle": [ "development" ],
      "tags": [],
      "read_only": false,
      "settings": { "recurring_tests": { "frequency": "weekly" }, "pull_requests": {} }
    },
    "relationships": {
      "organization": {
        "data": { "type": "org", "id": "2bd5babe-884b-49c5-9e82-0ce9f2a3a147" },
        "links": {}
      },
      "target": {
        "data": { "type": "target", "id": "ccd00feb-45c1-4e03-80c3-8eba50fac80a" },
        "links": { "related": "/rest/orgs/2bd5babe-884b-49c5-9e82-0ce9f2a3a147/targets/ccd00feb-45c1-4e03-80c3-8eba50fac80a" }
      },
      "owner": {
        "data": { "type": "user", "id": "f210d0ed-61b2-4588-a997-56346f61a7a7" },
        "links": { "related": "/rest/orgs/2bd5babe-884b-49c5-9e82-0ce9f2a3a147/users/f210d0ed-61b2-4588-a997-56346f61a7a7" }
			},
      "importer": {
        "data": { "type": "user", "id": "f210d0ed-61b2-4588-a997-56346f61a7a7" },
        "links": { "related": "/rest/orgs/2bd5babe-884b-49c5-9e82-0ce9f2a3a147/users/f210d0ed-61b2-4588-a997-56346f61a7a7" }
      }
    }
  },
  "links": {}
}
`)
	})
	createdAtTime, _ := time.Parse(time.RFC3339, "2012-12-12T12:12:12.127Z")
	expectedProject := &Project{
		ID:   "7f76f013-a3fd-4151-95e2-b3a347dc8546",
		Type: "project",
		Attributes: &ProjectAttributes{
			CreatedAt:       createdAtTime,
			Name:            "test-user/test-project-1:package.json",
			Origin:          "gitlab",
			Status:          "active",
			TargetFile:      "package.json",
			TargetReference: "main",
			Type:            "npm",
		},
	}

	actualProject, _, err := client.Projects.Get(ctx, "org-id", "project-id")

	assert.NoError(t, err)
	assert.Equal(t, expectedProject, actualProject)
}

func TestProject_Get_emptyOrgID(t *testing.T) {
	_, _, err := client.Projects.Get(ctx, "", "project-id")

	assert.Error(t, err)
	assert.ErrorContains(t, err, "orgID must be supplied")
}

func TestProject_Get_emptyProjectID(t *testing.T) {
	_, _, err := client.Projects.Get(ctx, "org-id", "")

	assert.Error(t, err)
	assert.ErrorContains(t, err, "projectID must be supplied")
}

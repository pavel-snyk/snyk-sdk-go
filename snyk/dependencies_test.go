package snyk

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDependencies_List(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/org/long-uuid/dependencies", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)

		_, _ = fmt.Fprintf(w, `
{
  "results": [
    {
      "id": "gulp@3.9.1",
      "name": "gulp",
      "version": "3.9.1",
      "latestVersion": "4.0.0",
      "latestVersionPublishedDate": "2018-01-01T01:29:06.863Z",
      "firstPublishedDate": "2013-07-04T23:27:07.828Z",
      "isDeprecated": false,
      "deprecatedVersions": [
        "0.0.1",
        "0.0.2",
        "0.0.3"
      ],
      "licenses": [
        {
          "id": "snyk:lic:npm:gulp:MIT",
          "title": "MIT license",
          "license": "MIT"
        }
      ],
      "dependenciesWithIssues": [
        "minimatch@2.0.10",
        "minimatch@0.2.14"
      ],
      "type": "npm",
      "projects": [
        {
          "name": "atokeneduser/goof",
          "id": "6d5813be-7e6d-4ab8-80c2-1e3e2a454545"
        }
      ],
      "copyright": [
        "Copyright (c) 2013-2018 Blaine Bublitz <blaine.bublitz@gmail.com>",
        "Copyright (c) Eric Schoffstall <yo@contra.io> and other contributors"
      ]
    }
  ],
  "total": 1
}`)
	})

	expected := []Dependency{
		{
			ID:                         "gulp@3.9.1",
			Name:                       "gulp",
			Type:                       "npm",
			Version:                    "3.9.1",
			LatestVersion:              "4.0.0",
			LatestVersionPublishedDate: ptr(mustParseTime(t, "2018-01-01T01:29:06.863Z")),
			FirstPublishedDate:         ptr(mustParseTime(t, "2013-07-04T23:27:07.828Z")),
			IsDeprecated:               false,
			DeprecatedVersions: []string{
				"0.0.1",
				"0.0.2",
				"0.0.3",
			},
			DependenciesWithIssues: []string{
				"minimatch@2.0.10",
				"minimatch@0.2.14",
			},
			Licenses: []DependencyLicense{
				{
					ID:      "snyk:lic:npm:gulp:MIT",
					Title:   "MIT license",
					License: "MIT",
				},
			},
			Projects: []Project{
				{
					ID:   "6d5813be-7e6d-4ab8-80c2-1e3e2a454545",
					Name: "atokeneduser/goof",
				},
			},
			Copyright: []string{
				"Copyright (c) 2013-2018 Blaine Bublitz <blaine.bublitz@gmail.com>",
				"Copyright (c) Eric Schoffstall <yo@contra.io> and other contributors",
			},
		},
	}
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

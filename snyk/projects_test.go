package snyk

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProject_MarshalJSON(t *testing.T) {
	project := Project{
		ID:               "project-id",
		Name:             "example",
		ProjectType:      ProjectType("npm"),
		Origin:           ProjectOrigin("cli"),
		MonitoringStatus: ProjectMonitoringStatusActive,
		TargetPath:       "",
		TargetReference:  "",
	}

	got, err := json.Marshal(project)
	require.NoError(t, err)
	require.JSONEq(t, `{
		"id": "project-id",
		"name": "example",
		"project_type": "npm",
		"origin": "cli",
		"monitoring_status": "active",
		"target_path": "",
		"target_reference": ""
	}`, string(got))
}

func TestProjects_projectFromResource(t *testing.T) {
	resource := projectResource{
		ID:   "project-id",
		Type: "project",
		Attributes: &projectAttributes{
			Name:            "example",
			Type:            "future-project-type",
			Origin:          "future-origin",
			Status:          "future-status",
			TargetFile:      "path/to/project",
			TargetReference: "future-reference",
		},
	}

	project, err := projectFromResource(resource)

	require.NoError(t, err)
	assert.Equal(t, Project{
		ID:               "project-id",
		Name:             "example",
		ProjectType:      "future-project-type",
		Origin:           "future-origin",
		MonitoringStatus: "future-status",
		TargetPath:       "path/to/project",
		TargetReference:  "future-reference",
	}, project)
}

func TestProjects_projectFromResource_rejectsStructuralViolations(t *testing.T) {
	tests := []struct {
		name     string
		resource projectResource
		wantErr  string
	}{
		{
			name:     "empty resource ID",
			resource: projectResource{Type: "project", Attributes: &projectAttributes{}},
			wantErr:  "resource ID is empty",
		},
		{
			name:     "wrong resource type",
			resource: projectResource{ID: "project-id", Type: "target", Attributes: &projectAttributes{}},
			wantErr:  `project "project-id": resource type is "target", expected "project"`,
		},
		{
			name:     "missing attributes",
			resource: projectResource{ID: "project-id", Type: "project"},
			wantErr:  `project "project-id": attributes are missing`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := projectFromResource(tt.resource)

			assert.EqualError(t, err, tt.wantErr)
		})
	}
}

func TestProjects_List_success(t *testing.T) {
	setup(t)
	defer teardown()

	fixture := loadFixture(t, "projects_list_success.json")
	mux.HandleFunc("/orgs/org-id/projects", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, restAPIMediaType, r.Header.Get("Accept"))
		assertRequestAPIVersion(t, r, projectsAPIVersion)
		assert.Equal(t, url.Values{
			"ending_before":    {"ending-cursor"},
			"ids":              {"project-a,project-b"},
			"limit":            {"20"},
			"names":            {"project-a,project-b"},
			"names_start_with": {"service-,library-"},
			"origins":          {"github-cloud-app,cli"},
			"target_file":      {"package.json"},
			"target_reference": {"main"},
			"types":            {"npm,sast"},
			"version":          {projectsAPIVersion},
		}, r.URL.Query())
		_, _ = w.Write(fixture)
	})

	projects, response, err := client.Projects.List(ctx, "org-id", &ProjectListOptions{
		ListOptions:     ListOptions{EndingBefore: "ending-cursor", Limit: 20},
		IDs:             []string{"project-a", "project-b"},
		Names:           []string{"project-a", "project-b"},
		NamePrefixes:    []string{"service-", "library-"},
		ProjectTypes:    []ProjectType{"npm", "sast"},
		Origins:         []ProjectOrigin{"github-cloud-app", "cli"},
		TargetPath:      "package.json",
		TargetReference: "main",
	})

	require.NoError(t, err)
	require.NotNil(t, response)
	require.Len(t, projects, 2)
	assert.Equal(t, Project{
		ID:               "11111111-1111-1111-1111-111111111111",
		Name:             "fake-image:/usr/local/lib/node_modules",
		ProjectType:      "npm",
		Origin:           "cli",
		MonitoringStatus: ProjectMonitoringStatusActive,
		TargetPath:       "/usr/local/lib/node_modules",
		TargetReference:  "docker-image|fake-image",
	}, projects[0])
	assert.Equal(t, Project{
		ID:               "22222222-2222-2222-2222-222222222222",
		Name:             "fake-org/fake-repository",
		ProjectType:      "sast",
		Origin:           "github-cloud-app",
		MonitoringStatus: ProjectMonitoringStatusInactive,
		TargetPath:       "",
		TargetReference:  "main",
	}, projects[1])
	require.NotNil(t, response.Links)
	assert.Equal(t, "next-cursor", mustStartingAfter(t, response.Links.Next))
}

func TestProjects_List_nilOptions(t *testing.T) {
	setup(t)
	defer teardown()

	mux.HandleFunc("/orgs/org-id/projects", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, url.Values{"version": {projectsAPIVersion}}, r.URL.Query())
		_, _ = fmt.Fprint(w, `{"data":[],"links":{}}`)
	})

	projects, _, err := client.Projects.List(ctx, "org-id", nil)

	require.NoError(t, err)
	assert.NotNil(t, projects)
	assert.Empty(t, projects)
}

func TestProjects_List_rejectsMalformedResourceWithContext(t *testing.T) {
	setup(t)
	defer teardown()

	mux.HandleFunc("/orgs/org-id/projects", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = fmt.Fprint(w, `{
			"data": [
				{"id":"project-1","type":"project","attributes":{}},
				{"id":"project-2","type":"target","attributes":{}}
			]
		}`)
	})

	projects, response, err := client.Projects.List(ctx, "org-id", nil)

	assert.Nil(t, projects)
	require.NotNil(t, response)
	assert.EqualError(t, err, `convert project at index 1: project "project-2": resource type is "target", expected "project"`)
}

func TestProjects_List_rejectsMissingOrNullData(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{name: "missing data", body: `{"links": {}}`},
		{name: "null data", body: `{"data": null, "links": {}}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setup(t)
			defer teardown()

			mux.HandleFunc("/orgs/org-id/projects", func(w http.ResponseWriter, _ *http.Request) {
				_, _ = fmt.Fprint(w, tt.body)
			})

			projects, response, err := client.Projects.List(ctx, "org-id", nil)

			assert.Nil(t, projects)
			require.NotNil(t, response)
			assert.EqualError(t, err, "convert projects: response data is missing")
		})
	}
}

func TestProjects_All_multiplePages(t *testing.T) {
	setup(t)
	defer teardown()

	page1 := loadFixture(t, "projects_all_page_1.json")
	page2 := loadFixture(t, "projects_all_page_2.json")
	options := &ProjectListOptions{
		ListOptions:     ListOptions{StartingAfter: "initial-cursor", Limit: 10},
		Names:           []string{"fake-project"},
		Origins:         []ProjectOrigin{"github-cloud-app"},
		TargetReference: "main",
	}
	requestCount := 0
	mux.HandleFunc("/orgs/org-id/projects", func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		assert.Equal(t, http.MethodGet, r.Method)
		assertRequestAPIVersion(t, r, projectsAPIVersion)
		switch requestCount {
		case 1:
			assert.Equal(t, url.Values{
				"limit":            {"10"},
				"names":            {"fake-project"},
				"origins":          {"github-cloud-app"},
				"starting_after":   {"initial-cursor"},
				"target_reference": {"main"},
				"version":          {projectsAPIVersion},
			}, r.URL.Query())
			_, _ = w.Write(page1)
		case 2:
			assert.Equal(t, url.Values{
				"limit":            {"10"},
				"names":            {"fake-project"},
				"origins":          {"github-cloud-app"},
				"starting_after":   {"next-cursor"},
				"target_reference": {"main"},
				"version":          {projectsAPIVersion},
			}, r.URL.Query())
			_, _ = w.Write(page2)
		default:
			http.Error(w, "unexpected pagination request", http.StatusInternalServerError)
		}
	})

	seq, iterErr := client.Projects.All(ctx, "org-id", options)
	var projects []Project
	for project := range seq {
		projects = append(projects, project)
	}

	require.NoError(t, iterErr())
	require.Len(t, projects, 2)
	assert.Equal(t, []string{
		"11111111-1111-1111-1111-111111111111",
		"22222222-2222-2222-2222-222222222222",
	}, []string{projects[0].ID, projects[1].ID})
	assert.Equal(t, ProjectType("sast"), projects[0].ProjectType)
	assert.Equal(t, "Dockerfile", projects[1].TargetPath)
	assert.Equal(t, 2, requestCount)
	assert.Equal(t, &ProjectListOptions{
		ListOptions:     ListOptions{StartingAfter: "initial-cursor", Limit: 10},
		Names:           []string{"fake-project"},
		Origins:         []ProjectOrigin{"github-cloud-app"},
		TargetReference: "main",
	}, options)
}

func TestProjects_All_snapshotsOptionsAtConstruction(t *testing.T) {
	setup(t)
	defer teardown()

	options := &ProjectListOptions{
		ListOptions:     ListOptions{StartingAfter: "original-cursor", Limit: 10},
		IDs:             []string{"original-id"},
		Names:           []string{"original-name"},
		NamePrefixes:    []string{"original-prefix"},
		ProjectTypes:    []ProjectType{"original-type"},
		Origins:         []ProjectOrigin{"original-origin"},
		TargetPath:      "original-path",
		TargetReference: "original-reference",
	}
	seq, iterErr := client.Projects.All(ctx, "org-id", options)

	options.StartingAfter = "changed-cursor"
	options.EndingBefore = "changed-ending-cursor"
	options.Limit = 99
	options.IDs[0] = "mutated-id"
	options.Names[0] = "mutated-name"
	options.NamePrefixes[0] = "mutated-prefix"
	options.ProjectTypes[0] = "mutated-type"
	options.Origins[0] = "mutated-origin"
	options.TargetPath = "mutated-path"
	options.TargetReference = "mutated-reference"

	requestCount := 0
	mux.HandleFunc("/orgs/org-id/projects", func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		assert.Equal(t, url.Values{
			"ids":              {"original-id"},
			"limit":            {"10"},
			"names":            {"original-name"},
			"names_start_with": {"original-prefix"},
			"origins":          {"original-origin"},
			"starting_after":   {"original-cursor"},
			"target_file":      {"original-path"},
			"target_reference": {"original-reference"},
			"types":            {"original-type"},
			"version":          {projectsAPIVersion},
		}, r.URL.Query())
		_, _ = fmt.Fprint(w, `{"data":[],"links":{}}`)
	})

	for range seq {
	}

	require.NoError(t, iterErr())
	assert.Equal(t, 1, requestCount)
}

func TestProjects_All_restartsSequentialIterationFromInitialCursor(t *testing.T) {
	setup(t)
	defer teardown()

	page1 := loadFixture(t, "projects_all_page_1.json")
	page2 := loadFixture(t, "projects_all_page_2.json")
	requestCount := 0
	mux.HandleFunc("/orgs/org-id/projects", func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		switch requestCount {
		case 1, 3:
			assert.Equal(t, "initial-cursor", r.URL.Query().Get("starting_after"))
			_, _ = w.Write(page1)
		case 2, 4:
			assert.Equal(t, "next-cursor", r.URL.Query().Get("starting_after"))
			_, _ = w.Write(page2)
		default:
			http.Error(w, "unexpected pagination request", http.StatusInternalServerError)
		}
	})

	seq, iterErr := client.Projects.All(ctx, "org-id", &ProjectListOptions{ListOptions: ListOptions{StartingAfter: "initial-cursor"}})

	var firstIteration []string
	for project := range seq {
		firstIteration = append(firstIteration, project.ID)
	}
	require.NoError(t, iterErr())

	var secondIteration []string
	for project := range seq {
		secondIteration = append(secondIteration, project.ID)
	}
	require.NoError(t, iterErr())

	wantProjectIDs := []string{
		"11111111-1111-1111-1111-111111111111",
		"22222222-2222-2222-2222-222222222222",
	}
	assert.Equal(t, wantProjectIDs, firstIteration)
	assert.Equal(t, wantProjectIDs, secondIteration)
	assert.Equal(t, 4, requestCount)
}

func TestProjects_All_rejectsEndingBefore(t *testing.T) {
	setup(t)
	defer teardown()

	seq, iterErr := client.Projects.All(ctx, "org-id", &ProjectListOptions{
		ListOptions: ListOptions{EndingBefore: "previous-cursor"},
	})
	for range seq {
		t.Fatal("unexpected Project")
	}

	assert.EqualError(t, iterErr(), "ending-before pagination is not supported when iterating all projects")
}

func TestProjects_All_convertsEachPageAtomically(t *testing.T) {
	setup(t)
	defer teardown()

	requestCount := 0
	mux.HandleFunc("/orgs/org-id/projects", func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		if r.URL.Query().Get("starting_after") == "second-page" {
			_, _ = fmt.Fprint(w, `{
				"data": [
					{"id":"project-3","type":"project","attributes":{"name":"Third"}},
					{"id":"project-4","type":"target","attributes":{"name":"Fourth"}}
				],
				"links": {}
			}`)
			return
		}

		_, _ = fmt.Fprint(w, `{
			"data": [
				{"id":"project-1","type":"project","attributes":{"name":"First"}},
				{"id":"project-2","type":"project","attributes":{"name":"Second"}}
			],
			"links": {"next":"/rest/orgs/org-id/projects?starting_after=second-page"}
		}`)
	})

	seq, iterErr := client.Projects.All(ctx, "org-id", nil)
	var projectIDs []string
	for project := range seq {
		projectIDs = append(projectIDs, project.ID)
	}

	assert.Equal(t, []string{"project-1", "project-2"}, projectIDs)
	assert.EqualError(t, iterErr(), `convert project at index 1: project "project-4": resource type is "target", expected "project"`)
	assert.Equal(t, 2, requestCount)
}

func TestProjects_All_rejectsMissingOrNullData(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{name: "missing data", body: `{"links": {}}`},
		{name: "null data", body: `{"data": null, "links": {}}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setup(t)
			defer teardown()

			mux.HandleFunc("/orgs/org-id/projects", func(w http.ResponseWriter, _ *http.Request) {
				_, _ = fmt.Fprint(w, tt.body)
			})

			seq, iterErr := client.Projects.All(ctx, "org-id", nil)
			for range seq {
				t.Fatal("unexpected Project")
			}

			assert.EqualError(t, iterErr(), "convert projects: response data is missing")
		})
	}
}

func TestProjects_Get_success(t *testing.T) {
	setup(t)
	defer teardown()

	const projectID = "11111111-1111-1111-1111-111111111111"
	fixture := loadFixture(t, "projects_get_success.json")
	mux.HandleFunc("/orgs/org-id/projects/"+projectID, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, restAPIMediaType, r.Header.Get("Accept"))
		assertRequestAPIVersion(t, r, projectsAPIVersion)
		assert.Equal(t, url.Values{"version": {projectsAPIVersion}}, r.URL.Query())
		_, _ = w.Write(fixture)
	})

	project, response, err := client.Projects.Get(ctx, "org-id", projectID)

	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Equal(t, &Project{
		ID:               projectID,
		Name:             "fake-org/fake-repository:catalog-info.yaml",
		ProjectType:      "k8sconfig",
		Origin:           "github-cloud-app",
		MonitoringStatus: ProjectMonitoringStatusActive,
		TargetPath:       "catalog-info.yaml",
		TargetReference:  "main",
	}, project)
}

func TestProjects_Get_rejectsMissingOrNullData(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{name: "missing data", body: `{}`},
		{name: "null data", body: `{"data": null}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setup(t)
			defer teardown()

			mux.HandleFunc("/orgs/org-id/projects/project-id", func(w http.ResponseWriter, _ *http.Request) {
				_, _ = fmt.Fprint(w, tt.body)
			})

			project, response, err := client.Projects.Get(ctx, "org-id", "project-id")

			assert.Nil(t, project)
			require.NotNil(t, response)
			assert.EqualError(t, err, "convert project: response data is missing")
		})
	}
}

func TestProjects_Get_rejectsMalformedResourceWithContext(t *testing.T) {
	setup(t)
	defer teardown()

	mux.HandleFunc("/orgs/org-id/projects/project-id", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = fmt.Fprint(w, `{
			"data": {
				"id": "project-id",
				"type": "target",
				"attributes": {}
			}
		}`)
	})

	project, response, err := client.Projects.Get(ctx, "org-id", "project-id")

	assert.Nil(t, project)
	require.NotNil(t, response)
	assert.EqualError(t, err, `convert project: project "project-id": resource type is "target", expected "project"`)
}

func TestProjects_Delete_success(t *testing.T) {
	setup(t)
	defer teardown()

	mux.HandleFunc("/orgs/org-id/projects/project-id", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		assert.Equal(t, restAPIMediaType, r.Header.Get("Accept"))
		assert.Equal(t, restAPIMediaType, r.Header.Get("Content-Type"))
		assertRequestAPIVersion(t, r, projectsAPIVersion)
		assert.Equal(t, url.Values{"version": {projectsAPIVersion}}, r.URL.Query())
		w.WriteHeader(http.StatusNoContent)
	})

	response, err := client.Projects.Delete(ctx, "org-id", "project-id")

	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Equal(t, http.StatusNoContent, response.StatusCode)
}

func TestProjects_Validation(t *testing.T) {
	c := newTestClient(t)

	tests := map[string]struct {
		call    func(t *testing.T) error
		target  error
		message string
	}{
		"List/missing organization ID": {
			call: func(t *testing.T) error {
				projects, response, err := c.Projects.List(ctx, "", nil)
				assert.Nil(t, projects)
				assert.Nil(t, response)
				return err
			},
			target:  ErrEmptyArgument,
			message: "organization ID: argument is empty",
		},
		"All/missing organization ID": {
			call: func(t *testing.T) error {
				seq, iterErr := c.Projects.All(ctx, "", nil)
				for range seq {
					t.Fatal("unexpected Project")
				}
				return iterErr()
			},
			target:  ErrEmptyArgument,
			message: "organization ID: argument is empty",
		},
		"Get/missing organization ID": {
			call: func(t *testing.T) error {
				project, response, err := c.Projects.Get(ctx, "", "project-id")
				assert.Nil(t, project)
				assert.Nil(t, response)
				return err
			},
			target:  ErrEmptyArgument,
			message: "organization ID: argument is empty",
		},
		"Get/missing project ID": {
			call: func(t *testing.T) error {
				project, response, err := c.Projects.Get(ctx, "org-id", "")
				assert.Nil(t, project)
				assert.Nil(t, response)
				return err
			},
			target:  ErrEmptyArgument,
			message: "project ID: argument is empty",
		},
		"Delete/missing organization ID": {
			call: func(t *testing.T) error {
				response, err := c.Projects.Delete(ctx, "", "project-id")
				assert.Nil(t, response)
				return err
			},
			target:  ErrEmptyArgument,
			message: "organization ID: argument is empty",
		},
		"Delete/missing project ID": {
			call: func(t *testing.T) error {
				response, err := c.Projects.Delete(ctx, "org-id", "")
				assert.Nil(t, response)
				return err
			},
			target:  ErrEmptyArgument,
			message: "project ID: argument is empty",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			err := test.call(t)
			require.ErrorIs(t, err, test.target)
			require.EqualError(t, err, test.message)
		})
	}
}

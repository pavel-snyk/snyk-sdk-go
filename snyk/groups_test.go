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

func TestGroup_MarshalJSON(t *testing.T) {
	group := Group{
		ID:   "group-id",
		Name: "Fictional Group",
	}

	got, err := json.Marshal(group)

	require.NoError(t, err)
	require.JSONEq(t, `{
		"id": "group-id",
		"name": "Fictional Group"
	}`, string(got))
}

func TestGroups_groupFromResource(t *testing.T) {
	resource := groupResource{
		ID:   "group-id",
		Type: "group",
		Attributes: &groupAttributes{
			Name: "Fictional Group",
		},
	}

	group, err := groupFromResource(resource)

	require.NoError(t, err)
	assert.Equal(t, Group{ID: "group-id", Name: "Fictional Group"}, group)
}

func TestGroups_groupFromResource_allowsEmptyName(t *testing.T) {
	group, err := groupFromResource(groupResource{
		ID:         "group-id",
		Type:       "group",
		Attributes: &groupAttributes{},
	})

	require.NoError(t, err)
	assert.Equal(t, Group{ID: "group-id"}, group)
}

func TestGroups_groupFromResource_rejectsStructuralViolations(t *testing.T) {
	tests := []struct {
		name     string
		resource groupResource
		wantErr  string
	}{
		{
			name:     "empty resource ID",
			resource: groupResource{Type: "group", Attributes: &groupAttributes{}},
			wantErr:  "resource ID is empty",
		},
		{
			name:     "wrong resource type",
			resource: groupResource{ID: "group-id", Type: "tenant", Attributes: &groupAttributes{}},
			wantErr:  `group "group-id": resource type is "tenant", expected "group"`,
		},
		{
			name:     "missing attributes",
			resource: groupResource{ID: "group-id", Type: "group"},
			wantErr:  `group "group-id": attributes are missing`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := groupFromResource(tt.resource)

			assert.EqualError(t, err, tt.wantErr)
		})
	}
}

func TestGroups_List_success(t *testing.T) {
	setup(t)
	defer teardown()

	fixture := loadFixture(t, "groups_list_success.json")
	mux.HandleFunc("/groups", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, restAPIMediaType, r.Header.Get("Accept"))
		assert.Equal(t, restAPIMediaType, r.Header.Get("Content-Type"))
		assertRequestAPIVersion(t, r, "2026-03-25")
		assert.Equal(t, url.Values{
			"ending_before": {"ending-cursor"},
			"limit":         {"20"},
			"version":       {"2026-03-25"},
		}, r.URL.Query())
		w.Header().Set(headerSnykVersionServed, "2023-01-30~beta")
		_, _ = w.Write(fixture)
	})

	groups, response, err := client.Groups.List(ctx, &GroupListOptions{
		ListOptions: ListOptions{EndingBefore: "ending-cursor", Limit: 20},
	})

	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Equal(t, "2023-01-30~beta", response.ServedAPIVersion)
	assert.Equal(t, []Group{
		{ID: "11111111-1111-1111-1111-111111111111", Name: "Fictional Platform Group"},
		{ID: "22222222-2222-2222-2222-222222222222", Name: "Example Engineering Group"},
	}, groups)
	require.NotNil(t, response.Links)
	assert.Equal(t, "next-cursor", mustStartingAfter(t, response.Links.Next))
}

func TestGroups_List_nilOptions(t *testing.T) {
	setup(t)
	defer teardown()

	mux.HandleFunc("/groups", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, url.Values{"version": {groupsAPIVersion}}, r.URL.Query())
		_, _ = fmt.Fprint(w, `{"data":[],"links":{}}`)
	})

	groups, _, err := client.Groups.List(ctx, nil)

	require.NoError(t, err)
	assert.NotNil(t, groups)
	assert.Empty(t, groups)
}

func TestGroups_List_rejectsMissingOrNullData(t *testing.T) {
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

			mux.HandleFunc("/groups", func(w http.ResponseWriter, _ *http.Request) {
				_, _ = fmt.Fprint(w, tt.body)
			})

			groups, response, err := client.Groups.List(ctx, nil)

			assert.Nil(t, groups)
			require.NotNil(t, response)
			assert.EqualError(t, err, "convert groups: response data is missing")
		})
	}
}

func TestGroups_List_convertsPageAtomicallyWithIndexContext(t *testing.T) {
	setup(t)
	defer teardown()

	mux.HandleFunc("/groups", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = fmt.Fprint(w, `{
			"data": [
				{"id":"group-1","type":"group","attributes":{"name":"First"}},
				{"id":"group-2","type":"tenant","attributes":{"name":"Second"}}
			]
		}`)
	})

	groups, response, err := client.Groups.List(ctx, nil)

	assert.Nil(t, groups)
	require.NotNil(t, response)
	assert.EqualError(t, err, `convert group at index 1: group "group-2": resource type is "tenant", expected "group"`)
}

func TestGroups_Get_success(t *testing.T) {
	setup(t)
	defer teardown()

	const groupID = "11111111-1111-1111-1111-111111111111"
	fixture := loadFixture(t, "groups_get_success.json")
	mux.HandleFunc("/groups/"+groupID, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, restAPIMediaType, r.Header.Get("Accept"))
		assert.Equal(t, restAPIMediaType, r.Header.Get("Content-Type"))
		assertRequestAPIVersion(t, r, "2026-03-25")
		assert.Equal(t, url.Values{"version": {"2026-03-25"}}, r.URL.Query())
		_, _ = w.Write(fixture)
	})

	group, response, err := client.Groups.Get(ctx, groupID)

	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Equal(t, &Group{ID: groupID, Name: "Fictional Platform Group"}, group)
}

func TestGroups_Get_rejectsMissingOrNullData(t *testing.T) {
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

			mux.HandleFunc("/groups/group-id", func(w http.ResponseWriter, _ *http.Request) {
				_, _ = fmt.Fprint(w, tt.body)
			})

			group, response, err := client.Groups.Get(ctx, "group-id")

			assert.Nil(t, group)
			require.NotNil(t, response)
			assert.EqualError(t, err, "convert group: response data is missing")
		})
	}
}

func TestGroups_Get_rejectsMalformedResourceWithContextAndPreservesResponse(t *testing.T) {
	setup(t)
	defer teardown()

	mux.HandleFunc("/groups/group-id", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set(headerSnykRequestID, "request-id")
		_, _ = fmt.Fprint(w, `{
			"data": {
				"id": "group-id",
				"type": "tenant",
				"attributes": {}
			}
		}`)
	})

	group, response, err := client.Groups.Get(ctx, "group-id")

	assert.Nil(t, group)
	require.NotNil(t, response)
	assert.Equal(t, "request-id", response.SnykRequestID)
	assert.EqualError(t, err, `convert group: group "group-id": resource type is "tenant", expected "group"`)
}

func TestGroups_Get_rejectsEmptyGroupID(t *testing.T) {
	c := newTestClient(t)

	group, response, err := c.Groups.Get(ctx, "")

	assert.Nil(t, group)
	assert.Nil(t, response)
	require.ErrorIs(t, err, ErrEmptyArgument)
	require.EqualError(t, err, "group ID: argument is empty")
}

func TestGroups_All_multiplePages(t *testing.T) {
	setup(t)
	defer teardown()

	page1 := loadFixture(t, "groups_all_page_1.json")
	page2 := loadFixture(t, "groups_all_page_2.json")
	options := &GroupListOptions{
		ListOptions: ListOptions{StartingAfter: "initial-cursor", Limit: 10},
	}
	requestCount := 0
	mux.HandleFunc("/groups", func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		assert.Equal(t, http.MethodGet, r.Method)
		assertRequestAPIVersion(t, r, groupsAPIVersion)
		switch requestCount {
		case 1:
			assert.Equal(t, url.Values{
				"limit":          {"10"},
				"starting_after": {"initial-cursor"},
				"version":        {groupsAPIVersion},
			}, r.URL.Query())
			_, _ = w.Write(page1)
		case 2:
			assert.Equal(t, url.Values{
				"limit":          {"10"},
				"starting_after": {"next-cursor"},
				"version":        {groupsAPIVersion},
			}, r.URL.Query())
			_, _ = w.Write(page2)
		default:
			http.Error(w, "unexpected pagination request", http.StatusInternalServerError)
		}
	})

	seq, iterErr := client.Groups.All(ctx, options)
	var groups []Group
	for group := range seq {
		groups = append(groups, group)
	}

	require.NoError(t, iterErr())
	assert.Equal(t, []Group{
		{ID: "11111111-1111-1111-1111-111111111111", Name: "Fictional Platform Group"},
		{ID: "22222222-2222-2222-2222-222222222222", Name: "Example Engineering Group"},
	}, groups)
	assert.Equal(t, 2, requestCount)
	assert.Equal(t, &GroupListOptions{
		ListOptions: ListOptions{StartingAfter: "initial-cursor", Limit: 10},
	}, options)
}

func TestGroups_All_snapshotsOptionsAtConstruction(t *testing.T) {
	setup(t)
	defer teardown()

	options := &GroupListOptions{
		ListOptions: ListOptions{StartingAfter: "original-cursor", Limit: 10},
	}
	seq, iterErr := client.Groups.All(ctx, options)

	options.StartingAfter = "changed-cursor"
	options.EndingBefore = "changed-ending-cursor"
	options.Limit = 99

	requestCount := 0
	mux.HandleFunc("/groups", func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		assert.Equal(t, url.Values{
			"limit":          {"10"},
			"starting_after": {"original-cursor"},
			"version":        {groupsAPIVersion},
		}, r.URL.Query())
		_, _ = fmt.Fprint(w, `{"data":[],"links":{}}`)
	})

	for range seq {
	}

	require.NoError(t, iterErr())
	assert.Equal(t, 1, requestCount)
}

func TestGroups_All_restartsSequentialIterationFromInitialCursor(t *testing.T) {
	setup(t)
	defer teardown()

	page1 := loadFixture(t, "groups_all_page_1.json")
	page2 := loadFixture(t, "groups_all_page_2.json")
	requestCount := 0
	mux.HandleFunc("/groups", func(w http.ResponseWriter, r *http.Request) {
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

	seq, iterErr := client.Groups.All(ctx, &GroupListOptions{
		ListOptions: ListOptions{StartingAfter: "initial-cursor"},
	})

	var firstIteration []string
	for group := range seq {
		firstIteration = append(firstIteration, group.ID)
	}
	require.NoError(t, iterErr())

	var secondIteration []string
	for group := range seq {
		secondIteration = append(secondIteration, group.ID)
	}
	require.NoError(t, iterErr())

	wantGroupIDs := []string{
		"11111111-1111-1111-1111-111111111111",
		"22222222-2222-2222-2222-222222222222",
	}
	assert.Equal(t, wantGroupIDs, firstIteration)
	assert.Equal(t, wantGroupIDs, secondIteration)
	assert.Equal(t, 4, requestCount)
}

func TestGroups_All_rejectsEndingBefore(t *testing.T) {
	setup(t)
	defer teardown()

	seq, iterErr := client.Groups.All(ctx, &GroupListOptions{
		ListOptions: ListOptions{EndingBefore: "previous-cursor"},
	})
	for range seq {
		t.Fatal("unexpected Group")
	}

	assert.EqualError(t, iterErr(), "ending-before pagination is not supported when iterating all groups")
}

func TestGroups_All_convertsEachPageAtomically(t *testing.T) {
	setup(t)
	defer teardown()

	requestCount := 0
	mux.HandleFunc("/groups", func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		if r.URL.Query().Get("starting_after") == "second-page" {
			_, _ = fmt.Fprint(w, `{
				"data": [
					{"id":"group-3","type":"group","attributes":{"name":"Third"}},
					{"id":"group-4","type":"tenant","attributes":{"name":"Fourth"}}
				],
				"links": {}
			}`)
			return
		}

		_, _ = fmt.Fprint(w, `{
			"data": [
				{"id":"group-1","type":"group","attributes":{"name":"First"}},
				{"id":"group-2","type":"group","attributes":{"name":"Second"}}
			],
			"links": {"next":"/groups?version=server-value&starting_after=second-page"}
		}`)
	})

	seq, iterErr := client.Groups.All(ctx, nil)
	var groupIDs []string
	for group := range seq {
		groupIDs = append(groupIDs, group.ID)
	}

	assert.Equal(t, []string{"group-1", "group-2"}, groupIDs)
	assert.EqualError(t, iterErr(), `convert group at index 1: group "group-4": resource type is "tenant", expected "group"`)
	assert.Equal(t, 2, requestCount)
}

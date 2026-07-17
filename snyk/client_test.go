package snyk

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	client *Client
	ctx    = context.TODO()
	mux    *http.ServeMux
	server *httptest.Server
)

func setup() {
	mux = http.NewServeMux()
	server = httptest.NewServer(mux)

	client, _ = NewClient("auth-token",
		WithRegion(Region{
			Alias:       "TEST",
			AppBaseURL:  fmt.Sprintf("%v/", server.URL),
			RESTBaseURL: fmt.Sprintf("%v/", server.URL),
			V1BaseURL:   fmt.Sprintf("%v/", server.URL),
		}),
	)
}

func teardown() {
	server.Close()
}

func TestClient_Regions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		expected Region
	}{
		{
			name: "SNYK-US-01",
			expected: Region{
				Alias:       "SNYK-US-01",
				AppBaseURL:  "https://app.snyk.io/",
				RESTBaseURL: "https://api.snyk.io/rest/",
				V1BaseURL:   "https://api.snyk.io/v1/",
			},
		},
		{
			name: "SNYK-US-02",
			expected: Region{
				Alias:       "SNYK-US-02",
				AppBaseURL:  "https://app.us.snyk.io/",
				RESTBaseURL: "https://api.us.snyk.io/rest/",
				V1BaseURL:   "https://api.us.snyk.io/v1/",
			},
		},
		{
			name: "SNYK-EU-01",
			expected: Region{
				Alias:       "SNYK-EU-01",
				AppBaseURL:  "https://app.eu.snyk.io/",
				RESTBaseURL: "https://api.eu.snyk.io/rest/",
				V1BaseURL:   "https://api.eu.snyk.io/v1/",
			},
		},
		{
			name: "SNYK-AU-01",
			expected: Region{
				Alias:       "SNYK-AU-01",
				AppBaseURL:  "https://app.au.snyk.io/",
				RESTBaseURL: "https://api.au.snyk.io/rest/",
				V1BaseURL:   "https://api.au.snyk.io/v1/",
			},
		},
	}

	actualRegions := Regions()
	require.Len(t, actualRegions, len(tests))

	for i, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, actualRegions[i])
		})
	}
}

func TestClient_NewClient_defaults(t *testing.T) {
	client, err := NewClient("auth-token")

	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, "https://app.snyk.io/", client.appBaseURL.String())
	assert.Equal(t, "https://api.snyk.io/rest/", client.restBaseURL.String())
	assert.Equal(t, "https://api.snyk.io/v1/", client.v1BaseURL.String())
}

func TestClient_NewClient_withCustomRegion(t *testing.T) {
	expectedAppBaseURL, _ := url.Parse("https://app.testing.snyk.io/")
	expectedRESTBaseURL, _ := url.Parse("https://api.testing.snyk.io/rest")
	expectedV1BaseURL, _ := url.Parse("https://api.testing.snyk.io/v1")
	client, err := NewClient("auth-token", WithRegion(
		Region{
			AppBaseURL:  "https://app.testing.snyk.io/",
			RESTBaseURL: "https://api.testing.snyk.io/rest",
			V1BaseURL:   "https://api.testing.snyk.io/v1",
		}),
	)

	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, expectedAppBaseURL, client.appBaseURL)
	assert.Equal(t, expectedRESTBaseURL, client.restBaseURL)
	assert.Equal(t, expectedV1BaseURL, client.v1BaseURL)
}

func TestClient_NewClient_withHTTPClient(t *testing.T) {
	httpClient := &http.Client{Timeout: 2 * time.Second}
	client, err := NewClient("auth-token",
		WithHTTPClient(httpClient),
	)

	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, 2*time.Second, client.httpClient.Timeout)
}

func TestClient_NewClient_withUserAgent(t *testing.T) {
	client, err := NewClient("auth-token",
		WithUserAgent("test-user-agent"),
	)

	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, "test-user-agent", client.userAgent)
}

func TestClient_restPath(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		endpoint        string
		version         string
		opts            any
		expectedPath    string
		expectedQueries url.Values
		expectedError   string
	}{
		"version-only": {
			endpoint:        "projects",
			version:         "2025-11-05",
			expectedPath:    "projects",
			expectedQueries: url.Values{"version": {"2025-11-05"}},
		},
		"typed-nil-options": {
			endpoint:        "projects",
			version:         "2025-11-05",
			opts:            (*ListOptions)(nil),
			expectedPath:    "projects",
			expectedQueries: url.Values{"version": {"2025-11-05"}},
		},
		"list-options": {
			endpoint:     "projects",
			version:      "2025-11-05",
			opts:         &ListOptions{StartingAfter: "cursor", Limit: 25},
			expectedPath: "projects",
			expectedQueries: url.Values{
				"limit":          {"25"},
				"starting_after": {"cursor"},
				"version":        {"2025-11-05"},
			},
		},
		"rejects-endpoint-query": {
			endpoint:      "projects?version=caller-value",
			version:       "2025-11-05",
			expectedError: "must not contain a query or fragment",
		},
		"rejects-endpoint-fragment": {
			endpoint:      "projects#section",
			version:       "2025-11-05",
			expectedError: "must not contain a query or fragment",
		},
		"empty-version": {
			endpoint:      "projects",
			expectedError: "API version is required",
		},
		"malformed-endpoint": {
			endpoint:      "://projects",
			version:       "2025-11-05",
			expectedError: "missing protocol scheme",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			path, err := restPath(test.endpoint, test.version, test.opts)
			if test.expectedError != "" {
				assert.ErrorContains(t, err, test.expectedError)
				return
			}

			if !assert.NoError(t, err) {
				return
			}
			parsedPath, err := url.Parse(path)
			if !assert.NoError(t, err) {
				return
			}
			assert.Equal(t, test.expectedPath, parsedPath.Path)
			assert.Equal(t, test.expectedQueries, parsedPath.Query())
		})
	}
}

func TestClient_newResponse_populatesServedAPIVersion(t *testing.T) {
	t.Parallel()

	header := make(http.Header)
	header.Set(headerSnykVersionServed, "2023-01-30~beta")
	httpResponse := &http.Response{Header: header}

	response := newResponse(httpResponse)
	errorResponse := &ErrorResponse{Response: response}

	assert.Equal(t, "2023-01-30~beta", response.ServedAPIVersion)
	assert.Equal(t, response.ServedAPIVersion, errorResponse.Response.ServedAPIVersion)
}

func TestClient_newResponse_withoutServedAPIVersion(t *testing.T) {
	t.Parallel()

	response := newResponse(&http.Response{Header: make(http.Header)})

	assert.Empty(t, response.ServedAPIVersion)
}

func assertRequestAPIVersion(t *testing.T, r *http.Request, expected string) {
	t.Helper()
	assert.Equal(t, []string{expected}, r.URL.Query()["version"])
}

func loadFixture(t *testing.T, fixtureName string) []byte {
	t.Helper()

	data, err := os.ReadFile(filepath.Join("testdata", fixtureName))
	if err != nil {
		t.Fatalf("failed to load fixture %q: %v", fixtureName, err)
	}

	return data
}

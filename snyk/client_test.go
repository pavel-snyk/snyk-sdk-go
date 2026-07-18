package snyk

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
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

func newTestClient(t testing.TB, options ...ClientOption) *Client {
	t.Helper()

	c, err := NewClient("auth-token", options...)
	require.NoError(t, err)

	return c
}

func setup(t testing.TB) {
	t.Helper()

	mux = http.NewServeMux()
	server = httptest.NewServer(mux)

	client = newTestClient(t,
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

func TestClient_checkResponse(t *testing.T) {
	tests := map[string]struct {
		statusCode      int
		body            string
		wantAPIErrors   []APIError
		wantErrContains string
	}{
		"success": {
			statusCode: http.StatusOK,
		},
		"REST error": {
			statusCode: http.StatusNotFound,
			body:       `{"errors":[{"status":"404","title":"Project not found"}]}`,
			wantAPIErrors: []APIError{{
				StatusCode: "404",
				Title:      "Project not found",
			}},
		},
		"legacy error": {
			statusCode: http.StatusBadRequest,
			body:       `{"message":"Invalid request","errorRef":"fake-error-reference"}`,
			wantAPIErrors: []APIError{{
				Detail:     "Invalid request",
				ID:         "fake-error-reference",
				StatusCode: "400",
				Title:      "Invalid request",
			}},
		},
		"undecodable error": {
			statusCode:      http.StatusInternalServerError,
			body:            "not JSON",
			wantErrContains: "failed to decode Snyk API error response; status: 500",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			response := createTestResponse(http.MethodGet, "https://api.snyk.io/rest/projects", test.statusCode, "")
			response.Body = io.NopCloser(strings.NewReader(test.body))

			err := checkResponse(response)

			if test.statusCode == http.StatusOK {
				require.NoError(t, err)
				return
			}
			if test.wantErrContains != "" {
				assert.ErrorContains(t, err, test.wantErrContains)
				return
			}

			var errorResponse *ErrorResponse
			require.ErrorAs(t, err, &errorResponse)
			assert.Same(t, response, errorResponse.Response)
			assert.Equal(t, test.wantAPIErrors, errorResponse.APIErrors)
		})
	}
}

func TestClient_do_decodesSuccessResponse(t *testing.T) {
	c := newTestClient(t, WithHTTPClient(&http.Client{Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(`{"name":"decoded"}`)),
			Request:    req,
		}, nil
	})}))
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "https://api.snyk.io/rest/projects", nil)
	require.NoError(t, err)
	destination := struct {
		Name string `json:"name"`
	}{Name: "original"}

	response, err := c.do(context.Background(), req, &destination)

	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Equal(t, http.StatusOK, response.StatusCode)
	assert.Equal(t, "decoded", destination.Name)
}

func TestClient_do_returnsErrorResponse(t *testing.T) {
	c := newTestClient(t, WithHTTPClient(&http.Client{Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		header := make(http.Header)
		header.Set(headerSnykRequestID, "fake-request-id")
		header.Set(headerSnykVersionServed, "2025-11-05")
		return &http.Response{
			StatusCode: http.StatusNotFound,
			Header:     header,
			Body:       io.NopCloser(strings.NewReader(`{"errors":[{"status":"404","title":"Project not found"}]}`)),
			Request:    req,
		}, nil
	})}))
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "https://api.snyk.io/rest/projects/missing", nil)
	require.NoError(t, err)
	destination := struct {
		Name string `json:"name"`
	}{Name: "unchanged"}
	wantDestination := destination

	response, err := c.do(context.Background(), req, &destination)

	require.NotNil(t, response)
	var errorResponse *ErrorResponse
	require.ErrorAs(t, err, &errorResponse)
	assert.Same(t, response, errorResponse.Response)
	assert.Equal(t, http.StatusNotFound, response.StatusCode)
	assert.Equal(t, "fake-request-id", response.SnykRequestID)
	assert.Equal(t, "2025-11-05", response.ServedAPIVersion)
	assert.Equal(t, wantDestination, destination)
}

func TestClient_do_propagatesNetworkError(t *testing.T) {
	networkErr := errors.New("network unavailable")
	c := newTestClient(t, WithHTTPClient(&http.Client{Transport: roundTripperFunc(func(*http.Request) (*http.Response, error) {
		return nil, networkErr
	})}))
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "https://api.snyk.io/rest/projects", nil)
	require.NoError(t, err)

	response, err := c.do(context.Background(), req, nil)

	assert.Nil(t, response)
	require.ErrorIs(t, err, networkErr)
}

func TestClient_do_prefersCanceledContextError(t *testing.T) {
	c := newTestClient(t, WithHTTPClient(&http.Client{Transport: roundTripperFunc(func(*http.Request) (*http.Response, error) {
		return nil, errors.New("transport error")
	})}))
	canceledCtx, cancel := context.WithCancel(context.Background())
	cancel()
	req, err := http.NewRequestWithContext(canceledCtx, http.MethodGet, "https://api.snyk.io/rest/projects", nil)
	require.NoError(t, err)

	response, err := c.do(canceledCtx, req, nil)

	assert.Nil(t, response)
	require.ErrorIs(t, err, context.Canceled)
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

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

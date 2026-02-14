package snyk

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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

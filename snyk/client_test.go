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

	client = NewClient("auth-token",
		WithBaseURL(fmt.Sprintf("%v/", server.URL)),
	)
}

func teardown() {
	server.Close()
}

func TestClient_NewClient_defaults(t *testing.T) {
	setup()
	defer teardown()

	assert.NotNil(t, client.baseURL)
	assert.Equal(t, fmt.Sprintf("%v/", server.URL), client.baseURL.String())
	assert.Equal(t, "auth-token", client.token)
}

func TestClient_NewClient_withBaseURL(t *testing.T) {
	expectedBaseURL, _ := url.Parse("https://testing.snyk.io/api")
	client := NewClient("auth-token",
		WithBaseURL("https://testing.snyk.io/api"),
	)

	assert.Equal(t, expectedBaseURL, client.baseURL)
}

func TestClient_NewClient_withHTTPClient(t *testing.T) {
	httpClient := &http.Client{Timeout: 2 * time.Second}
	client := NewClient("auth-token",
		WithHTTPClient(httpClient),
	)

	assert.Equal(t, 2*time.Second, client.httpClient.Timeout)
}

func TestClient_NewClient_withUserAgent(t *testing.T) {
	client := NewClient("auth-token",
		WithUserAgent("test-user-agent"),
	)

	assert.Equal(t, "test-user-agent", client.userAgent)
}

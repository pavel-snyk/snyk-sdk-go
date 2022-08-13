package snyk

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

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

func TestClient_NewClient(t *testing.T) {
	setup()
	defer teardown()

	assert.NotNil(t, client.BaseURL)
	assert.Equal(t, fmt.Sprintf("%v/", server.URL), client.BaseURL.String())
	assert.Equal(t, "auth-token", client.Token)
}

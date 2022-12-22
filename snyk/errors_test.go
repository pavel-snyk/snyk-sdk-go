package snyk

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrorResponse_Error_withoutSnykRequestID(t *testing.T) {
	errorResponse := &ErrorResponse{
		Response: &Response{
			Response: &http.Response{
				Request: &http.Request{
					Method: http.MethodGet,
					URL:    &url.URL{Scheme: "https", Path: "snyk.io/api"},
				},
				StatusCode: 401,
			},
		},
		ErrorElement: ErrorElement{
			Code:    401,
			Message: "Invalid auth token provided",
		},
	}

	expectedMessage := "GET https://snyk.io/api: 401 {Code:401 Message:Invalid auth token provided}"

	assert.Error(t, errorResponse)
	assert.Equal(t, expectedMessage, errorResponse.Error())
}

func TestErrorResponse_Error_withSnykRequestID(t *testing.T) {
	errorResponse := &ErrorResponse{
		Response: &Response{
			Response: &http.Response{
				Request: &http.Request{
					Method: http.MethodGet,
					URL:    &url.URL{Scheme: "https", Path: "snyk.io/api"},
				},
				StatusCode: 401,
			},
			SnykRequestID: "some-uuid-for-snyk-request-id",
		},
	}

	expectedMessage := "GET https://snyk.io/api: 401 (snyk-request-id: some-uuid-for-snyk-request-id) {Code:0 Message:}"

	assert.Error(t, errorResponse)
	assert.Equal(t, expectedMessage, errorResponse.Error())
}

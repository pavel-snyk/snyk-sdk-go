package snyk

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestError_parseRESTError(t *testing.T) {
	tests := map[string]struct {
		data      string
		want      []APIError
		wantValid bool
	}{
		"single error": {
			data: `{"errors":[{
				"status":"404",
				"title":"Project not found",
				"detail":"The requested Project does not exist"
			}]}`,
			want: []APIError{{
				StatusCode: "404",
				Title:      "Project not found",
				Detail:     "The requested Project does not exist",
			}},
			wantValid: true,
		},
		"multiple errors": {
			data: `{"errors":[
				{"status":"400","title":"Invalid request"},
				{"status":"403","detail":"Permission denied"}
			]}`,
			want: []APIError{
				{StatusCode: "400", Title: "Invalid request"},
				{StatusCode: "403", Detail: "Permission denied"},
			},
			wantValid: true,
		},
		"malformed JSON": {
			data: `{"errors":[`,
		},
		"unexpected shape": {
			data: `{"message":"Not found"}`,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			apiErrors, valid := parseRESTError([]byte(test.data))

			assert.Equal(t, test.wantValid, valid)
			assert.Equal(t, test.want, apiErrors)
		})
	}
}

func TestErrorResponse_Error_withoutSnykRequestID(t *testing.T) {
	errorResponse := &ErrorResponse{
		Response: createTestResponse(http.MethodGet, "https://api.snyk.io/rest/somepath", 401, ""),
		APIErrors: []APIError{{
			Detail:     "Invalid auth token provided",
			StatusCode: "401",
		}},
	}

	var expectedError *ErrorResponse
	expectedMessage := "GET https://api.snyk.io/rest/somepath: 401 Invalid auth token provided"

	assert.ErrorAs(t, errorResponse, &expectedError)
	assert.Len(t, errorResponse.APIErrors, 1)
	assert.Equal(t, expectedMessage, errorResponse.Error())
}

func TestErrorResponse_Error_withSnykRequestID(t *testing.T) {
	errorResponse := &ErrorResponse{
		Response: createTestResponse(http.MethodGet, "https://api.snyk.io/rest/somepath", 401, "some-uuid-for-snyk-request-id"),
		APIErrors: []APIError{{
			Detail:     "Invalid auth token provided",
			StatusCode: "401",
		}},
	}

	var expectedError *ErrorResponse
	expectedMessage := "GET https://api.snyk.io/rest/somepath: 401 (snyk-request-id: some-uuid-for-snyk-request-id) Invalid auth token provided"

	assert.ErrorAs(t, errorResponse, &expectedError)
	assert.Len(t, errorResponse.APIErrors, 1)
	assert.Equal(t, expectedMessage, errorResponse.Error())
}

func TestErrorResponse_Error_withSingleErrorElement(t *testing.T) {
	errorResponse := &ErrorResponse{
		Response: createTestResponse(http.MethodGet, "https://api.snyk.io/rest/somepath", 400, "some-uuid-for-snyk-request-id"),
		APIErrors: []APIError{{
			Detail:     "parameter \"expand\" in query has an error: Error at \"/0\": value is not one of the allowed values [\"app\"]\nSchema:\n  {\n    \"enum\": [\n      \"app\"\n",
			StatusCode: "400",
			Title:      "Client request did not conform to OpenAPI specification",
		}},
	}
	var expectedError *ErrorResponse
	expectedMessage := "GET https://api.snyk.io/rest/somepath: 400 (snyk-request-id: some-uuid-for-snyk-request-id) Client request did not conform to OpenAPI specification"

	assert.ErrorAs(t, errorResponse, &expectedError)
	assert.Len(t, errorResponse.APIErrors, 1)
	assert.Equal(t, expectedMessage, errorResponse.Error())
}

func TestErrorResponse_Error_withMultipleErrorElements(t *testing.T) {
	errorResponse := &ErrorResponse{
		Response: createTestResponse(http.MethodGet, "https://api.snyk.io/rest/somepath", 400, "some-uuid-for-snyk-request-id"),
		APIErrors: []APIError{
			{
				Detail:     "parameter \"expand\" in query has an error",
				StatusCode: "400",
				Title:      "Client request did not conform to OpenAPI specification",
			},
			{
				Detail:     "Permission denied for this resource",
				StatusCode: "403",
			},
		},
	}
	var expectedError *ErrorResponse
	expectedMessage := "GET https://api.snyk.io/rest/somepath: 400 (snyk-request-id: some-uuid-for-snyk-request-id) Client request did not conform to OpenAPI specification, Permission denied for this resource"

	assert.ErrorAs(t, errorResponse, &expectedError)
	assert.Len(t, errorResponse.APIErrors, 2)
	assert.Equal(t, expectedMessage, errorResponse.Error())
}

func createTestResponse(method, urlStr string, statusCode int, requestID string) *Response {
	u, _ := url.Parse(urlStr)
	r := &Response{
		Response: &http.Response{
			Request: &http.Request{
				Method: method,
				URL:    u,
			},
			StatusCode: statusCode,
		},
	}
	if requestID != "" {
		r.SnykRequestID = requestID
	}
	return r
}

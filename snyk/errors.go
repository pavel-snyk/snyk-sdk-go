package snyk

import "fmt"

// An ErrorResponse reports an error caused by an API request.
type ErrorResponse struct {
	Response     *Response
	ErrorElement ErrorElement
}

// ErrorElement represents a single error caused by an API request.
type ErrorElement struct {
	Code    int    `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

func (r *ErrorResponse) Error() string {
	if r.Response.SnykRequestID != "" {
		return fmt.Sprintf("%v %v: %d (snyk-request-id: %v) %+v",
			r.Response.Request.Method, r.Response.Request.URL, r.Response.StatusCode, r.Response.SnykRequestID, r.ErrorElement,
		)
	}
	return fmt.Sprintf("%v %v: %d %+v",
		r.Response.Request.Method, r.Response.Request.URL, r.Response.StatusCode, r.ErrorElement,
	)
}

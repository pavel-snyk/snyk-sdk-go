package snyk

import (
	"errors"
	"fmt"
)

var (
	// ErrEmptyArgument indicates that mandatory argument is empty.
	ErrEmptyArgument = errors.New("snyk-sdk-go: argument cannot be empty")
	// ErrEmptyPayloadNotAllowed indicates that empty payload is not allowed.
	ErrEmptyPayloadNotAllowed = errors.New("snyk-sdk-go: empty payload is not allowed")
)

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

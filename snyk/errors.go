package snyk

import (
	"encoding/json"
	"fmt"
	"strings"
)

// An ErrorResponse reports an error caused by an API request.
type ErrorResponse struct {
	Response *Response

	APIErrors []APIError
}

// APIError represents a single error caused by an API request.
type APIError struct {
	// A human-readable explanation specific to this occurrence of the problem.
	Detail string `json:"detail"`

	// A unique identifier for this particular occurrence of the problem.
	ID string `json:"id,omitempty"`

	//  HTTP status code applicable to this problem, expressed as a string value.
	StatusCode string `json:"status"`

	// A short, human-readable summary of the problem.
	Title string `json:"title,omitempty"`
}

func (e APIError) String() string { return Stringify(e) }

// UnmarshalJSON is a custom unmarshaller for APIError to handle inconsistent "detail" vs "details" fields.
//
//goland:noinspection GoMixedReceiverTypes (https://youtrack.jetbrains.com/issue/GO-13587)
func (e *APIError) UnmarshalJSON(data []byte) error {
	type tempAPIError struct {
		Detail     string `json:"detail"`
		Details    string `json:"details"`
		ID         string `json:"id"`
		StatusCode string `json:"status"`
		Title      string `json:"title"`
	}
	var temp tempAPIError
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	// coalesce "details" into "detail"
	e.Detail = temp.Detail
	if e.Detail == "" {
		e.Detail = temp.Details
	}
	e.ID = temp.ID
	e.StatusCode = temp.StatusCode
	e.Title = temp.Title

	return nil
}

func (r *ErrorResponse) Error() string {
	errorMessages := make([]string, 0, len(r.APIErrors))
	for _, apiError := range r.APIErrors {
		// prioritize the Title for summary if exists
		if apiError.Title != "" {
			errorMessages = append(errorMessages, apiError.Title)
			continue
		}

		// if not title, fall back to the first line of the Detail
		firstLine, _, _ := strings.Cut(apiError.Detail, "\n")
		errorMessages = append(errorMessages, firstLine)
	}

	if r.Response.SnykRequestID != "" {
		return fmt.Sprintf("%v %v: %d (snyk-request-id: %v) %s",
			r.Response.Request.Method, r.Response.Request.URL, r.Response.StatusCode, r.Response.SnykRequestID, strings.Join(errorMessages, ", "),
		)
	}
	return fmt.Sprintf("%v %v: %d %s",
		r.Response.Request.Method, r.Response.Request.URL, r.Response.StatusCode, strings.Join(errorMessages, ", "),
	)
}

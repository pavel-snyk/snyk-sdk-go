package snyk

import (
	"context"
	"net/http"
)

const (
	usersAPIVersion = "2025-11-05"
)

// UsersService handles communication with the user related methods of the Snyk API.
type UsersService service

// User represents a Snyk user.
//
// See: https://docs.snyk.io/snyk-platform-administration/groups-and-organizations#snyk-features-for-user-management
type User struct {
	ID         string          `json:"id"`                   // The User identifier.
	Type       string          `json:"type"`                 // The resource type `user`.
	Attributes *UserAttributes `json:"attributes,omitempty"` // The User resource data.
}

type UserAttributes struct {
	DefaultOrgID string `json:"default_org_context,omitempty"` // The ID of the default Organization for the User.
	Email        string `json:"email,omitempty"`               // The email of the User.
	Name         string `json:"name"`                          // The name of the User.
	Username     string `json:"username,omitempty"`            // The username of the User.
}

type userRoot struct {
	User *User `json:"data,omitempty"`
}

func (u User) String() string { return Stringify(u) }

// GetSelf provides the details about the user making the request.
//
// See: https://docs.snyk.io/snyk-api/reference/users#get-self
func (s *UsersService) GetSelf(ctx context.Context) (*User, *Response, error) {
	path, err := restPath("self", usersAPIVersion, nil)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.prepareRequest(ctx, http.MethodGet, s.client.restBaseURL, path, nil)
	if err != nil {
		return nil, nil, err
	}

	root := new(userRoot)
	resp, err := s.client.do(ctx, req, &root)
	if err != nil {
		return nil, resp, err
	}

	return root.User, resp, nil
}

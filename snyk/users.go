package snyk

import (
	"context"
	"fmt"
	"net/http"
)

const userBasePath = "user"

// UsersService handles communication with the user related methods of the Snyk API.
type UsersService service

// User represents a Snyk user.
type User struct {
	Email         string         `json:"email,omitempty"`
	ID            string         `json:"id,omitempty"`
	Name          string         `json:"name,omitempty"`
	Username      string         `json:"username,omitempty"`
	Organizations []Organization `json:"orgs,omitempty"`
}

// GetCurrent retrieves information about the user making the request.
//
// Note: the retrieved user will include information about organizations
// that the user belongs to.
func (s *UsersService) GetCurrent(ctx context.Context) (*User, *Response, error) {
	path := fmt.Sprintf("%s/me", userBasePath)

	req, err := s.client.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}

	user := new(User)
	resp, err := s.client.Do(ctx, req, user)
	if err != nil {
		return nil, resp, err
	}

	return user, resp, nil
}

// Get retrieves information about a user identified by id.
func (s *UsersService) Get(ctx context.Context, userID string) (*User, *Response, error) {
	path := fmt.Sprintf("%s/%s", userBasePath, userID)

	req, err := s.client.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}

	user := new(User)
	resp, err := s.client.Do(ctx, req, user)
	if err != nil {
		return nil, resp, err
	}

	return user, resp, nil
}

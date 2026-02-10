package snyk

import (
	"context"
	"errors"
	"fmt"
	"net/http"
)

const orgV1BasePath = "org"

// OrgsServiceV1API is an interface for interacting with the orgs endpoints of the Snyk V1 API.
//
// Note: Snyk V1 API endpoints are being gradually deprecated. It is recommended
// to use the REST API via OrgsServiceAPI where possible.
//
// See: https://docs.snyk.io/snyk-api/reference/organizations-v1
type OrgsServiceV1API interface {
	// Create makes a new organization with given payload.
	//
	// See: https://docs.snyk.io/snyk-api/reference/organizations-v1#post-org
	Create(ctx context.Context, createRequest *OrganizationV1CreateRequest) (*OrganizationV1, *Response, error)

	// Delete removes an organization identified by id.
	//
	// See: https://docs.snyk.io/snyk-api/reference/organizations-v1#delete-org-orgid
	Delete(ctx context.Context, orgID string) (*Response, error)
}

// OrgsServiceV1 handles communication with the org related methods of the Snyk V1 API.
type OrgsServiceV1 service

var _ OrgsServiceV1API = &OrgsServiceV1{}

// OrganizationV1 represents a Snyk organization for V1 API.
type OrganizationV1 struct {
	Group *OrganizationV1Group `json:"group,omitempty"`
	ID    string               `json:"id,omitempty"`
	Name  string               `json:"name,omitempty"`
	Slug  string               `json:"slug,omitempty"`
	URL   string               `json:"url,omitempty"`
}

// OrganizationV1Group represents a Snyk group for V1 API.
type OrganizationV1Group struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

type OrganizationV1CreateRequest struct {
	Name        string `json:"name,omitempty"`
	GroupID     string `json:"groupId,omitempty"`
	SourceOrgID string `json:"sourceOrgId,omitempty"` // id of the organization to copy settings from.
}

func (o OrganizationV1) String() string { return Stringify(o) }

func (s *OrgsServiceV1) Create(ctx context.Context, createRequest *OrganizationV1CreateRequest) (*OrganizationV1, *Response, error) {
	if createRequest == nil {
		return nil, nil, errors.New("failed to create organization: payload must be supplied")
	}

	req, err := s.client.prepareRequest(ctx, http.MethodPost, s.client.v1BaseURL, orgV1BasePath, createRequest)
	if err != nil {
		return nil, nil, err
	}

	orgV1 := new(OrganizationV1)
	resp, err := s.client.do(ctx, req, orgV1)
	if err != nil {
		return nil, resp, err
	}

	return orgV1, resp, nil
}

func (s *OrgsServiceV1) Delete(ctx context.Context, orgID string) (*Response, error) {
	if orgID == "" {
		return nil, errors.New("failed to delete organization: id must be supplied")
	}

	path := fmt.Sprintf("%v/%v", orgV1BasePath, orgID)

	req, err := s.client.prepareRequest(ctx, http.MethodDelete, s.client.v1BaseURL, path, nil)
	if err != nil {
		return nil, err
	}

	return s.client.do(ctx, req, nil)
}

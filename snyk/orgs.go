package snyk

import (
	"context"
	"errors"
	"fmt"
	"iter"
	"net/http"
	"time"
)

const (
	orgsBasePath   = "orgs"
	orgsAPIVersion = "2024-10-15"
)

// OrgsService handles communication with the org related methods of the Snyk API.
type OrgsService service

// Organization represents a Snyk organization.
//
// See: https://docs.snyk.io/discover-snyk/getting-started/glossary#organization
type Organization struct {
	ID            string                     `json:"id"`                      // The Organization identifier.
	Type          string                     `json:"type"`                    // The resource type `org`.
	Attributes    *OrganizationAttributes    `json:"attributes,omitempty"`    // The Organization resource data.
	Relationships *OrganizationRelationships `json:"relationships,omitempty"` // The relationships object describing relationships between Organization and Tenant.
}

type OrganizationAttributes struct {
	CreatedAt  time.Time `json:"created_at,omitempty"` // The time the Organization was created.
	GroupID    string    `json:"group_id,omitempty"`   // The ID of the group to which the Organization belongs.
	IsPersonal bool      `json:"is_personal"`          // Whether the Organization is independent (that is, not part of a group).
	Name       string    `json:"name"`                 // The display name of the Organization.
	Slug       string    `json:"slug"`                 // The canonical (unique and URL-friendly) name of the Organization.
	UpdatedAt  time.Time `json:"updated_at,omitempty"` // The time the Organization was last modified.
}

type OrganizationRelationships struct {
	Tenant *tenantRoot `json:"tenant,omitempty"`
}

type ListOrganizationOptions struct {
	ListOptions
	GroupID string `url:"group_id,omitempty"` // If set, only return organizations within the specified group.
	Expand  string `url:"expand,omitempty"`
}

type GetOrganizationOptions struct {
	ListOptions
	Expand string `url:"expand,omitempty"`
}

type OrganizationUpdateRequest struct {
	Name string
}

type orgRoot struct {
	Organization *Organization `json:"data,omitempty"`
}

type orgsRoot struct {
	Organizations []Organization  `json:"data"`
	Links         *PaginatedLinks `json:"links,omitempty"`
}

func (o Organization) String() string { return Stringify(o) }

// ListAccessibleOrgs gets a paginated list of organizations you have access to. If ListOrganizationOptions is nil,
// then relationship for MemberRole will be always expanded.
//
// See: https://docs.snyk.io/snyk-api/reference/orgs#get-orgs
func (s *OrgsService) ListAccessibleOrgs(ctx context.Context, opts *ListOrganizationOptions) ([]Organization, *Response, error) {
	if opts == nil {
		opts = &ListOrganizationOptions{}
	}

	path, err := restPath(orgsBasePath, orgsAPIVersion, opts)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.prepareRequest(ctx, http.MethodGet, s.client.restBaseURL, path, nil)
	if err != nil {
		return nil, nil, err
	}

	root := new(orgsRoot)
	resp, err := s.client.do(ctx, req, &root)
	if err != nil {
		return nil, resp, err
	}
	if l := root.Links; l != nil {
		resp.Links = l
	}

	return root.Organizations, resp, nil
}

// AllAccessibleOrgs returns an iterator over all organizations you have access to.
//
// This method handles the pagination logic internally for each page.
//
// Note: This function is experimental and its signature may change in a future release.
//
// See: https://docs.snyk.io/snyk-api/reference/orgs#get-orgs
func (s *OrgsService) AllAccessibleOrgs(ctx context.Context, opts *ListOptions) (iter.Seq2[Organization, *Response], func() error) {
	if opts == nil {
		opts = &ListOptions{}
	}
	return newPaginator[Organization](ctx, s.client, s.client.restBaseURL, orgsBasePath, orgsAPIVersion, opts)
}

// Get provides the full details of an organization.
//
// See: https://docs.snyk.io/snyk-api/reference/orgs#get-orgs-org_id
func (s *OrgsService) Get(ctx context.Context, orgID string, opts *GetOrganizationOptions) (*Organization, *Response, error) {
	if orgID == "" {
		return nil, nil, errors.New("failed to get org: id must be supplied")
	}

	if opts == nil {
		opts = &GetOrganizationOptions{Expand: "tenant"}
	}

	path, err := restPath(fmt.Sprintf("%v/%v", orgsBasePath, orgID), orgsAPIVersion, opts)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.prepareRequest(ctx, http.MethodGet, s.client.restBaseURL, path, nil)
	if err != nil {
		return nil, nil, err
	}

	root := new(orgRoot)
	resp, err := s.client.do(ctx, req, &root)
	if err != nil {
		return nil, resp, err
	}

	return root.Organization, resp, nil
}

// Update changes the details of an organization.
//
// See: https://docs.snyk.io/snyk-api/reference/orgs#patch-orgs-org_id
func (s *OrgsService) Update(ctx context.Context, orgID string, updateRequest *OrganizationUpdateRequest) (*Organization, *Response, error) {
	if orgID == "" {
		return nil, nil, errors.New("failed to update org: id must be supplied")
	}
	if updateRequest == nil {
		return nil, nil, errors.New("failed to update org: payload must be supplied")
	}

	path, err := restPath(fmt.Sprintf("%v/%v", orgsBasePath, orgID), orgsAPIVersion, nil)
	if err != nil {
		return nil, nil, err
	}

	// inline jsonapi create payload to keep function simple
	var updateRequestJSON struct {
		Data struct {
			Attributes struct {
				Name string `json:"name"`
			} `json:"attributes"`
			ID   string `json:"id"`
			Type string `json:"type"`
		} `json:"data"`
	}
	updateRequestJSON.Data.Attributes.Name = updateRequest.Name
	updateRequestJSON.Data.ID = orgID
	updateRequestJSON.Data.Type = "org"

	req, err := s.client.prepareRequest(ctx, http.MethodPatch, s.client.restBaseURL, path, updateRequestJSON)
	if err != nil {
		return nil, nil, err
	}

	root := new(orgRoot)
	resp, err := s.client.do(ctx, req, &root)
	if err != nil {
		return nil, resp, err
	}

	return root.Organization, resp, nil
}

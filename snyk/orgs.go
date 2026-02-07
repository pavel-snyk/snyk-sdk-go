package snyk

import (
	"context"
	"errors"
	"fmt"
	"iter"
	"net/http"
)

const (
	orgsBasePath   = "orgs"
	orgsAPIVersion = "2024-10-15"
)

// OrgsServiceAPI is an interface for interacting with the orgs endpoints of the Snyk API.
//
// See: https://docs.snyk.io/snyk-api/reference/orgs
type OrgsServiceAPI interface {
	// ListAccessibleOrgs get a paginated list of organizations you have access to. If ListOrganizationOptions is nil,
	// then relationship for MemberRole will be always expanded.
	//
	// See: https://docs.snyk.io/snyk-api/reference/orgs#get-orgs
	ListAccessibleOrgs(ctx context.Context, opts *ListOrganizationOptions) ([]Organization, *Response, error)

	// AllAccessibleOrgs returns an iterator to paginate over all organizations you have access to.
	//
	// This method handles the pagination logic internally by calling ListAccessibleOrgs for each page.
	// The return iterated can be used in a for...range loop to easily process all organizations.
	//
	// Note: This function is experimental and its signature may change in a future release.
	AllAccessibleOrgs(ctx context.Context, opts *ListOptions) (iter.Seq2[Organization, *Response], func() error)

	// Get provides the full details of an organization.
	//
	// See: https://docs.snyk.io/snyk-api/reference/orgs#get-orgs-org_id
	Get(ctx context.Context, orgID string) (*Organization, *Response, error)
}

// OrgsService handles communication with the org related methods of the Snyk API.
type OrgsService service

var _ OrgsServiceAPI = &OrgsService{}

// Organization represents a Snyk organization.
//
// See: https://docs.snyk.io/discover-snyk/getting-started/glossary#organization
type Organization struct {
	ID         string                  `json:"id"`                   // The Organization identifier.
	Type       string                  `json:"type"`                 // The resource type `org`.
	Attributes *OrganizationAttributes `json:"attributes,omitempty"` // The Organization resource data.
}

type OrganizationAttributes struct {
	GroupID    string `json:"group_id,omitempty"` // The ID of the group to which the organization belongs.
	IsPersonal bool   `json:"is_personal"`        // Whether the organization is independent (that is, not part of a group).
	Name       string `json:"name"`               // The display name of the organization.
	Slug       string `json:"slug"`               // The canonical (unique and URL-friendly) name of the organization.
}

type ListOrganizationOptions struct {
	ListOptions
	GroupID string `url:"group_id,omitempty"` // If set, only return organizations within the specified group.
	Expand  string `url:"expand,omitempty"`
}

type orgRoot struct {
	Organization *Organization `json:"data,omitempty"`
}

type orgsRoot struct {
	Organizations []Organization  `json:"data"`
	Links         *PaginatedLinks `json:"links,omitempty"`
}

func (o Organization) String() string { return Stringify(o) }

func (s *OrgsService) ListAccessibleOrgs(ctx context.Context, opts *ListOrganizationOptions) ([]Organization, *Response, error) {
	if opts == nil {
		opts = &ListOrganizationOptions{}
	}
	if opts.Version == "" {
		opts.Version = orgsAPIVersion
	}

	path, err := addOptions(orgsBasePath, opts)
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

func (s *OrgsService) AllAccessibleOrgs(ctx context.Context, opts *ListOptions) (iter.Seq2[Organization, *Response], func() error) {
	if opts == nil {
		opts = &ListOptions{}
	}
	if opts.Version == "" {
		opts.Version = orgsAPIVersion
	}
	return newPaginator[Organization](ctx, s.client, s.client.restBaseURL, orgsBasePath, opts)
}

func (s *OrgsService) Get(ctx context.Context, orgID string) (*Organization, *Response, error) {
	if orgID == "" {
		return nil, nil, errors.New("failed to get org: id must be supplied")
	}

	opts := &BaseOptions{Version: orgsAPIVersion}
	path, err := addOptions(fmt.Sprintf("%v/%v", orgsBasePath, orgID), opts)
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

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
	groupsBasePath   = "groups"
	groupsAPIVersion = "2025-11-05"
)

// GroupsServiceAPI is an interface for interacting with the groups endpoints of the Snyk API.
//
// See: https://docs.snyk.io/snyk-api/reference/groups
type GroupsServiceAPI interface {
	// List gets a paginated list of all groups you are a member of.
	//
	// Note: Group attributes will contain only name. If you want to access full details
	// of a group, use Get method.
	//
	// See: https://docs.snyk.io/snyk-api/reference/groups#get-groups
	List(ctx context.Context, opts *ListOptions) ([]Group, *Response, error)

	// All returns an iterator to paginate over all groups you are a member of.
	//
	// This method handles the pagination logic internally by calling List for each page.
	// The return iterated can be used in a for...range loop to easily process all groups.
	//
	// Note: This function is experimental and its signature may change in a future release.
	All(ctx context.Context, opts *ListOptions) (iter.Seq2[Group, *Response], func() error)

	// Get provides the full details of an group.
	//
	// See: https://docs.snyk.io/snyk-api/reference/group#get-groups-group_id
	Get(ctx context.Context, groupID string) (*Group, *Response, error)
}

// GroupsService handles communication with the group related methods of the Snyk API.
type GroupsService service

var _ GroupsServiceAPI = (*GroupsService)(nil)

// Group represents a Snyk group.
//
// See: https://docs.snyk.io/snyk-platform-administration/groups-and-organizations/groups
type Group struct {
	ID            string              `json:"id"`                      // The Group identifier.
	Type          string              `json:"type"`                    // The resource type `group`.
	Attributes    *GroupAttributes    `json:"attributes,omitempty"`    // The Group resource data.
	Relationships *GroupRelationships `json:"relationships,omitempty"` // The relationships object describing relationships between Group and Tenant.
}

type GroupAttributes struct {
	CreatedAt time.Time `json:"created_at,omitempty"` // The time the Group was created.
	Name      string    `json:"name"`                 // The display name of the Group.
	Slug      string    `json:"slug,omitempty"`       // The canonical (unique and URL-friendly) name of the Group.
	UpdatedAt time.Time `json:"updated_at,omitempty"` // The time the Group was last modified.
}

type GroupRelationships struct {
	Tenant *tenantRoot `json:"tenant,omitempty"`
}

type ListGroupsOptions struct {
	ListOptions
}

type groupRoot struct {
	Group *Group `json:"data,omitempty"`
}

type groupsRoot struct {
	Groups []Group         `json:"data"`
	Links  *PaginatedLinks `json:"links,omitempty"`
}

func (g Group) String() string { return Stringify(g) }

func (s *GroupsService) List(ctx context.Context, opts *ListOptions) ([]Group, *Response, error) {
	if opts == nil {
		opts = &ListOptions{}
	}
	if opts.Version == "" {
		opts.Version = groupsAPIVersion
	}

	path, err := addOptions(groupsBasePath, opts)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.prepareRequest(ctx, http.MethodGet, s.client.restBaseURL, path, nil)
	if err != nil {
		return nil, nil, err
	}

	root := new(groupsRoot)
	resp, err := s.client.do(ctx, req, &root)
	if err != nil {
		return nil, resp, err
	}
	if l := root.Links; l != nil {
		resp.Links = l
	}

	return root.Groups, resp, nil
}

func (s *GroupsService) All(ctx context.Context, opts *ListOptions) (iter.Seq2[Group, *Response], func() error) {
	if opts == nil {
		opts = &ListOptions{}
	}
	if opts.Version == "" {
		opts.Version = groupsAPIVersion
	}
	return newPaginator[Group](ctx, s.client, s.client.restBaseURL, groupsBasePath, opts)
}

func (s *GroupsService) Get(ctx context.Context, groupID string) (*Group, *Response, error) {
	if groupID == "" {
		return nil, nil, errors.New("failed to get org: id must be supplied")
	}

	opts := &BaseOptions{Version: groupsAPIVersion}

	path, err := addOptions(fmt.Sprintf("%v/%v", groupsBasePath, groupID), opts)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.prepareRequest(ctx, http.MethodGet, s.client.restBaseURL, path, nil)
	if err != nil {
		return nil, nil, err
	}

	root := new(groupRoot)
	resp, err := s.client.do(ctx, req, &root)
	if err != nil {
		return nil, resp, err
	}

	return root.Group, resp, nil
}

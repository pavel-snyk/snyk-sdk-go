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

// GroupsService handles communication with the group related methods of the Snyk API.
type GroupsService service

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

// List gets a paginated list of all groups you are a member of.
//
// Note: Group attributes will contain only name. If you want to access full details
// of a group, use Get method.
//
// See: https://docs.snyk.io/snyk-api/reference/groups#get-groups
func (s *GroupsService) List(ctx context.Context, opts *ListOptions) ([]Group, *Response, error) {
	if opts == nil {
		opts = &ListOptions{}
	}

	path, err := restPath(groupsBasePath, groupsAPIVersion, opts)
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

// All returns an iterator to paginate over all groups you are a member of.
//
// This method handles the pagination logic internally by calling List for each page.
// The returned sequence may be iterated multiple times sequentially. It is not safe
// for concurrent or overlapping iteration.
//
// Note: This function is experimental and its signature may change in a future release.
//
// See: https://docs.snyk.io/snyk-api/reference/groups#get-groups
func (s *GroupsService) All(ctx context.Context, opts *ListOptions) (iter.Seq2[Group, *Response], func() error) {
	initialOptions := ListOptions{}
	if opts != nil {
		initialOptions = *opts
	}

	return newPaginator(ctx, initialOptions, func(ctx context.Context, pageOptions ListOptions) ([]Group, *Response, error) {
		return s.List(ctx, &pageOptions)
	})
}

// Get provides the full details of a group.
//
// See: https://docs.snyk.io/snyk-api/reference/group#get-groups-group_id
func (s *GroupsService) Get(ctx context.Context, groupID string) (*Group, *Response, error) {
	if groupID == "" {
		return nil, nil, errors.New("failed to get org: id must be supplied")
	}

	path, err := restPath(fmt.Sprintf("%v/%v", groupsBasePath, groupID), groupsAPIVersion, nil)
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

package snyk

import (
	"context"
	"errors"
	"fmt"
	"iter"
	"net/http"
)

const (
	groupsBasePath   = "groups"
	groupsAPIVersion = "2026-03-25"
)

// GroupsService handles communication with the Groups REST API.
type GroupsService service

// Group represents a Snyk Group.
//
// Its JSON representation is SDK-owned and distinct from the Snyk REST
// JSON:API representation. Group is a read model and is not a REST request
// payload.
type Group struct {
	ID   string `json:"id"`   // The Group identifier.
	Name string `json:"name"` // The Group display name.
}

// String returns a string representation of the Group.
func (g Group) String() string { return Stringify(g) }

// GroupListOptions specifies pagination for List.
type GroupListOptions struct {
	ListOptions
}

type groupAttributes struct {
	Name string `json:"name"`
}

type groupResource struct {
	ID         string           `json:"id"`
	Type       string           `json:"type"`
	Attributes *groupAttributes `json:"attributes"`
}

type groupRoot struct {
	Group *groupResource `json:"data"`
}

type groupsRoot struct {
	Groups []groupResource `json:"data"`
	Links  *PaginatedLinks `json:"links,omitempty"`
}

// List provides one page of Groups accessible to the authenticated user.
//
// See: https://docs.snyk.io/snyk-api/reference/groups#get-groups
func (s *GroupsService) List(ctx context.Context, opts *GroupListOptions) ([]Group, *Response, error) {
	path, err := restPath(groupsBasePath, groupsAPIVersion, opts)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.prepareRequest(ctx, http.MethodGet, s.client.restBaseURL, path, nil)
	if err != nil {
		return nil, nil, err
	}

	root := new(groupsRoot)
	resp, err := s.client.do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}
	if root.Links != nil {
		resp.Links = root.Links
	}
	if root.Groups == nil {
		return nil, resp, errors.New("convert groups: response data is missing")
	}

	groups, err := groupsFromResources(root.Groups)
	if err != nil {
		return nil, resp, err
	}

	return groups, resp, nil
}

// All returns an iterator over Groups accessible to the authenticated user.
//
// Pagination starts from opts.StartingAfter when supplied, so the iterator returns all
// remaining Groups after that cursor rather than restarting from the first page.
// Each page is converted atomically: if a page is malformed, none of its Groups are
// yielded, while Groups from earlier pages remain yielded.
//
// The returned sequence may be iterated multiple times sequentially. It is not safe
// for concurrent or overlapping iteration.
//
// This method is experimental and its signature may change in a future release.
//
// See: https://docs.snyk.io/snyk-api/reference/groups#get-groups
func (s *GroupsService) All(ctx context.Context, opts *GroupListOptions) (iter.Seq2[Group, *Response], func() error) {
	baseOptions := cloneGroupListOptions(opts)
	if baseOptions.EndingBefore != "" {
		validationErr := errors.New("ending-before pagination is not supported when iterating all groups")
		return func(func(Group, *Response) bool) {}, func() error { return validationErr }
	}

	return newPaginator(ctx, baseOptions.ListOptions, func(ctx context.Context, pageOptions ListOptions) ([]Group, *Response, error) {
		currentOptions := baseOptions
		currentOptions.ListOptions = pageOptions
		return s.List(ctx, &currentOptions)
	})
}

// Get provides one Group by Group ID.
//
// See: https://docs.snyk.io/snyk-api/reference/group#get-groups-group_id
func (s *GroupsService) Get(ctx context.Context, groupID string) (*Group, *Response, error) {
	if groupID == "" {
		return nil, nil, fmt.Errorf("group ID: %w", ErrEmptyArgument)
	}

	path, err := restPath(fmt.Sprintf("%s/%s", groupsBasePath, groupID), groupsAPIVersion, nil)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.prepareRequest(ctx, http.MethodGet, s.client.restBaseURL, path, nil)
	if err != nil {
		return nil, nil, err
	}

	root := new(groupRoot)
	resp, err := s.client.do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}
	if root.Group == nil {
		return nil, resp, errors.New("convert group: response data is missing")
	}

	group, err := groupFromResource(*root.Group)
	if err != nil {
		return nil, resp, fmt.Errorf("convert group: %w", err)
	}

	return &group, resp, nil
}

func groupFromResource(resource groupResource) (Group, error) {
	if resource.ID == "" {
		return Group{}, errors.New("resource ID is empty")
	}
	if resource.Type != "group" {
		return Group{}, fmt.Errorf("group %q: resource type is %q, expected %q", resource.ID, resource.Type, "group")
	}
	if resource.Attributes == nil {
		return Group{}, fmt.Errorf("group %q: attributes are missing", resource.ID)
	}

	return Group{
		ID:   resource.ID,
		Name: resource.Attributes.Name,
	}, nil
}

func groupsFromResources(resources []groupResource) ([]Group, error) {
	groups := make([]Group, 0, len(resources))
	for i, resource := range resources {
		group, err := groupFromResource(resource)
		if err != nil {
			return nil, fmt.Errorf("convert group at index %d: %w", i, err)
		}
		groups = append(groups, group)
	}

	return groups, nil
}

func cloneGroupListOptions(opts *GroupListOptions) GroupListOptions {
	if opts == nil {
		return GroupListOptions{}
	}

	return *opts
}

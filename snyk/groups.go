package snyk

import (
	"context"
	"errors"
	"fmt"
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

type groupRoot struct {
	Group *Group `json:"data,omitempty"`
}

func (g Group) String() string { return Stringify(g) }

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

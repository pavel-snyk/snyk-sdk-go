package snyk

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"
)

const (
	projectsBaseBase   = orgsBasePath + "/%v/projects"
	projectsAPIVersion = "2025-11-05"
)

// ProjectsServiceAPI is an interface for interacting with the projects endpoints of the Snyk API.
//
// See: https://docs.snyk.io/snyk-api/reference/projects
type ProjectsServiceAPI interface {
	// List provides a list of all projects for the organization.
	//
	// See: https://docs.snyk.io/snyk-api/reference/projects#get-orgs-org_id-projects
	List(ctx context.Context, orgID string, opts *ListProjectsOptions) ([]Project, *Response, error)

	// Get provides the full details about the project.
	//
	// See: https://docs.snyk.io/snyk-api/reference/projects#get-orgs-org_id-projects-project_id
	Get(ctx context.Context, orgID, projectID string) (*Project, *Response, error)
}

// ProjectsService handles communication with the projects related methods of the Snyk API.
type ProjectsService service

var _ ProjectsServiceAPI = (*ProjectsService)(nil)

// Project represents a Snyk project.
//
// See: https://docs.snyk.io/discover-snyk/getting-started/glossary#project
type Project struct {
	ID         string             `json:"id"`                   // The Project identifier.
	Type       string             `json:"type"`                 // The resource type `project`.
	Attributes *ProjectAttributes `json:"attributes,omitempty"` // The Project resource data.
}

type ProjectAttributes struct {
	CreatedAt       time.Time `json:"created,omitempty"`          // The date that the Project was created at.
	Name            string    `json:"name,omitempty"`             // The name of the Project.
	Origin          string    `json:"origin,omitempty"`           // The origin the Project was added from, e.g. 'github'.
	Status          string    `json:"status,omitempty"`           // The status of the Project describes if the project is monitored or de-activated.
	TargetFile      string    `json:"target_file,omitempty"`      // Path within the target to identify a specific file/directory/image etc.
	TargetReference string    `json:"target_reference,omitempty"` // The additional information required to resolve which revision of the resource should be scanned.
	Type            string    `json:"type,omitempty"`             // The ID of the organization containing the AppInstall.
}

type ListProjectsOptions struct {
	ListOptions
}

type projectRoot struct {
	Project *Project `json:"data"`
}

type projectsRoot struct {
	Projects []Project       `json:"data"`
	Links    *PaginatedLinks `json:"links,omitempty"`
}

func (p Project) String() string { return Stringify(p) }

func (s *ProjectsService) List(ctx context.Context, orgID string, opts *ListProjectsOptions) ([]Project, *Response, error) {
	if orgID == "" {
		return nil, nil, errors.New("orgID must be supplied")
	}

	if opts == nil {
		opts = &ListProjectsOptions{ListOptions: ListOptions{Limit: 100}}
	}
	opts.Version = projectsAPIVersion

	path, err := addOptions(fmt.Sprintf(projectsBaseBase, orgID), opts)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.prepareRequest(ctx, http.MethodGet, s.client.restBaseURL, path, nil)
	if err != nil {
		return nil, nil, err
	}

	root := new(projectsRoot)
	resp, err := s.client.do(ctx, req, &root)
	if err != nil {
		return nil, resp, err
	}
	if l := root.Links; l != nil {
		resp.Links = l
	}

	return root.Projects, resp, nil
}

func (s *ProjectsService) Get(ctx context.Context, orgID, projectID string) (*Project, *Response, error) {
	if orgID == "" {
		return nil, nil, errors.New("orgID must be supplied")
	}
	if projectID == "" {
		return nil, nil, errors.New("projectID must be supplied")
	}

	opts := BaseOptions{Version: projectsAPIVersion}

	path, err := addOptions(fmt.Sprintf(projectsBaseBase+"/%v", orgID, projectID), opts)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.prepareRequest(ctx, http.MethodGet, s.client.restBaseURL, path, nil)
	if err != nil {
		return nil, nil, err
	}

	root := new(projectRoot)
	resp, err := s.client.do(ctx, req, &root)
	if err != nil {
		return nil, resp, err
	}

	return root.Project, resp, nil
}

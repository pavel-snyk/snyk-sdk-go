package snyk

import (
	"context"
	"fmt"
	"net/http"
)

const projectBasePath = "org/%v/projects"

// ProjectsService handles communication with the project related methods of the Snyk API.
type ProjectsService service

// Project represents a Snyk project.
type Project struct {
	ID     string `json:"id,omitempty"`
	Name   string `json:"name,omitempty"`
	Origin string `json:"origin,omitempty"`
}

type projectsRoot struct {
	Organization Organization `json:"org,omitempty"`
	Projects     []Project    `json:"projects,omitempty"`
}

// List provides a list of all projects for the given organization.
func (s *ProjectsService) List(ctx context.Context, organizationID string) ([]Project, *Response, error) {
	if organizationID == "" {
		return nil, nil, ErrEmptyArgument
	}

	path := fmt.Sprintf(projectBasePath, organizationID)
	req, err := s.client.NewRequest(http.MethodPost, path, nil)
	if err != nil {
		return nil, nil, err
	}

	root := new(projectsRoot)
	resp, err := s.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}

	return root.Projects, resp, nil
}

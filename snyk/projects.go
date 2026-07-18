package snyk

import (
	"context"
	"errors"
	"fmt"
	"iter"
	"net/http"
	"slices"
)

const (
	projectsBasePath   = orgsBasePath + "/%v/projects"
	projectsAPIVersion = "2025-11-05"
)

// ProjectsService handles communication with the Projects REST API.
type ProjectsService service

// ProjectType identifies the server-reported ecosystem or analysis type of Project.
// Unknown values remain valid so the SDK can represent new Project types introduced by Snyk.
type ProjectType string

// ProjectOrigin identifies how or through which integration a Project was created.
// Unknown values remain valid so the SDK can represent new origins introduced by Snyk.
type ProjectOrigin string

// ProjectMonitoringStatus identifies whether Snyk is actively monitoring a Project.
// Unknown values remain valid so the SDK can represent new statuses introduced by Snyk.
type ProjectMonitoringStatus string

const (
	// ProjectMonitoringStatusActive identifies a Project that Snyk actively monitors.
	ProjectMonitoringStatusActive ProjectMonitoringStatus = "active"
	// ProjectMonitoringStatusInactive identifies a Project that Snyk does not actively monitor.
	ProjectMonitoringStatusInactive ProjectMonitoringStatus = "inactive"
)

// Project represents a Snyk Project.
//
// Its JSON representation is SDK-owned and distinct from the Snyk REST
// JSON:API representation. Existing JSON field names are part of the public
// SDK contract. Project is a read model and is not a REST request payload.
type Project struct {
	ID               string                  `json:"id"`                // The Project identifier.
	Name             string                  `json:"name"`              // The Project display name.
	ProjectType      ProjectType             `json:"project_type"`      // The ecosystem or analysis type.
	Origin           ProjectOrigin           `json:"origin"`            // How or through which integration the Project was created.
	MonitoringStatus ProjectMonitoringStatus `json:"monitoring_status"` // Whether Snyk actively monitors the Project.

	// TargetPath identifies the Project's location within its target, such as a
	// manifest file or a directory within a scanned image. It is empty when Snyk
	// does not report a narrower location within the target.
	TargetPath string `json:"target_path"`

	// TargetReference contains target-specific information used to resolve the
	// revision or variant that was scanned. For example, it may be a source-control
	// branch or another origin-specific reference.
	TargetReference string `json:"target_reference"`
}

// String returns a string representation of the Project.
func (p Project) String() string { return Stringify(p) }

// ProjectListOptions specifies filters and pagination for List.
type ProjectListOptions struct {
	ListOptions
	IDs             []string        `url:"ids,comma,omitempty"`              // Return Projects matching these IDs.
	Names           []string        `url:"names,comma,omitempty"`            // Return Projects matching these names.
	NamePrefixes    []string        `url:"names_start_with,comma,omitempty"` // Return Projects whose names start with these prefixes.
	ProjectTypes    []ProjectType   `url:"types,comma,omitempty"`            // Return Projects matching these Project types.
	Origins         []ProjectOrigin `url:"origins,comma,omitempty"`          // Return Projects matching these origins.
	TargetPath      string          `url:"target_file,omitempty"`            // Return Projects matching this location within their target.
	TargetReference string          `url:"target_reference,omitempty"`       // Return Projects matching this target reference.
}

type projectAttributes struct {
	Name            string `json:"name"`
	Type            string `json:"type"`
	Origin          string `json:"origin"`
	Status          string `json:"status"`
	TargetFile      string `json:"target_file"`
	TargetReference string `json:"target_reference"`
}

type projectResource struct {
	ID         string             `json:"id"`
	Type       string             `json:"type"`
	Attributes *projectAttributes `json:"attributes"`
}

type projectRoot struct {
	Project *projectResource `json:"data"`
}

type projectsRoot struct {
	Projects []projectResource `json:"data"`
	Links    *PaginatedLinks   `json:"links,omitempty"`
}

// List provides one page of Projects for an organization.
//
// See: https://docs.snyk.io/snyk-api/reference/projects#get-orgs-org_id-projects
func (s *ProjectsService) List(ctx context.Context, orgID string, opts *ProjectListOptions) ([]Project, *Response, error) {
	if orgID == "" {
		return nil, nil, fmt.Errorf("organization ID: %w", ErrEmptyArgument)
	}

	path, err := restPath(fmt.Sprintf(projectsBasePath, orgID), projectsAPIVersion, opts)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.prepareRequest(ctx, http.MethodGet, s.client.restBaseURL, path, nil)
	if err != nil {
		return nil, nil, err
	}

	root := new(projectsRoot)
	resp, err := s.client.do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}
	if root.Links != nil {
		resp.Links = root.Links
	}
	if root.Projects == nil {
		return nil, resp, errors.New("convert projects: response data is missing")
	}

	projects, err := projectsFromResources(root.Projects)
	if err != nil {
		return nil, resp, err
	}

	return projects, resp, nil
}

// All returns an iterator over Projects for an organization.
//
// Pagination starts from opts.StartingAfter when supplied, so the iterator returns all
// remaining Projects after that cursor rather than restarting from the first page.
// Each page is converted atomically: if a page is malformed, none of its Projects are
// yielded, while Projects from earlier pages remain yielded.
//
// The returned sequence may be iterated multiple times sequentially. It is not safe
// for concurrent or overlapping iteration.
//
// This method is experimental and its signature may change in a future release.
//
// See: https://docs.snyk.io/snyk-api/reference/projects#get-orgs-org_id-projects
func (s *ProjectsService) All(ctx context.Context, orgID string, opts *ProjectListOptions) (iter.Seq2[Project, *Response], func() error) {
	if orgID == "" {
		validationErr := fmt.Errorf("organization ID: %w", ErrEmptyArgument)
		return func(func(Project, *Response) bool) {}, func() error { return validationErr }
	}

	baseOptions := cloneProjectListOptions(opts)
	if baseOptions.EndingBefore != "" {
		validationErr := errors.New("ending-before pagination is not supported when iterating all projects")
		return func(func(Project, *Response) bool) {}, func() error { return validationErr }
	}

	return newPaginator(ctx, baseOptions.ListOptions, func(ctx context.Context, pageOptions ListOptions) ([]Project, *Response, error) {
		currentOptions := baseOptions
		currentOptions.ListOptions = pageOptions
		return s.List(ctx, orgID, &currentOptions)
	})
}

// Get provides one Project by organization and Project ID.
//
// See: https://docs.snyk.io/snyk-api/reference/projects#get-orgs-org_id-projects-project_id
func (s *ProjectsService) Get(ctx context.Context, orgID, projectID string) (*Project, *Response, error) {
	if orgID == "" {
		return nil, nil, fmt.Errorf("organization ID: %w", ErrEmptyArgument)
	}
	if projectID == "" {
		return nil, nil, fmt.Errorf("project ID: %w", ErrEmptyArgument)
	}

	path, err := restPath(fmt.Sprintf(projectsBasePath+"/%v", orgID, projectID), projectsAPIVersion, nil)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.prepareRequest(ctx, http.MethodGet, s.client.restBaseURL, path, nil)
	if err != nil {
		return nil, nil, err
	}

	root := new(projectRoot)
	resp, err := s.client.do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}
	if root.Project == nil {
		return nil, resp, errors.New("convert project: response data is missing")
	}

	project, err := projectFromResource(*root.Project)
	if err != nil {
		return nil, resp, fmt.Errorf("convert project: %w", err)
	}

	return &project, resp, nil
}

// Delete deletes one Project.
//
// See: https://docs.snyk.io/snyk-api/reference/projects#delete-orgs-org_id-projects-project_id
func (s *ProjectsService) Delete(ctx context.Context, orgID, projectID string) (*Response, error) {
	if orgID == "" {
		return nil, fmt.Errorf("organization ID: %w", ErrEmptyArgument)
	}
	if projectID == "" {
		return nil, fmt.Errorf("project ID: %w", ErrEmptyArgument)
	}

	path, err := restPath(fmt.Sprintf(projectsBasePath+"/%v", orgID, projectID), projectsAPIVersion, nil)
	if err != nil {
		return nil, err
	}

	req, err := s.client.prepareRequest(ctx, http.MethodDelete, s.client.restBaseURL, path, nil)
	if err != nil {
		return nil, err
	}

	return s.client.do(ctx, req, nil)
}

func projectFromResource(resource projectResource) (Project, error) {
	if resource.ID == "" {
		return Project{}, errors.New("resource ID is empty")
	}
	if resource.Type != "project" {
		return Project{}, fmt.Errorf("project %q: resource type is %q, expected %q", resource.ID, resource.Type, "project")
	}
	if resource.Attributes == nil {
		return Project{}, fmt.Errorf("project %q: attributes are missing", resource.ID)
	}

	return Project{
		ID:               resource.ID,
		Name:             resource.Attributes.Name,
		ProjectType:      ProjectType(resource.Attributes.Type),
		Origin:           ProjectOrigin(resource.Attributes.Origin),
		MonitoringStatus: ProjectMonitoringStatus(resource.Attributes.Status),
		TargetPath:       resource.Attributes.TargetFile,
		TargetReference:  resource.Attributes.TargetReference,
	}, nil
}

func projectsFromResources(resources []projectResource) ([]Project, error) {
	projects := make([]Project, 0, len(resources))
	for i, resource := range resources {
		project, err := projectFromResource(resource)
		if err != nil {
			return nil, fmt.Errorf("convert project at index %d: %w", i, err)
		}
		projects = append(projects, project)
	}

	return projects, nil
}

func cloneProjectListOptions(opts *ProjectListOptions) ProjectListOptions {
	if opts == nil {
		return ProjectListOptions{}
	}

	cloned := *opts
	cloned.IDs = slices.Clone(opts.IDs)
	cloned.Names = slices.Clone(opts.Names)
	cloned.NamePrefixes = slices.Clone(opts.NamePrefixes)
	cloned.ProjectTypes = slices.Clone(opts.ProjectTypes)
	cloned.Origins = slices.Clone(opts.Origins)
	return cloned
}

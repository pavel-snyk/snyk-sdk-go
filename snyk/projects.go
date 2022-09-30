package snyk

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

const projectBasePath = "org/%v/projects"
const projectPath = "org/%s/project/%s"
const projectTagsPath = "org/%s/project/%s/tags"

// ProjectsService handles communication with the project related methods of the Snyk API.
type ProjectsService service

// Project represents a Snyk project.
type Project struct {
	ID     string `json:"id,omitempty"`
	Name   string `json:"name,omitempty"`
	Origin string `json:"origin,omitempty"`

	OrgId string `json:"-"`
}

type projectsRoot struct {
	Organization Organization `json:"org,omitempty"`
	Projects     []Project    `json:"projects,omitempty"`
}

type projectAddTagRequest struct {
	Key   string `json:"key"`
	Value string `json:"value"`
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

	for k, v := range root.Projects {
		v.OrgId = root.Organization.ID
		root.Projects[k] = v
	}

	return root.Projects, resp, nil
}

type Tag struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type ProjectDetails struct {
	Name                  string             `json:"name"`
	Id                    string             `json:"id"`
	Created               time.Time          `json:"created"`
	Origin                string             `json:"origin"`
	Type                  string             `json:"type"`
	ReadOnly              bool               `json:"readOnly"`
	TestFrequency         string             `json:"testFrequency"`
	TotalDependencies     int                `json:"totalDependencies"`
	IssueCountsBySeverity IssueCounts        `json:"issueCountsBySeverity"`
	ImageId               string             `json:"imageId"`
	ImageTag              string             `json:"imageTag"`
	ImageBaseImage        string             `json:"imageBaseImage"`
	ImagePlatform         string             `json:"imagePlatform"`
	ImageCluster          string             `json:"imageCluster"`
	Hostname              string             `json:"hostname"`
	RemoteRepoUrl         string             `json:"remoteRepoUrl"`
	LastTestedDate        time.Time          `json:"lastTestedDate"`
	BrowseUrl             string             `json:"browseUrl"`
	ImportingUser         UserRef            `json:"importingUser"`
	IsMonitored           bool               `json:"isMonitored"`
	Branch                string             `json:"branch"`
	TargetReference       string             `json:"targetReference"`
	Tags                  []Tag              `json:"tags"`
	Attributes            ProjectAttributes  `json:"attributes"`
	Remediation           ProjectRemediation `json:"remediation"`
}

type ProjectAttributes struct {
	Criticality []string `json:"criticality"`
	Environment []string `json:"environment"`
	Lifecycle   []string `json:"lifecycle"`
}

type ProjectRemediation struct {
	Upgrade struct{} `json:"upgrade"`
	Patch   struct{} `json:"patch"`
	Pin     struct{} `json:"pin"`
}

func (s *ProjectsService) AddTag(ctx context.Context, p *Project, key string, value string) ([]Tag, *Response, error) {
	path := fmt.Sprintf(projectTagsPath, p.OrgId, p.ID)

	req, err := s.client.NewRequest(http.MethodPost, path, projectAddTagRequest{
		Key:   key,
		Value: value,
	})
	if err != nil {
		return nil, nil, err
	}

	root := new([]Tag)
	resp, err := s.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}

	return *root, resp, nil
}

func (s *ProjectsService) Details(ctx context.Context, p *Project) (*ProjectDetails, *Response, error) {
	path := fmt.Sprintf(projectPath, p.OrgId, p.ID)

	req, err := s.client.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}

	root := new(ProjectDetails)
	resp, err := s.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}

	return root, resp, nil
}

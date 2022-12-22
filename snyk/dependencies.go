package snyk

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

const (
	dependencyBasePath = "org/%v/dependencies"
	dependencyPageSize = 1000
)

// DependenciesService handles communication with the dependencies related methods of the Snyk API.
type DependenciesService service

// Dependency represents a Snyk dependency.
type Dependency struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
	Type string `json:"type,omitempty"`

	Version string `json:"version,omitempty"`

	LatestVersion              string     `json:"latestVersion,omitempty"`
	LatestVersionPublishedDate *time.Time `json:"latestVersionPublishedDate,omitempty"`

	FirstPublishedDate *time.Time `json:"firstPublishedDate,omitempty"`

	IsDeprecated       bool     `json:"isDeprecated,omitempty"`
	DeprecatedVersions []string `json:"deprecatedVersions,omitempty"`

	DependenciesWithIssues []string `json:"dependenciesWithIssues,omitempty"`

	IssuesCritical int `json:"issuesCritical,omitempty"`
	IssuesHigh     int `json:"issuesHigh,omitempty"`
	IssuesMedium   int `json:"issuesMedium,omitempty"`
	IssuesLow      int `json:"issuesLow,omitempty"`

	Licenses []DependencyLicense `json:"licenses,omitempty"`

	Projects []Project `json:"projects,omitempty"` // origin will be always empty, but that's OK

	Copyright []string `json:"copyright,omitempty"`
}

type DependencyLicense struct {
	ID      string `json:"id,omitempty"`
	Title   string `json:"title,omitempty"`
	License string `json:"license,omitempty"`
}

type dependenciesRoot struct {
	Total   int          `json:"total,omitempty"`
	Results []Dependency `json:"results"`
}

// List provides a list of all dependencies for the given organization.
// Response returned will be the last response from the service.
func (s *DependenciesService) List(ctx context.Context, organizationID string) ([]Dependency, *Response, error) {
	if organizationID == "" {
		return nil, nil, ErrEmptyArgument
	}

	var result []Dependency
	for page := 1; ; page++ {
		deps, resp, err := s.ListPage(ctx, organizationID, page)
		if err != nil {
			return result, resp, err
		}

		result = append(result, deps...)
		if len(deps) < dependencyPageSize {
			return result, resp, nil
		}
	}
}

// ListPage lists dependencies page, returning both dependencies & total amount returned
func (s *DependenciesService) ListPage(ctx context.Context, organizationID string, page int) ([]Dependency, *Response, error) {
	if organizationID == "" {
		return nil, nil, ErrEmptyArgument
	}

	path := fmt.Sprintf(dependencyBasePath+"?sortBy=dependency&order=asc&perPage=%d&page=%d",
		organizationID, dependencyPageSize, page)

	req, err := s.client.NewRequest(http.MethodPost, path, nil)
	if err != nil {
		return nil, nil, err
	}

	root := new(dependenciesRoot)
	resp, err := s.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}

	return root.Results, resp, nil
}

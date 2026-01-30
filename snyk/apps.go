package snyk

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

const (
	appsBasePath   = "apps"
	appsAPIVersion = "2025-11-05"
)

// AppsServiceAPI is an interface for interacting with the apps endpoints of the Snyk API.
//
// See: https://docs.snyk.io/snyk-api/reference/apps
type AppsServiceAPI interface {
	// ListAppInstallsForOrg gets a list of Snyk Apps installed for an Organization. If ListAppInstallOptions is nil,
	// then relationship for App will be always expanded.
	//
	// See: https://docs.snyk.io/snyk-api/reference/apps#get-orgs-org_id-apps-installs
	ListAppInstallsForOrg(ctx context.Context, orgID string, opts *ListAppInstallOptions) ([]AppInstall, *Response, error)
}

// AppsService handles communication with the app related methods of the Snyk API.
type AppsService service

var _ AppsServiceAPI = &AppsService{}

// App represents a Snyk app.
//
// See: https://docs.snyk.io/snyk-api/using-specific-snyk-apis/snyk-apps-apis/about-snyk-apps
type App struct {
	ID         string         `json:"id"`                   // The App identifier.
	Type       string         `json:"type"`                 // The resource type `app`.
	Attributes *AppAttributes `json:"attributes,omitempty"` // The App resource data.
}

type AppAttributes struct {
	ClientID string   `json:"client_id"`         // The oauth2 client id for the app.
	Context  string   `json:"context,omitempty"` // Allow installing the app to at org/group level or user level. Defaults to tenant.
	Name     string   `json:"name"`              // The name of the app.
	Scopes   []string `json:"scopes,omitempty"`  // The scopes this app is allowed to request during authorization.
}

type appRoot struct {
	Data *App `json:"data,omitempty"`
}

// AppInstall represents an installation of an App.
type AppInstall struct {
	ID            string                   `json:"id"`                      // The AppInstall identifier.
	Type          string                   `json:"type"`                    // The resource type `app_install`.
	Attributes    *AppInstallAttributes    `json:"attributes,omitempty"`    // The AppInstall resource data.
	Relationships *AppInstallRelationships `json:"relationships,omitempty"` // The relationships object describing relationships between AppInstall and App.
}

type AppInstallAttributes struct {
	ClientID     string    `json:"client_id"`               // ClientID of the client that the AppInstall belongs to.
	ClientSecret string    `json:"client_secret,omitempty"` // ClientSecret available only when client secret is rotated.
	InstalledAt  time.Time `json:"installed_at"`            // Timestamp at which the App was first installed at.
}

type AppInstallRelationships struct {
	App appRoot `json:"app"`
}

type ListAppInstallOptions struct {
	ListOptions
	Expand string `url:"expand,omitempty"`
}

type appInstallsRoot struct {
	AppInstalls []AppInstall `json:"data"`
}

func (ai AppInstall) String() string { return Stringify(ai) }

func (s *AppsService) ListAppInstallsForOrg(ctx context.Context, orgID string, opts *ListAppInstallOptions) ([]AppInstall, *Response, error) {
	if opts == nil {
		opts = &ListAppInstallOptions{Expand: "app"}
	}
	opts.Version = appsAPIVersion

	path, err := addOptions(fmt.Sprintf("orgs/%v/%v/installs", orgID, appsBasePath), opts)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.prepareRequest(ctx, http.MethodGet, s.client.restBaseURL, path, nil)
	if err != nil {
		return nil, nil, err
	}

	root := new(appInstallsRoot)
	resp, err := s.client.do(ctx, req, &root)
	if err != nil {
		return nil, resp, err
	}

	return root.AppInstalls, resp, nil
}

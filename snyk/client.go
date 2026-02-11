package snyk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"slices"
	"strings"
	"time"

	"github.com/google/go-querystring/query"
)

const (
	libraryVersion = "2.0.0-dev"

	defaultRegion    = "SNYK-US-01"
	defaultMediaType = "application/json"
	restAPIMediaType = "application/vnd.api+json"
	defaultUserAgent = "snyk-sdk-go/" + libraryVersion + " (+https://github.com/pavel-snyk/snyk-sdk-go)"

	headerSnykRequestID = "snyk-request-id"
)

// A Client manages communication with the Snyk API.
type Client struct {
	httpClient *http.Client

	appBaseURL  *url.URL // base URL for App related requests (used to get access token).
	restBaseURL *url.URL // base URL for REST API requests.
	v1BaseURL   *url.URL // base URL for V1 API requests.

	userAgent string
	token     string

	common service // reuse a single struct instead of allocating one for each service on the heap.

	Apps   AppsServiceAPI
	Groups GroupsServiceAPI
	Orgs   OrgsServiceAPI
	OrgsV1 OrgsServiceV1API
}

// Region is used to configure the SDK to communicate with different Snyk regional instances.
// Snyk operates several independent, isolated instances (e.g. in the US, EU and AU).
//
// Snyk docs: https://docs.snyk.io/snyk-data-and-governance/regional-hosting-and-data-residency#api-urls
type Region struct {
	Alias       string
	AppBaseURL  string
	RESTBaseURL string
	V1BaseURL   string
}

var regions = []Region{
	{
		Alias:       defaultRegion,
		AppBaseURL:  "https://app.snyk.io/",
		RESTBaseURL: "https://api.snyk.io/rest/",
		V1BaseURL:   "https://api.snyk.io/v1/",
	},
	{
		Alias:       "SNYK-US-02",
		AppBaseURL:  "https://app.us.snyk.io/",
		RESTBaseURL: "https://api.us.snyk.io/rest/",
		V1BaseURL:   "https://api.us.snyk.io/v1/",
	},
	{
		Alias:       "SNYK-EU-01",
		AppBaseURL:  "https://app.eu.snyk.io/",
		RESTBaseURL: "https://app.eu.snyk.io/rest/",
		V1BaseURL:   "https://api.eu.snyk.io/v1/",
	},
	{
		Alias:       "SNYK-AU-01",
		AppBaseURL:  "https://app.au.snyk.io/",
		RESTBaseURL: "https://api.au.snyk.io/rest/",
		V1BaseURL:   "https://api.au.snyk.io/v1/",
	},
}

// Regions provides a slice of all supported Snyk Regions.
func Regions() []Region {
	regionsCopy := make([]Region, len(regions))
	copy(regionsCopy, regions)
	return regionsCopy
}

type service struct {
	client *Client
}

type BaseOptions struct {
	// The requested API version. This query parameter is required.
	Version string `url:"version"`
}

// ListOptions specifies the optional parameters to various List methods.
type ListOptions struct {
	BaseOptions

	// The page of results immediately after this cursor.
	StartingAfter string `url:"starting_after,omitempty"`

	// The page of results immediately before this cursor.
	EndingBefore string `url:"ending_before,omitempty"`

	// Number of results to return per page
	Limit int `url:"limit,omitempty"`
}

// addOptions adds the parameters in opts as URL query parameters to s.
// opts must be a struct whose  fields may contain "url" tags.
func addOptions(s string, opts any) (string, error) {
	v := reflect.ValueOf(opts)
	if v.Kind() == reflect.Ptr && v.IsNil() {
		return s, nil
	}

	u, err := url.Parse(s)
	if err != nil {
		return s, err
	}

	qs, err := query.Values(opts)
	if err != nil {
		return s, err
	}

	u.RawQuery = qs.Encode()
	return u.String(), nil
}

type ClientOption func(*Client) error

// WithRegionAlias configures Client to use a Snyk Region by querying Regions by Region.Alias.
func WithRegionAlias(regionAlias string) ClientOption {
	return func(client *Client) error {
		regionIndex := slices.IndexFunc(regions, func(r Region) bool {
			return r.Alias == regionAlias
		})
		if regionIndex == -1 {
			return fmt.Errorf("region with alias (%s) not found", regionAlias)
		}

		region := regions[regionIndex]
		return WithRegion(region)(client)
	}
}

// WithRegion configures Client to use a specific Snyk Region.
func WithRegion(region Region) ClientOption {
	return func(client *Client) error {
		parsedAppBaseURL, err := url.Parse(region.AppBaseURL)
		if err != nil {
			return fmt.Errorf("invalid AppBaseURL: %w", err)
		}
		client.appBaseURL = parsedAppBaseURL

		parsedRestBaseURL, err := url.Parse(region.RESTBaseURL)
		if err != nil {
			return fmt.Errorf("invalid RESTBaseURL: %w", err)
		}
		client.restBaseURL = parsedRestBaseURL

		parsedV1BaseURL, err := url.Parse(region.V1BaseURL)
		if err != nil {
			return fmt.Errorf("invalid V1BaseURL: %w", err)
		}
		client.v1BaseURL = parsedV1BaseURL

		return nil
	}
}

// WithHTTPClient configures Client to use a specific http client for communication.
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(client *Client) error {
		client.httpClient = httpClient
		return nil
	}
}

// WithUserAgent configures Client to use a specific user agent.
func WithUserAgent(userAgent string) ClientOption {
	return func(client *Client) error {
		client.userAgent = userAgent
		return nil
	}
}

// NewClient creates a new Snyk API client.
func NewClient(token string, opts ...ClientOption) (*Client, error) {
	httpClient := &http.Client{
		Timeout: 15 * time.Second,
	}

	c := &Client{
		httpClient: httpClient,

		userAgent: defaultUserAgent,
		token:     token,
	}

	// apply default region first, can be overridden by options later
	if err := WithRegionAlias(defaultRegion)(c); err != nil {
		return nil, err
	}

	// apply user options
	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, err
		}
	}

	c.common.client = c

	c.Apps = (*AppsService)(&c.common)
	c.Groups = (*GroupsService)(&c.common)
	c.Orgs = (*OrgsService)(&c.common)
	c.OrgsV1 = (*OrgsServiceV1)(&c.common)

	return c, nil
}

// prepareRequest creates an API request. A relative URL can be provided in endpointURL, which will be resolved to the baseURL.
func (c *Client) prepareRequest(ctx context.Context, method string, baseURL *url.URL, endpointURL string, body any) (*http.Request, error) {
	if strings.HasPrefix(endpointURL, "/") {
		return nil, fmt.Errorf("endpointURL %q is invalid, cannot begin with a leading slash", endpointURL)
	}

	u, err := baseURL.Parse(endpointURL)
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	if body != nil {
		err := json.NewEncoder(buf).Encode(body)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, u.String(), buf)
	if err != nil {
		return nil, err
	}
	if req.Header.Get("Authorization") == "" {
		req.Header.Set("Authorization", fmt.Sprintf("Token %s", c.token))
	}

	mediaType := defaultMediaType
	if baseURL == c.restBaseURL {
		mediaType = restAPIMediaType
	}
	req.Header.Set("Accept", mediaType)
	req.Header.Set("Content-Type", mediaType)
	req.Header.Set("User-Agent", c.userAgent)

	return req, nil
}

// do sends an API request and returns the API response.
func (c *Client) do(ctx context.Context, req *http.Request, v interface{}) (*Response, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		// if we got an error and the context has been canceled, the context's error is more useful.
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	response := newResponse(resp)
	err = checkResponse(response)
	if err != nil {
		return response, err
	}

	if resp.StatusCode != http.StatusNoContent && v != nil {
		if w, ok := v.(io.Writer); ok {
			_, err := io.Copy(w, resp.Body)
			if err != nil {
				return nil, err
			}
		} else {
			decodedErr := json.NewDecoder(resp.Body).Decode(v)
			if decodedErr == io.EOF {
				// ignore EOF errors cause by empty response body
				decodedErr = nil
			}
			if decodedErr != nil {
				err = decodedErr
			}
		}
	}

	return response, err
}

// Response is a Snyk response. This wraps the standard http.Response.
type Response struct {
	*http.Response

	// Links that were returned with the response. These are parsed from request body and not the header.
	Links *PaginatedLinks

	SnykRequestID string // SnykRequestID returned from the API, useful to contact support.
}

// newResponse creates a new Response for the provided http.Response. r must be not nil.
func newResponse(r *http.Response) *Response {
	response := &Response{Response: r}
	response.populateSnykRequestID()
	return response
}

func (r *Response) populateSnykRequestID() {
	if snykRequestID := r.Header.Get(headerSnykRequestID); snykRequestID != "" {
		r.SnykRequestID = snykRequestID
	}
}

// checkResponse checks the API response for errors and returns them if present.
// An error is returned if the status code is outside the 2xx range. It attempts
// to parse the error body first as a Snyk REST API error (JSON:API), then falls
// back to the legacy V1 API error format.
func checkResponse(resp *Response) error {
	if code := resp.StatusCode; code >= 200 && code <= 299 {
		return nil
	}

	errorResponse := &ErrorResponse{Response: resp}
	data, err := io.ReadAll(resp.Body)
	if err != nil || len(data) == 0 {
		return errorResponse
	}

	// try parsing as jsonapi error
	if apiErrors, ok := parseRESTError(data); ok {
		errorResponse.APIErrors = apiErrors
		return errorResponse
	}
	// fallback to parsing as legacy V1 error
	if apiErrors, ok := parseLegacyV1Error(data, resp.StatusCode); ok {
		errorResponse.APIErrors = apiErrors
		return errorResponse
	}

	return fmt.Errorf("failed to decode Snyk API error response; status: %d, body: %s", resp.StatusCode, string(data))
}

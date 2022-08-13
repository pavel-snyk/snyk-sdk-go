package snyk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	libraryVersion   = "0.1.0-dev"
	defaultBaseURL   = "https://snyk.io/api/"
	defaultMediaType = "application/json"
	defaultUserAgent = "snyk-sdk-go/" + libraryVersion + " (+https://github.com/pavel-snyk/snyk-sdk-go)"

	headerSnykRequestID = "snyk-request-id"
)

// A Client manages communication with the Snyk API.
type Client struct {
	httpClient *http.Client

	BaseURL   *url.URL // Base URL for API requests.
	UserAgent string
	Token     string

	common service // Reuse a single struct instead of allocating one for each service on the heap.

	Users *UsersService
}

type service struct {
	client *Client
}

type ClientOption func(*Client)

// WithBaseURL configures Client to use a specific API endpoint.
func WithBaseURL(baseURL string) ClientOption {
	return func(client *Client) {
		parsedURL, _ := url.Parse(baseURL)
		client.BaseURL = parsedURL
	}
}

// WithHTTPClient configures Client to use a specific http client for communication.
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(client *Client) {
		client.httpClient = httpClient
	}
}

// WithUserAgent configures Client to use a specific user agent.
func WithUserAgent(userAgent string) ClientOption {
	return func(client *Client) {
		client.UserAgent = userAgent
	}
}

// NewClient creates a new Snyk API client.
func NewClient(token string, opts ...ClientOption) *Client {
	baseURL, _ := url.Parse(defaultBaseURL)
	httpClient := &http.Client{
		Timeout: 15 * time.Second,
	}

	c := &Client{
		httpClient: httpClient,

		BaseURL:   baseURL,
		UserAgent: defaultUserAgent,
		Token:     token,
	}
	for _, opt := range opts {
		opt(c)
	}

	c.common.client = c

	c.Users = (*UsersService)(&c.common)

	return c
}

// NewRequest creates an API request.
func (c *Client) NewRequest(method, urlStr string, body interface{}) (*http.Request, error) {
	if !strings.HasSuffix(c.BaseURL.Path, "/") {
		return nil, fmt.Errorf("BaseURL must have a trailing slash, but %q does not", c.BaseURL)
	}
	u, err := c.BaseURL.Parse(urlStr)
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

	req, err := http.NewRequest(method, u.String(), buf)
	if err != nil {
		return nil, err
	}
	if req.Header.Get("Authorization") == "" {
		req.Header.Set("Authorization", fmt.Sprintf("token %s", c.Token))
	}
	req.Header.Set("Accept", defaultMediaType)
	req.Header.Set("Content-Type", defaultMediaType)
	req.Header.Set("User-Agent", c.UserAgent)

	return req, nil
}

// Response is a Snyk response. This wraps the standard http.Response.
type Response struct {
	*http.Response

	SnykRequestID string // SnykRequestID returned from the API, useful to contact support.
}

// Do sends an API request and returns the API response.
func (c *Client) Do(ctx context.Context, req *http.Request, v interface{}) (*Response, error) {
	req = req.WithContext(ctx)
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
	err = CheckResponse(response)
	if err != nil {
		return response, err
	}

	if v != nil {
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

// CheckResponse verifies the API responds for errors, and returns them if present.
// A response is considered an error if it has a status code outside the 200 range.
func CheckResponse(resp *Response) error {
	if code := resp.StatusCode; code >= 200 && code <= 299 {
		return nil
	}

	errorResponse := &ErrorResponse{Response: resp}
	data, err := io.ReadAll(resp.Body)
	if err == nil && len(data) > 0 {
		err := json.Unmarshal(data, &errorResponse.ErrorElement)
		if err != nil {
			return err
		}
	}
	return errorResponse
}

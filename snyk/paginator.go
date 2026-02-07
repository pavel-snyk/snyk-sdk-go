package snyk

import (
	"context"
	"fmt"
	"iter"
	"net/http"
	"net/url"
)

// paginatedResponse is a generic "container" used to unmarshal any paginated list response
// from Snyk REST API. It assumes the response body has a "data" field (according jsonapi),
// containing a slice of items and a "links" field for pagination.
type paginatedResponse[T any] struct {
	Data  []T             `json:"data,omitempty"`
	Links *PaginatedLinks `json:"links"`
}

func newPaginator[T any](ctx context.Context, client *Client, baseURL *url.URL, endpointURL string, opts *ListOptions) (iter.Seq2[T, *Response], func() error) {
	var iterErr error

	seq := func(yield func(item T, resp *Response) bool) {
		if opts == nil {
			iterErr = fmt.Errorf("ListOptions cannot be nil, API version is required for endpoint %q", endpointURL)
			return
		}

		for {
			select {
			// if the context has been canceled, the context's error is more useful
			case <-ctx.Done():
				iterErr = ctx.Err()
				return
			default:
			}

			path, err := addOptions(endpointURL, opts)
			if err != nil {
				iterErr = fmt.Errorf("failed to construct URL with options: %w", err)
				return
			}
			req, err := client.prepareRequest(ctx, http.MethodGet, baseURL, path, nil)
			if err != nil {
				iterErr = fmt.Errorf("failed to prepare pagination request: %w", err)
				return
			}

			page := new(paginatedResponse[T])

			resp, err := client.do(ctx, req, page)
			if err != nil {
				iterErr = err
				return
			}
			if l := page.Links; l != nil {
				resp.Links = l
			}

			for _, item := range page.Data {
				if !yield(item, resp) {
					// stop iteration if the consumer stops
					return
				}
			}

			if resp.Links == nil || resp.Links.Next == "" {
				// no more next pages, exit from pagination
				break
			}

			startingAfter, err := extractStartingAfterQueryParam(resp.Links.Next)
			if err != nil {
				iterErr = fmt.Errorf("failed to extract starting_after query param: %w", err)
				return
			}
			opts.StartingAfter = startingAfter
		}
	}

	return seq, func() error { return iterErr }
}

// extractStartingAfterQueryParam extracts the value of the "starting_after" query parameter from a URL path.
// The Snyk API uses this token for cursor-based pagination.
func extractStartingAfterQueryParam(path string) (string, error) {
	u, err := url.Parse(path)
	if err != nil {
		return "", fmt.Errorf("failed to parse pagination path: %w", err)
	}

	q := u.Query()
	return q.Get("starting_after"), nil
}

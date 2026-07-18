package snyk

import (
	"context"
	"errors"
	"fmt"
	"iter"
	"net/url"
)

type pageFetcher[T any] func(context.Context, ListOptions) ([]T, *Response, error)

func newPaginator[T any](ctx context.Context, initialOptions ListOptions, fetchPage pageFetcher[T]) (iter.Seq2[T, *Response], func() error) {
	var iterErr error

	seq := func(yield func(item T, resp *Response) bool) {
		iterErr = nil
		paginationOptions := initialOptions
		seenCursors := make(map[string]struct{})
		if paginationOptions.StartingAfter != "" {
			seenCursors[paginationOptions.StartingAfter] = struct{}{}
		}

		for {
			if err := ctx.Err(); err != nil {
				iterErr = err
				return
			}

			items, resp, err := fetchPage(ctx, paginationOptions)
			if err != nil {
				iterErr = err
				return
			}
			if resp == nil {
				iterErr = errors.New("pagination response is missing")
				return
			}

			for _, item := range items {
				if !yield(item, resp) {
					return
				}
			}

			if resp.Links == nil || resp.Links.Next == "" {
				return
			}

			startingAfter, err := extractStartingAfterQueryParam(resp.Links.Next)
			if err != nil {
				iterErr = fmt.Errorf("failed to extract starting_after query param: %w", err)
				return
			}
			if startingAfter == "" {
				iterErr = errors.New("next-page link has no starting_after cursor")
				return
			}
			if _, seen := seenCursors[startingAfter]; seen {
				iterErr = errors.New("next-page link repeats a previously seen starting_after cursor")
				return
			}

			seenCursors[startingAfter] = struct{}{}
			paginationOptions.StartingAfter = startingAfter
			paginationOptions.EndingBefore = ""
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

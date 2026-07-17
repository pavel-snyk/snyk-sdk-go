package snyk

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPaginator_newPaginator_multiplePages(t *testing.T) {
	var gotOptions []ListOptions
	firstResponse := &Response{Links: &PaginatedLinks{Next: "/rest/items?starting_after=next-cursor"}}
	secondResponse := &Response{}
	fetch := func(_ context.Context, opts ListOptions) ([]string, *Response, error) {
		gotOptions = append(gotOptions, opts)
		if len(gotOptions) == 1 {
			return []string{"first", "second"}, firstResponse, nil
		}
		return []string{"third"}, secondResponse, nil
	}

	seq, iterErr := newPaginator(context.Background(), ListOptions{
		EndingBefore: "initial-ending-cursor",
		Limit:        20,
	}, fetch)

	var gotItems []string
	var gotResponses []*Response
	for item, response := range seq {
		gotItems = append(gotItems, item)
		gotResponses = append(gotResponses, response)
	}

	require.NoError(t, iterErr())
	assert.Equal(t, []string{"first", "second", "third"}, gotItems)
	assert.Equal(t, []*Response{firstResponse, firstResponse, secondResponse}, gotResponses)
	assert.Equal(t, []ListOptions{
		{EndingBefore: "initial-ending-cursor", Limit: 20},
		{StartingAfter: "next-cursor", Limit: 20},
	}, gotOptions)
}

func TestPaginator_newPaginator_restartsSequentialIteration(t *testing.T) {
	var gotOptions []ListOptions
	fetch := func(_ context.Context, opts ListOptions) ([]string, *Response, error) {
		gotOptions = append(gotOptions, opts)
		if opts.StartingAfter == "initial-cursor" {
			return []string{"first"}, &Response{Links: &PaginatedLinks{Next: "/rest/items?starting_after=next-cursor"}}, nil
		}
		return []string{"second"}, &Response{}, nil
	}

	seq, iterErr := newPaginator(context.Background(), ListOptions{StartingAfter: "initial-cursor"}, fetch)

	for range seq {
	}
	require.NoError(t, iterErr())
	for range seq {
	}
	require.NoError(t, iterErr())

	assert.Equal(t, []ListOptions{
		{StartingAfter: "initial-cursor"},
		{StartingAfter: "next-cursor"},
		{StartingAfter: "initial-cursor"},
		{StartingAfter: "next-cursor"},
	}, gotOptions)
}

func TestPaginator_newPaginator_resetsErrorForNextSequentialIteration(t *testing.T) {
	fetchCount := 0
	fetchErr := errors.New("fetch page")
	fetch := func(_ context.Context, _ ListOptions) ([]string, *Response, error) {
		fetchCount++
		if fetchCount == 1 {
			return nil, nil, fetchErr
		}
		return []string{"recovered"}, &Response{}, nil
	}

	seq, iterErr := newPaginator(context.Background(), ListOptions{}, fetch)
	for range seq {
	}
	require.ErrorIs(t, iterErr(), fetchErr)

	var gotItems []string
	for item := range seq {
		gotItems = append(gotItems, item)
	}

	assert.Equal(t, []string{"recovered"}, gotItems)
	require.NoError(t, iterErr())
}

func TestPaginator_newPaginator_propagatesFetchError(t *testing.T) {
	fetchErr := errors.New("fetch page")
	seq, iterErr := newPaginator(context.Background(), ListOptions{}, func(context.Context, ListOptions) ([]string, *Response, error) {
		return nil, &Response{}, fetchErr
	})

	for range seq {
		t.Fatal("unexpected item")
	}

	require.ErrorIs(t, iterErr(), fetchErr)
}

func TestPaginator_newPaginator_rejectsMissingResponse(t *testing.T) {
	seq, iterErr := newPaginator(context.Background(), ListOptions{}, func(context.Context, ListOptions) ([]string, *Response, error) {
		return []string{"item"}, nil, nil
	})

	for range seq {
		t.Fatal("unexpected item")
	}

	assert.EqualError(t, iterErr(), "pagination response is missing")
}

func TestPaginator_newPaginator_rejectsInvalidNextCursor(t *testing.T) {
	tests := []struct {
		name           string
		initialOptions ListOptions
		next           string
		wantErr        string
	}{
		{
			name:    "malformed next link",
			next:    "://invalid",
			wantErr: "failed to extract starting_after query param: failed to parse pagination path:",
		},
		{
			name:    "missing cursor",
			next:    "/rest/items?limit=10",
			wantErr: "next-page link has no starting_after cursor",
		},
		{
			name:           "repeated current cursor",
			initialOptions: ListOptions{StartingAfter: "current-cursor"},
			next:           "/rest/items?starting_after=current-cursor",
			wantErr:        "next-page link repeats a previously seen starting_after cursor",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			seq, iterErr := newPaginator(context.Background(), tt.initialOptions, func(context.Context, ListOptions) ([]string, *Response, error) {
				return nil, &Response{Links: &PaginatedLinks{Next: tt.next}}, nil
			})

			for range seq {
			}

			assert.ErrorContains(t, iterErr(), tt.wantErr)
		})
	}
}

func TestPaginator_newPaginator_rejectsCursorCycle(t *testing.T) {
	fetchCount := 0
	fetch := func(_ context.Context, _ ListOptions) ([]string, *Response, error) {
		fetchCount++
		nextCursor := map[int]string{1: "cursor-a", 2: "cursor-b", 3: "cursor-a"}[fetchCount]
		return nil, &Response{Links: &PaginatedLinks{Next: "/rest/items?starting_after=" + nextCursor}}, nil
	}

	seq, iterErr := newPaginator(context.Background(), ListOptions{}, fetch)
	for range seq {
	}

	assert.Equal(t, 3, fetchCount)
	assert.EqualError(t, iterErr(), "next-page link repeats a previously seen starting_after cursor")
}

func TestPaginator_newPaginator_honorsCanceledContextBeforeFetch(t *testing.T) {
	canceledCtx, cancel := context.WithCancel(context.Background())
	cancel()
	fetchCalled := false
	seq, iterErr := newPaginator(canceledCtx, ListOptions{}, func(context.Context, ListOptions) ([]string, *Response, error) {
		fetchCalled = true
		return nil, &Response{}, nil
	})

	for range seq {
	}

	assert.False(t, fetchCalled)
	require.ErrorIs(t, iterErr(), context.Canceled)
}

func TestPaginator_newPaginator_stopsWhenConsumerStops(t *testing.T) {
	fetchCount := 0
	seq, iterErr := newPaginator(context.Background(), ListOptions{}, func(context.Context, ListOptions) ([]string, *Response, error) {
		fetchCount++
		return []string{"first", "second"}, &Response{Links: &PaginatedLinks{Next: "/rest/items?starting_after=next-cursor"}}, nil
	})

	var gotItems []string
	for item := range seq {
		gotItems = append(gotItems, item)
		break
	}

	assert.Equal(t, []string{"first"}, gotItems)
	assert.Equal(t, 1, fetchCount)
	require.NoError(t, iterErr())
}

func TestPaginator_extractStartingAfterQueryParam(t *testing.T) {
	tests := map[string]struct {
		path          string
		expectedToken string
		errorExpected bool
	}{
		"success-token-extraction": {
			path:          "/rest/orgs?limit=20&starting_after=v1.eyJuYW1&version=2024-10-15",
			expectedToken: "v1.eyJuYW1",
		},
		"empty-string-when-token-not-present": {
			path:          "/rest/orgs?limit=20&version=2024-10-15",
			expectedToken: "",
		},
		"empty-string-when-path-without-query-params": {
			path:          "/rest/orgs",
			expectedToken: "",
		},
		"error-malformed-url": {
			path:          "://a",
			errorExpected: true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			actualToken, err := extractStartingAfterQueryParam(test.path)

			assert.Equal(t, test.expectedToken, actualToken)
			if test.errorExpected {
				assert.Error(t, err)
			} else {
				require.NoError(t, err, fmt.Sprintf("extract cursor from %q", test.path))
			}
		})
	}
}

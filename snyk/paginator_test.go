package snyk

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_extractStartingAfterQueryParam(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		path          string
		expectedToken string
		errorExpected bool
	}{
		"success-token-extraction": {
			path:          "/rest/orgs?limit=20&starting_after=v1.eyJuYW1&version=2024-10-15",
			expectedToken: "v1.eyJuYW1",
			errorExpected: false,
		},
		"empty-string-when-token-not-present": {
			path:          "/rest/orgs?limit=20&version=2024-10-15",
			expectedToken: "",
			errorExpected: false,
		},
		"empty-string-when-path-without-query-params": {
			path:          "/rest/orgs",
			expectedToken: "",
			errorExpected: false,
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
				assert.NoError(t, err)
			}
		})
	}
}

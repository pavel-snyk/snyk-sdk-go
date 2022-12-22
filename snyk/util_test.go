package snyk

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func ptr[A any](a A) *A {
	return &a
}

func mustParseTime(t *testing.T, s string) time.Time {
	timestamp, err := time.Parse(time.RFC3339, s)
	assert.NoError(t, err)
	return timestamp
}

//nolint:testpackage
package version

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetVersion(t *testing.T) {
	t.Parallel()

	require.Contains(t, []string{"(devel)", "unknown-local"}, GetVersion())

	t.Run("returns explicitly set version when version variable is set", func(t *testing.T) {
		t.Parallel()

		// Save original value
		original := version
		defer func() { version = original }()

		// Set explicit version
		version = "v1.2.3"
		assert.Equal(t, "v1.2.3", GetVersion())
	})
}

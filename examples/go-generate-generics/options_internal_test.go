package gogenerate

import (
	"os"
	"path/filepath"
	"testing"

	optionsgen "github.com/kazhuravlev/options-gen/options-gen"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	fallbackValue = 42
	storedValue   = 7
)

func TestGenericOptions(t *testing.T) {
	t.Parallel()

	t.Run("defaults come from the generic function", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions[string, int]()
		assert.Zero(t, opts.defaultVal)
		require.NoError(t, opts.Validate())
	})

	t.Run("setter overrides the default", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions[string, int](WithDefaultVal[string, int](fallbackValue))
		assert.Equal(t, fallbackValue, opts.defaultVal)
	})

	t.Run("client falls back to the default value", func(t *testing.T) {
		t.Parallel()

		client, err := New(NewOptions[string, int](WithDefaultVal[string, int](fallbackValue)))
		require.NoError(t, err)

		assert.Equal(t, fallbackValue, client.Get("missing"))

		client.Set("key", storedValue)
		assert.Equal(t, storedValue, client.Get("key"))
	})
}

// TestGeneratedFileIsUpToDate regenerates the file with the same settings as the go:generate directive
// in options.go (-defaults-from=func), but into a temp dir, and compares it with the committed one.
func TestGeneratedFileIsUpToDate(t *testing.T) {
	t.Parallel()

	outFile := filepath.Join(t.TempDir(), "options_generated.go")

	err := optionsgen.Run(optionsgen.NewOptions(
		optionsgen.WithVersion("qa-version"),
		optionsgen.WithInFilename("./options.go"),
		optionsgen.WithOutFilename(outFile),
		optionsgen.WithStructName("Options"),
		optionsgen.WithPackageName("gogenerate"),
		optionsgen.WithDefaults(optionsgen.Defaults{From: optionsgen.DefaultsFromFunc, Param: ""}),
	))
	require.NoError(t, err)

	expected, err := os.ReadFile("./options_generated.go")
	require.NoError(t, err)

	actual, err := os.ReadFile(outFile)
	require.NoError(t, err)

	assert.Equal(t, string(expected), string(actual))
}

package gogenerate

import (
	"net/http"
	"os"
	"path/filepath"
	"testing"

	optionsgen "github.com/kazhuravlev/options-gen/options-gen"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewOptions(t *testing.T) {
	t.Parallel()

	t.Run("defaults from the tag and mandatory parameters", func(t *testing.T) {
		t.Parallel()

		client := new(http.Client)
		opts := NewOptions(client, "token")

		assert.Same(t, client, opts.httpClient)
		assert.Equal(t, "token", opts.token)
		assert.Equal(t, "127.0.0.1:8000", opts.addr)
		assert.Empty(t, opts.tls)
		require.NoError(t, opts.Validate())
	})

	t.Run("setters override the defaults", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(new(http.Client), "token", WithAddr("127.0.0.1:9000"), WithTLS("strict"))

		assert.Equal(t, "127.0.0.1:9000", opts.addr)
		assert.Equal(t, "strict", opts.tls)
	})

	t.Run("nil http client fails validation and the constructor", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(nil, "token")

		err := opts.Validate()
		require.Error(t, err)
		assert.ErrorContains(t, err, "(httpClient)")

		client, err := New(opts)
		require.ErrorContains(t, err, "bad configuration")
		assert.Nil(t, client)
	})

	t.Run("client is built from valid options", func(t *testing.T) {
		t.Parallel()

		client, err := New(NewOptions(new(http.Client), "token"))
		require.NoError(t, err)
		assert.NotNil(t, client)
	})
}

// TestGeneratedFileIsUpToDate regenerates the file with the same settings as the go:generate directive
// in options.go, but into a temp dir, and compares it with the committed one.
func TestGeneratedFileIsUpToDate(t *testing.T) {
	t.Parallel()

	outFile := filepath.Join(t.TempDir(), "options_generated.go")

	err := optionsgen.Run(optionsgen.NewOptions(
		optionsgen.WithVersion("qa-version"),
		optionsgen.WithInFilename("./options.go"),
		optionsgen.WithOutFilename(outFile),
		optionsgen.WithStructName("Options"),
		optionsgen.WithPackageName("gogenerate"),
		optionsgen.WithDefaults(optionsgen.Defaults{From: optionsgen.DefaultsFromTag, Param: "default"}),
	))
	require.NoError(t, err)

	expected, err := os.ReadFile("./options_generated.go")
	require.NoError(t, err)

	actual, err := os.ReadFile(outFile)
	require.NoError(t, err)

	assert.Equal(t, string(expected), string(actual))
}

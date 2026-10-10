package main

import (
	"os"
	"path/filepath"
	"testing"

	subpackage "github.com/kazhuravlev/options-gen/examples/library/sub-package"
	optionsgen "github.com/kazhuravlev/options-gen/options-gen"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	validPort    = 8080
	tooSmallPort = 9
)

func TestOptionsValidation(t *testing.T) {
	t.Parallel()

	service := new(subpackage.Service1)

	t.Run("valid options", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(service, "https://example.com", WithSomePort(validPort))
		require.NoError(t, opts.Validate())
	})

	t.Run("mandatory fields are validated", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(nil, "not a url", WithSomePort(validPort))

		err := opts.Validate()
		require.Error(t, err)
		assert.ErrorContains(t, err, "(service1)")
		assert.ErrorContains(t, err, "(s3Endpoint)")
		assert.NotContains(t, err.Error(), "(port)")
	})

	t.Run("optional port is required and bounded", func(t *testing.T) {
		t.Parallel()

		noPort := NewOptions(service, "https://example.com")

		err := noPort.Validate()
		require.Error(t, err)
		assert.ErrorContains(t, err, "(port)")

		smallPort := NewOptions(service, "https://example.com", WithSomePort(tooSmallPort))

		err = smallPort.Validate()
		require.Error(t, err)
		assert.ErrorContains(t, err, "(port)")
	})
}

func TestConfigAndParams(t *testing.T) {
	t.Parallel()

	t.Run("config has no validation rules", func(t *testing.T) {
		t.Parallel()

		emptyConfig := NewConfig()
		require.NoError(t, emptyConfig.Validate())

		namedConfig := NewConfig(WithSomeName("name"))
		require.NoError(t, namedConfig.Validate())
	})

	t.Run("params hash must be hexadecimal", func(t *testing.T) {
		t.Parallel()

		hexParams := NewParams("deadbeef")
		require.NoError(t, hexParams.Validate())

		badParams := NewParams("not-hex")

		err := badParams.Validate()
		require.Error(t, err)
		assert.ErrorContains(t, err, "(hash)")
	})
}

// TestGeneratedFilesAreUpToDate regenerates the files the same way main.go does, but into a temp dir,
// and compares them with the committed ones.
func TestGeneratedFilesAreUpToDate(t *testing.T) {
	t.Parallel()

	for _, params := range []struct {
		outFname   string
		structName string
	}{
		{outFname: "example_out_options.go", structName: "Options"},
		{outFname: "example_out_config.go", structName: "Config"},
		{outFname: "example_out_params.go", structName: "Params"},
	} {
		t.Run(params.structName, func(t *testing.T) {
			t.Parallel()

			outFile := filepath.Join(t.TempDir(), params.outFname)

			err := optionsgen.Run(optionsgen.NewOptions(
				optionsgen.WithVersion("qa-version"),
				optionsgen.WithInFilename("./example_in.go"),
				optionsgen.WithOutFilename(outFile),
				optionsgen.WithStructName(params.structName),
				optionsgen.WithPackageName("main"),
				optionsgen.WithOutPrefix("Some"),
				optionsgen.WithDefaults(optionsgen.Defaults{From: optionsgen.DefaultsFromTag, Param: ""}),
				optionsgen.WithShowWarnings(true),
				optionsgen.WithWithIsset(false),
				optionsgen.WithAllVariadic(false),
				optionsgen.WithConstructorTypeRender(optionsgen.ConstructorPublicRender),
				optionsgen.WithOutOptionTypeName(""),
			))
			require.NoError(t, err)

			expected, err := os.ReadFile(params.outFname)
			require.NoError(t, err)

			actual, err := os.ReadFile(outFile)
			require.NoError(t, err)

			assert.Equal(t, string(expected), string(actual))
		})
	}
}

package gogenerate

import (
	"os"
	"path/filepath"
	"testing"

	optionsgen "github.com/kazhuravlev/options-gen/options-gen"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const bigEnough = 10

// NOTE: the generated IsSet state is package-level (one array per prefix), so these tests are not parallel.
func TestTwoOptionsInOnePackage(t *testing.T) {
	t.Run("defaults come from the variables", func(t *testing.T) {
		opts1 := NewOptions1()
		assert.Equal(t, defaultOptions1, opts1)

		opts2 := NewOptions2()
		assert.Equal(t, defaultOptions2, opts2)
	})

	t.Run("prefixed setters do not collide", func(t *testing.T) {
		opts1 := NewOptions1(WithKKKField1(bigEnough))
		assert.Equal(t, bigEnough, opts1.field1)
		assert.Equal(t, defaultOptions1.field0, opts1.field0)

		opts2 := NewOptions2(WithNNNField1(bigEnough))
		assert.Equal(t, bigEnough, opts2.field1)
		assert.Equal(t, defaultOptions2.field4, opts2.field4)
	})

	t.Run("IsSet reports fields set by defaults and setters", func(t *testing.T) {
		opts1 := NewOptions1(WithKKKField2(bigEnough))
		assert.True(t, opts1.IsSet(FieldKKKfield0), "defaults from a variable mark the field as set")
		assert.True(t, opts1.IsSet(FieldKKKfield2))
	})

	t.Run("validation applies the min rule", func(t *testing.T) {
		opts1 := NewOptions1(
			WithKKKField0(bigEnough), WithKKKField1(bigEnough), WithKKKField2(bigEnough), WithKKKField3(bigEnough),
		)
		require.NoError(t, opts1.Validate())

		defaults := NewOptions1()

		err := defaults.Validate()
		require.Error(t, err)
		assert.ErrorContains(t, err, "(field0)")
		assert.NotContains(t, err.Error(), "(field3)")

		opts2 := NewOptions2(
			WithNNNField1(bigEnough), WithNNNField2(bigEnough), WithNNNField3(bigEnough), WithNNNField4(bigEnough),
		)
		client, err := New(opts1, opts2)
		require.NoError(t, err)
		assert.NotNil(t, client)
	})
}

// TestGeneratedFilesAreUpToDate regenerates both files with the same settings as the go:generate
// directives in options.go, but into a temp dir, and compares them with the committed ones.
func TestGeneratedFilesAreUpToDate(t *testing.T) {
	t.Parallel()

	for _, params := range []struct {
		structName string
		outPrefix  string
		outFname   string
		varName    string
	}{
		{structName: "Options1", outPrefix: "KKK", outFname: "options1_generated.go", varName: "defaultOptions1"},
		{structName: "Options2", outPrefix: "NNN", outFname: "options2_generated.go", varName: "defaultOptions2"},
	} {
		t.Run(params.structName, func(t *testing.T) {
			t.Parallel()

			outFile := filepath.Join(t.TempDir(), params.outFname)

			err := optionsgen.Run(optionsgen.NewOptions(
				optionsgen.WithVersion("qa-version"),
				optionsgen.WithInFilename("./options.go"),
				optionsgen.WithOutFilename(outFile),
				optionsgen.WithStructName(params.structName),
				optionsgen.WithPackageName("gogenerate"),
				optionsgen.WithOutPrefix(params.outPrefix),
				optionsgen.WithDefaults(optionsgen.Defaults{From: optionsgen.DefaultsFromVar, Param: params.varName}),
				optionsgen.WithWithIsset(true),
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

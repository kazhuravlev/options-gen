package optionsgen

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/kazhuravlev/options-gen/internal/ctype"
	pkgerrors "github.com/kazhuravlev/options-gen/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const simpleSource = "package test\n\ntype Options struct {\n\tfield string\n}\n"

// writeSource writes sourceCode into a fresh temp dir and returns the input and output file paths.
func writeSource(t *testing.T, sourceCode string) (inputFile, outputFile string) {
	t.Helper()

	tmpDir := t.TempDir()
	inputFile = filepath.Join(tmpDir, "options.go")
	outputFile = filepath.Join(tmpDir, "options_generated.go")

	require.NoError(t, os.WriteFile(inputFile, []byte(sourceCode), ctype.DefaultPermission))

	return inputFile, outputFile
}

func readGenerated(t *testing.T, filename string) string {
	t.Helper()

	content, err := os.ReadFile(filename)
	require.NoError(t, err)

	return string(content)
}

func validOptions(inputFile, outputFile string, extra ...OptOptionsSetter) Options {
	setters := []OptOptionsSetter{
		WithVersion("test"),
		WithPackageName("test"),
		WithStructName("Options"),
		WithInFilename(inputFile),
		WithOutFilename(outputFile),
	}

	return NewOptions(append(setters, extra...)...)
}

func TestRun_BadConfiguration(t *testing.T) {
	t.Parallel()

	t.Run("missing required options are aggregated into one error", func(t *testing.T) {
		t.Parallel()

		inputFile, outputFile := writeSource(t, simpleSource)

		err := Run(NewOptions(WithInFilename(inputFile), WithOutFilename(outputFile)))
		require.Error(t, err)
		assert.ErrorContains(t, err, "bad configuration: ValidationErrors: ")
		assert.ErrorContains(t, err, "(version): ")
		assert.ErrorContains(t, err, "(structName): ")
		assert.ErrorContains(t, err, "(packageName): ")

		var validationErrs pkgerrors.ValidationErrors
		require.ErrorAs(t, err, &validationErrs)
		assert.Len(t, validationErrs, 3)

		assert.NoFileExists(t, outputFile)
	})

	t.Run("invalid constructor type render", func(t *testing.T) {
		t.Parallel()

		inputFile, outputFile := writeSource(t, simpleSource)

		err := Run(validOptions(inputFile, outputFile, WithConstructorTypeRender("bogus")))
		require.Error(t, err)
		assert.ErrorContains(t, err, "bad configuration")
		assert.ErrorContains(t, err, "(constructorTypeRender)")
		assert.ErrorContains(t, err, "oneof")
		assert.NoFileExists(t, outputFile)
	})

	t.Run("empty constructor type render", func(t *testing.T) {
		t.Parallel()

		inputFile, outputFile := writeSource(t, simpleSource)

		err := Run(validOptions(inputFile, outputFile, WithConstructorTypeRender("")))
		require.Error(t, err)
		assert.ErrorContains(t, err, "(constructorTypeRender)")
		assert.NoFileExists(t, outputFile)
	})

	t.Run("zero defaults value is accepted and generates without defaults", func(t *testing.T) {
		t.Parallel()

		// Documented corner: `validate:"required"` on the `defaults` struct field is a no-op, because
		// go-playground/validator ignores `required` on struct values unless WithRequiredStructEnabled()
		// is set. The zero value therefore behaves like an unknown source: no defaults at all.
		inputFile, outputFile := writeSource(t, simpleSource)

		err := Run(validOptions(inputFile, outputFile, WithDefaults(Defaults{From: "", Param: ""})))
		require.NoError(t, err)

		generated := readGenerated(t, outputFile)
		assert.Contains(t, generated, "func WithField(opt string) OptOptionsSetter")
		assert.NotContains(t, generated, "Setting defaults from")
	})

	t.Run("unknown defaults source passes validation and generates without defaults", func(t *testing.T) {
		t.Parallel()

		inputFile, outputFile := writeSource(t, simpleSource)

		err := Run(validOptions(inputFile, outputFile, WithDefaults(Defaults{From: "bogus", Param: "x"})))
		require.NoError(t, err)

		generated := readGenerated(t, outputFile)
		assert.Contains(t, generated, "func WithField(opt string) OptOptionsSetter")
		assert.NotContains(t, generated, "Setting defaults from")
	})
}

func TestRun_WarningsHandler(t *testing.T) {
	t.Parallel()

	const publicFieldSource = "package test\n\ntype Options struct {\n\tPublicField string\n}\n"

	t.Run("nil handler falls back to the default handler instead of panicking", func(t *testing.T) {
		t.Parallel()

		inputFile, outputFile := writeSource(t, publicFieldSource)

		var err error
		require.NotPanics(t, func() {
			err = Run(validOptions(inputFile, outputFile,
				WithShowWarnings(true), WithIgnoreErrors(true), WithWarningsHandler(nil)))
		})
		require.NoError(t, err)
		assert.FileExists(t, outputFile)
	})

	t.Run("warnings are not emitted when they are muted", func(t *testing.T) {
		t.Parallel()

		inputFile, outputFile := writeSource(t, publicFieldSource)

		var warnings []string
		handler := func(msg string) { warnings = append(warnings, msg) }

		err := Run(validOptions(inputFile, outputFile,
			WithShowWarnings(false), WithIgnoreErrors(true), WithWarningsHandler(handler)))
		require.NoError(t, err)
		assert.Empty(t, warnings)
	})

	t.Run("handler is not called when there are no warnings", func(t *testing.T) {
		t.Parallel()

		inputFile, outputFile := writeSource(t, simpleSource)

		var warnings []string
		handler := func(msg string) { warnings = append(warnings, msg) }

		err := Run(validOptions(inputFile, outputFile, WithShowWarnings(true), WithWarningsHandler(handler)))
		require.NoError(t, err)
		assert.Empty(t, warnings)
	})
}

func TestRun_WriteFailures(t *testing.T) {
	t.Parallel()

	t.Run("output directory does not exist", func(t *testing.T) {
		t.Parallel()

		inputFile, _ := writeSource(t, simpleSource)
		outputFile := filepath.Join(t.TempDir(), "missing", "options_generated.go")

		err := Run(validOptions(inputFile, outputFile))
		require.Error(t, err)
		assert.ErrorContains(t, err, "cannot write result")
		assert.ErrorIs(t, err, fs.ErrNotExist)
		assert.NoFileExists(t, outputFile)
	})

	t.Run("output path is a directory", func(t *testing.T) {
		t.Parallel()

		inputFile, _ := writeSource(t, simpleSource)

		err := Run(validOptions(inputFile, t.TempDir()))
		require.Error(t, err)
		assert.ErrorContains(t, err, "cannot write result")
	})

	t.Run("input file does not exist", func(t *testing.T) {
		t.Parallel()

		inputFile := filepath.Join(t.TempDir(), "nope.go")
		outputFile := filepath.Join(t.TempDir(), "options_generated.go")

		err := Run(validOptions(inputFile, outputFile))
		require.Error(t, err)
		assert.ErrorContains(t, err, "cannot get options spec")
		assert.NoFileExists(t, outputFile)
	})
}

func TestRun_ExcludeMatchesTitleCasedName(t *testing.T) {
	t.Parallel()

	const sourceCode = `package test

type Options struct {
	addr         string
	internalConn string
	debugMode    bool
	logLevel     string
}
`

	tests := []struct {
		name        string
		exclude     []string
		wantPresent []string
		wantAbsent  []string
	}{
		{
			name:        "lowercase pattern does not match the Title-cased option name",
			exclude:     []string{"internal*", "debug*"},
			wantPresent: []string{"WithAddr(", "WithInternalConn(", "WithDebugMode(", "WithLogLevel("},
			wantAbsent:  nil,
		},
		{
			name:        "Title-cased pattern excludes the option",
			exclude:     []string{"^Internal", "^Debug"},
			wantPresent: []string{"WithAddr(", "WithLogLevel("},
			wantAbsent:  []string{"WithInternalConn(", "WithDebugMode("},
		},
		{
			name:        "case-insensitive pattern matches regardless of spelling",
			exclude:     []string{"(?i)^internal"},
			wantPresent: []string{"WithAddr(", "WithDebugMode(", "WithLogLevel("},
			wantAbsent:  []string{"WithInternalConn("},
		},
		{
			name:        "pattern matching everything leaves only the constructor",
			exclude:     []string{".*"},
			wantPresent: []string{"func NewOptions("},
			wantAbsent:  []string{"WithAddr(", "WithInternalConn(", "WithDebugMode(", "WithLogLevel("},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			patterns := make([]*regexp.Regexp, 0, len(tt.exclude))
			for _, pattern := range tt.exclude {
				patterns = append(patterns, regexp.MustCompile(pattern))
			}

			inputFile, outputFile := writeSource(t, sourceCode)

			err := Run(validOptions(inputFile, outputFile, WithExclude(patterns...)))
			require.NoError(t, err)

			generated := readGenerated(t, outputFile)
			for _, want := range tt.wantPresent {
				assert.Contains(t, generated, want)
			}

			for _, unwanted := range tt.wantAbsent {
				assert.NotContains(t, generated, unwanted)
			}
		})
	}
}

func TestRun_AllVariadic(t *testing.T) {
	t.Parallel()

	const sourceCode = "package test\n\ntype Options struct {\n" +
		"\tcount    int\n" +
		"\ttags     []string\n" +
		"\texplicit []int `option:\"variadic=false\"`\n" +
		"\tids      []int `option:\"mandatory\"`\n" +
		"}\n"

	t.Run("all-variadic applies only to optional slice fields", func(t *testing.T) {
		t.Parallel()

		inputFile, outputFile := writeSource(t, sourceCode)

		err := Run(validOptions(inputFile, outputFile, WithAllVariadic(true)))
		require.NoError(t, err)

		generated := readGenerated(t, outputFile)
		// A non-slice field is silently left as is.
		assert.Contains(t, generated, "func WithCount(opt int) OptOptionsSetter")
		// A plain slice field becomes variadic.
		assert.Contains(t, generated, "func WithTags(opt ...string) OptOptionsSetter")
		// An explicit `variadic=false` tag wins over the global flag.
		assert.Contains(t, generated, "func WithExplicit(opt []int) OptOptionsSetter")
		// A mandatory field stays a plain constructor parameter.
		assert.Contains(t, generated, "ids []int,")
		assert.NotContains(t, generated, "opt ...int")
	})

	t.Run("without all-variadic slices stay slices", func(t *testing.T) {
		t.Parallel()

		inputFile, outputFile := writeSource(t, sourceCode)

		err := Run(validOptions(inputFile, outputFile, WithAllVariadic(false)))
		require.NoError(t, err)

		generated := readGenerated(t, outputFile)
		assert.Contains(t, generated, "func WithTags(opt []string) OptOptionsSetter")
		assert.Contains(t, generated, "func WithExplicit(opt []int) OptOptionsSetter")
		assert.NotContains(t, generated, "...string")
	})
}

func TestResolveDefaults_UnknownSource(t *testing.T) {
	t.Parallel()

	tagName, varName, funcName := resolveDefaults(Defaults{From: "bogus", Param: "whatever"}, "Options")
	assert.Empty(t, tagName)
	assert.Empty(t, varName)
	assert.Empty(t, funcName)
}

func TestResolveOutOptionTypeName_Pattern(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{name: "letters only", input: "OptSetter", wantErr: false},
		{name: "lowercase only", input: "opt", wantErr: false},
		{name: "uppercase only", input: "OPT", wantErr: false},
		{name: "digit", input: "Opt2", wantErr: true},
		{name: "underscore", input: "Opt_Setter", wantErr: true},
		{name: "dash", input: "Opt-Setter", wantErr: true},
		{name: "space", input: "Opt Setter", wantErr: true},
		{name: "non-ascii letter", input: "Ünicode", wantErr: true},
		{name: "whitespace only", input: " ", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := resolveOutOptionTypeName("Options", tt.input)
			if tt.wantErr {
				require.Error(t, err)
				assert.Empty(t, got)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.input, got)
		})
	}
}

package optionsgen

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)


// TestRun_ErrorCases tests error handling in various scenarios
func TestRun_ErrorCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		sourceCode string
		opts       Options
		wantErr    bool
		errSubstr  string
	}{
		{
			name: "mandatory field with default value",
			sourceCode: `package test
type Options struct {
	Field string ` + "`option:\"mandatory\" default:\"value\"`" + `
}`,
			opts: Options{
				version:     "test",
				packageName: "test",
				structName:  "Options",
				defaults: Defaults{
					From:  DefaultsFromTag,
					Param: "default",
				},
				constructorTypeRender: ConstructorPublicRender,
			},
			wantErr:   true,
			errSubstr: "mandatory option cannot have a default value",
		},
		{
			name: "invalid default value type",
			sourceCode: `package test
type Options struct {
	Field int ` + "`default:\"not_a_number\"`" + `
}`,
			opts: Options{
				version:     "test",
				packageName: "test",
				structName:  "Options",
				defaults: Defaults{
					From:  DefaultsFromTag,
					Param: "default",
				},
				constructorTypeRender: ConstructorPublicRender,
			},
			wantErr:   true,
			errSubstr: "invalid",
		},
		{
			name: "struct not found",
			sourceCode: `package test
type OtherStruct struct {
	Field string
}`,
			opts: Options{
				version:     "test",
				packageName: "test",
				structName:  "Options",
				defaults: Defaults{
					From: DefaultsFromNone,
				},
				constructorTypeRender: ConstructorPublicRender,
			},
			wantErr:   true,
			errSubstr: "cannot find target struct",
		},
		{
			name: "invalid option type name",
			sourceCode: `package test
type Options struct {
	Field string
}`,
			opts: Options{
				version:           "test",
				packageName:       "test",
				structName:        "Options",
				outOptionTypeName: "Invalid-Name-123",
				defaults: Defaults{
					From: DefaultsFromNone,
				},
				constructorTypeRender: ConstructorPublicRender,
			},
			wantErr:   true,
			errSubstr: "must be a valid type name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			inputFile := filepath.Join(tmpDir, "options.go")
			outputFile := filepath.Join(tmpDir, "options_generated.go")

			err := os.WriteFile(inputFile, []byte(tt.sourceCode), 0o644)
			require.NoError(t, err)

			tt.opts.inFilename = inputFile
			tt.opts.outFilename = outputFile

			if err := Run(tt.opts); tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			if tt.wantErr && err != nil && tt.errSubstr != "" {
				require.ErrorContains(t, err, tt.errSubstr)
			}
		})
	}
}

// TestRun_WarningGeneration tests that warnings are properly generated
func TestRun_WarningGeneration(t *testing.T) {
	t.Parallel()

	// Public field should generate warning
	sourceCode := `package test
type Options struct {
	PublicField string
}`

	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "options.go")
	outputFile := filepath.Join(tmpDir, "options_generated.go")

	err := os.WriteFile(inputFile, []byte(sourceCode), 0o644)
	require.NoError(t, err)

	var warnings []string
	handler := func(w string) {
		warnings = append(warnings, w)
	}

	opts := NewOptions(
		WithVersion("test"),
		WithPackageName("test"),
		WithStructName("Options"),
		WithInFilename(inputFile),
		WithOutFilename(outputFile),
		WithShowWarnings(true),
		WithWarningsHandler(handler),
	)

	require.NoError(t, Run(opts))
	require.Equal(t, []string{
		"Warning: consider to make `PublicField` is private. This is will not allow to users to avoid constructor method.",
	}, warnings)
}

// TestDefaultsFrom_AllModes tests all defaults modes
func TestDefaultsFrom_AllModes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		defaultsFrom DefaultsFrom
		param        string
		sourceCode   string
		wantErr      bool
	}{
		{
			name:         "defaults from tag",
			defaultsFrom: DefaultsFromTag,
			param:        "default",
			sourceCode: `package test
type Options struct {
	Field string ` + "`default:\"value\"`" + `
}`,
			wantErr: false,
		},
		{
			name:         "defaults from var",
			defaultsFrom: DefaultsFromVar,
			param:        "defaultOpts",
			sourceCode: `package test
var defaultOpts = Options{}
type Options struct {
	Field string
}`,
			wantErr: false,
		},
		{
			name:         "defaults from func",
			defaultsFrom: DefaultsFromFunc,
			param:        "getDefaults",
			sourceCode: `package test
func getDefaults() Options { return Options{} }
type Options struct {
	Field string
}`,
			wantErr: false,
		},
		{
			name:         "defaults none",
			defaultsFrom: DefaultsFromNone,
			param:        "",
			sourceCode: `package test
type Options struct {
	Field string
}`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			inputFile := filepath.Join(tmpDir, "options.go")
			outputFile := filepath.Join(tmpDir, "options_generated.go")

			err := os.WriteFile(inputFile, []byte(tt.sourceCode), 0o644)
			require.NoError(t, err)

			opts := NewOptions(
				WithVersion("test"),
				WithPackageName("test"),
				WithStructName("Options"),
				WithInFilename(inputFile),
				WithOutFilename(outputFile),
				WithDefaults(Defaults{
					From:  tt.defaultsFrom,
					Param: tt.param,
				}),
			)
			if err := Run(opts); tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestRun_OutputFilePermissions tests that output file has correct permissions
func TestRun_OutputFilePermissions(t *testing.T) {
	t.Parallel()

	sourceCode := `package test
type Options struct {
	field string
}`

	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "options.go")
	outputFile := filepath.Join(tmpDir, "options_generated.go")

	err := os.WriteFile(inputFile, []byte(sourceCode), 0o644)
	require.NoError(t, err)

	err2 := Run(NewOptions(
		WithVersion("test"),
		WithPackageName("test"),
		WithStructName("Options"),
		WithInFilename(inputFile),
		WithOutFilename(outputFile),
	))
	require.NoError(t, err2)

	info, err := os.Stat(outputFile)
	require.NoError(t, err)
	require.Equal(t, os.FileMode(0o644), info.Mode().Perm())
}

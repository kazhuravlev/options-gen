//nolint:exhaustruct
package generator //nolint:testpackage

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
)

// writeOptionsModule writes a go.mod and the given files (paths relative to
// the module root) into a fresh temporary module and returns its root.
func writeOptionsModule(t *testing.T, files map[string]string) string {
	t.Helper()

	dir := t.TempDir()
	writeTestFile(t, filepath.Join(dir, "go.mod"), "module example.com/corner\n\ngo 1.25\n")
	for name, content := range files {
		writeTestFile(t, filepath.Join(dir, filepath.FromSlash(name)), content)
	}

	return dir
}

// buildGoModule compiles every package of the module in dir with the local toolchain.
func buildGoModule(t *testing.T, dir string) {
	t.Helper()

	ctx, cancel := context.WithTimeout(t.Context(), compileGeneratedPackageTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "go", "build", "./...")
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GOWORK=off", "GOTOOLCHAIN=local")
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, string(out))
}

// generateInto runs the whole pipeline for the Options struct in inFile and
// writes the generated code next to it.
func generateInto(t *testing.T, inFile, pkgName string, allVariadic bool) string {
	t.Helper()

	spec, err := GetOptionSpec(inFile, "Options", "default", allVariadic, nil)
	require.NoError(t, err)

	rendered, err := Render(NewOptions(
		WithVersion("test"),
		WithPackageName(pkgName),
		WithOptionsStructName("Options"),
		WithFileImports(spec.Imports),
		WithSpec(&spec.Spec),
		WithTagName("default"),
		WithConstructorTypeRender("public"),
		WithOptionTypeName("OptOptionsSetter"),
	))
	require.NoError(t, err)

	writeTestFile(t, filepath.Join(filepath.Dir(inFile), "options_generated.go"), string(rendered))

	return string(rendered)
}

func TestGetOptionSpec_TagCornerCases(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		source      string
		allVariadic bool
		wantErr     string
		check       func(t *testing.T, res *GetOptionSpecRes)
	}{
		{
			name: "mandatory_and_variadic",
			source: "package corner\n\ntype Options struct {\n" +
				"\titems []int `option:\"mandatory,variadic=true\"`\n}\n",
			wantErr: "field `items`: this field is mandatory and could not be variadic",
		},
		{
			name:        "mandatory_stays_plain_with_all_variadic",
			source:      "package corner\n\ntype Options struct {\n\titems []int `option:\"mandatory\"`\n}\n",
			allVariadic: true,
			check: func(t *testing.T, res *GetOptionSpecRes) {
				t.Helper()

				require.Len(t, res.Spec.Options, 1)
				require.Equal(t, "[]int", res.Spec.Options[0].Type)
				require.True(t, res.Spec.Options[0].TagOption.IsRequired)
				require.False(t, res.Spec.Options[0].TagOption.Variadic)
			},
		},
		{
			name:    "mandatory_with_default",
			source:  "package corner\n\ntype Options struct {\n\tport int `option:\"mandatory\" default:\"80\"`\n}\n",
			wantErr: "field `port`: mandatory option cannot have a default value",
		},
		{
			name:   "invalid_default_names_field_and_tag",
			source: "package corner\n\ntype Options struct {\n\tport int8 `default:\"300\"`\n}\n",
			wantErr: "field `port`: invalid `default` tag value: " +
				"bad default value strconv.ParseInt: parsing \"300\": value out of range 300",
		},
		{
			name:   "skipped_public_field_produces_no_warning",
			source: "package corner\n\ntype Options struct {\n\tPublic string `option:\"-\"`\n}\n",
			check: func(t *testing.T, res *GetOptionSpecRes) {
				t.Helper()

				require.Empty(t, res.Spec.Options)
				require.Empty(t, res.Warnings)
			},
		},
		{
			// Documents the current behaviour: only the first name of a multi-name field is used.
			name:   "multi_name_field_uses_the_first_name_only",
			source: "package corner\n\ntype Options struct {\n\tfirst, second string\n}\n",
			check: func(t *testing.T, res *GetOptionSpecRes) {
				t.Helper()

				require.Len(t, res.Spec.Options, 1)
				require.Equal(t, "first", res.Spec.Options[0].Field)
			},
		},
		{
			name:   "interpreted_string_literal_tag",
			source: "package corner\n\ntype Options struct {\n\tname string \"option:\\\"mandatory\\\"\"\n}\n",
			check: func(t *testing.T, res *GetOptionSpecRes) {
				t.Helper()

				require.Len(t, res.Spec.Options, 1)
				require.True(t, res.Spec.Options[0].TagOption.IsRequired)
			},
		},
		{
			name:        "type_parameter_field_with_all_variadic",
			source:      "package corner\n\ntype Options[T any] struct {\n\titem T\n\titems []T\n}\n",
			allVariadic: true,
			check: func(t *testing.T, res *GetOptionSpecRes) {
				t.Helper()

				require.Len(t, res.Spec.Options, 2)
				require.Equal(t, "T", res.Spec.Options[0].Type)
				require.False(t, res.Spec.Options[0].TagOption.Variadic)
				require.Equal(t, "T", res.Spec.Options[1].Type)
				require.True(t, res.Spec.Options[1].TagOption.Variadic)
			},
		},
		{
			name:    "fixed_size_array_cannot_be_variadic",
			source:  "package corner\n\ntype Options struct {\n\tarr [3]int `option:\"variadic=true\"`\n}\n",
			wantErr: "field `arr`: this type could not be variadic: it is not slice",
		},
		{
			name:        "fixed_size_array_stays_plain_with_all_variadic",
			source:      "package corner\n\ntype Options struct {\n\tarr [3]int\n}\n",
			allVariadic: true,
			check: func(t *testing.T, res *GetOptionSpecRes) {
				t.Helper()

				require.Len(t, res.Spec.Options, 1)
				require.Equal(t, "[3]int", res.Spec.Options[0].Type)
				require.False(t, res.Spec.Options[0].TagOption.Variadic)
			},
		},
		{
			name: "embedded_instantiated_generic_type",
			source: "package corner\n\ntype Box[T any] struct{ V T }\n\n" +
				"type Options struct {\n\tBox[int]\n\t*Box[string] `option:\"name=Ptr\"`\n}\n",
			check: func(t *testing.T, res *GetOptionSpecRes) {
				t.Helper()

				require.Len(t, res.Spec.Options, 2)
				require.Equal(t, "Box", res.Spec.Options[0].Name)
				require.Equal(t, "Box", res.Spec.Options[0].Field)
				require.Equal(t, "Box[int]", res.Spec.Options[0].Type)
				require.Equal(t, "Box", res.Spec.Options[1].Field)
				require.Equal(t, "*Box[string]", res.Spec.Options[1].Type)
			},
		},
		{
			name: "crlf_line_endings",
			source: "package corner\r\n\r\ntype Options struct {\r\n" +
				"\t// Port of the server.\r\n\tport int `default:\"80\"`\r\n}\r\n",
			check: func(t *testing.T, res *GetOptionSpecRes) {
				t.Helper()

				require.Len(t, res.Spec.Options, 1)
				require.Equal(t, "// Port of the server.", res.Spec.Options[0].Docstring)
				require.Equal(t, "80", res.Spec.Options[0].TagOption.Default)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			dir := writeOptionsModule(t, map[string]string{"options.go": testCase.source})

			res, err := GetOptionSpec(filepath.Join(dir, "options.go"), "Options", "default", testCase.allVariadic, nil)
			if testCase.wantErr != "" {
				require.EqualError(t, err, testCase.wantErr)

				return
			}

			require.NoError(t, err)
			testCase.check(t, res)
		})
	}
}

func TestGetOptionSpec_DefaultsTagName(t *testing.T) {
	t.Parallel()

	dir := writeOptionsModule(t, map[string]string{
		"options.go": "package corner\n\ntype Options struct {\n\tname string `default:\"x\" validate:\"required\"`\n}\n",
	})
	inFile := filepath.Join(dir, "options.go")

	// Defaults from var/func pass an empty tag name: the `default` tag is ignored.
	res, err := GetOptionSpec(inFile, "Options", "", false, nil)
	require.NoError(t, err)
	require.Empty(t, res.Spec.Options[0].TagOption.Default)
	require.Equal(t, "required", res.Spec.Options[0].TagOption.GoValidator)

	// The tag name may collide with the validate tag: both values are read from it.
	res, err = GetOptionSpec(inFile, "Options", "validate", false, nil)
	require.NoError(t, err)
	require.Equal(t, "required", res.Spec.Options[0].TagOption.GoValidator)
	require.Equal(t, "required", res.Spec.Options[0].TagOption.Default)
}

func TestGetOptionSpec_ExcludesMatchTitleCasedName(t *testing.T) {
	t.Parallel()

	dir := writeOptionsModule(t, map[string]string{
		"options.go": "package corner\n\ntype Options struct {\n\tfield string\n\tother string\n}\n",
	})
	inFile := filepath.Join(dir, "options.go")

	res, err := GetOptionSpec(inFile, "Options", "default", false, []*regexp.Regexp{regexp.MustCompile("^field$")})
	require.NoError(t, err)
	require.Len(t, res.Spec.Options, 2, "patterns are matched against the title-cased name, not the field")

	res, err = GetOptionSpec(inFile, "Options", "default", false, []*regexp.Regexp{regexp.MustCompile("^Field$")})
	require.NoError(t, err)
	require.Len(t, res.Spec.Options, 1)
	require.Equal(t, "other", res.Spec.Options[0].Field)
}

func TestGetOptionSpec_DirectoryLayout(t *testing.T) {
	t.Parallel()

	t.Run("struct_in_a_test_file", func(t *testing.T) {
		t.Parallel()

		dir := writeOptionsModule(t, map[string]string{
			"options.go":      "package corner\n",
			"options_test.go": "package corner\n\ntype Options struct {\n\tport int\n}\n",
		})

		res, err := GetOptionSpec(filepath.Join(dir, "options.go"), "Options", "default", false, nil)
		require.NoError(t, err)
		require.Len(t, res.Spec.Options, 1)
	})

	t.Run("imports_come_from_the_file_that_declares_the_struct", func(t *testing.T) {
		t.Parallel()

		dir := writeOptionsModule(t, map[string]string{
			"options.go": "package corner\n\nimport \"fmt\"\n\nvar _ = fmt.Sprint\n",
			"other.go":   "package corner\n\nimport \"time\"\n\ntype Options struct {\n\ttimeout time.Duration\n}\n",
		})

		res, err := GetOptionSpec(filepath.Join(dir, "options.go"), "Options", "default", false, nil)
		require.NoError(t, err)
		require.Equal(t, []Import{{Path: `"time"`, Alias: nil}}, res.Imports)
	})

	t.Run("sibling_file_with_syntax_error_is_ignored_when_the_struct_is_in_the_given_file", func(t *testing.T) {
		t.Parallel()

		dir := writeOptionsModule(t, map[string]string{
			"options.go": "package corner\n\ntype Options struct {\n\tport int\n}\n",
			"broken.go":  "package corner\n\nfunc broken( {\n",
		})

		res, err := GetOptionSpec(filepath.Join(dir, "options.go"), "Options", "default", false, nil)
		require.NoError(t, err)
		require.Len(t, res.Spec.Options, 1)
	})

	t.Run("sibling_file_with_syntax_error_fails_when_the_directory_must_be_parsed", func(t *testing.T) {
		t.Parallel()

		dir := writeOptionsModule(t, map[string]string{
			"options.go": "package corner\n",
			"other.go":   "package corner\n\ntype Options struct {\n\tport int\n}\n",
			"broken.go":  "package corner\n\nfunc broken( {\n",
		})

		_, err := GetOptionSpec(filepath.Join(dir, "options.go"), "Options", "default", false, nil)
		require.ErrorContains(t, err, "cannot parse file")
	})

	t.Run("path_is_a_directory", func(t *testing.T) {
		t.Parallel()

		dir := writeOptionsModule(t, map[string]string{
			"sub/options.go": "package sub\n\ntype Options struct {\n\tport int\n}\n",
		})

		_, err := GetOptionSpec(filepath.Join(dir, "sub"), "Options", "default", false, nil)
		require.ErrorContains(t, err, "is a directory")
	})

	t.Run("alias_of_an_unknown_package_is_skipped", func(t *testing.T) {
		t.Parallel()

		dir := writeOptionsModule(t, map[string]string{
			"options.go": "package corner\n\ntype Options unknown.Options\n",
		})

		_, err := GetOptionSpec(filepath.Join(dir, "options.go"), "Options", "default", false, nil)
		require.EqualError(t, err, "cannot find target struct: cannot find target struct")
	})

	t.Run("alias_of_a_package_that_cannot_be_loaded", func(t *testing.T) {
		t.Parallel()

		dir := writeOptionsModule(t, map[string]string{
			"options.go": "package corner\n\nimport \"example.com/elsewhere/pkg\"\n\ntype Options pkg.Options\n",
		})

		_, err := GetOptionSpec(filepath.Join(dir, "options.go"), "Options", "default", false, nil)
		require.ErrorContains(t, err, "cannot find target struct: package contains errors")
	})
}

func TestGetOptionSpec_ImportedSliceAliasAllVariadic(t *testing.T) {
	t.Parallel()

	dir := writeOptionsModule(t, map[string]string{
		"other/other.go": "package other\n\ntype Ints = []int\n\ntype Users []*User\n\ntype User struct{}\n",
		"options.go": "package corner\n\nimport \"example.com/corner/other\"\n\n" +
			"type Options struct {\n\tints  other.Ints\n\tusers other.Users\n}\n",
	})

	res, err := GetOptionSpec(filepath.Join(dir, "options.go"), "Options", "default", true, nil)
	require.NoError(t, err)
	require.Len(t, res.Spec.Options, 2)
	require.Equal(t, "int", res.Spec.Options[0].Type)
	require.True(t, res.Spec.Options[0].TagOption.Variadic)
	require.Equal(t, "*other.User", res.Spec.Options[1].Type)
	require.True(t, res.Spec.Options[1].TagOption.Variadic)

	generateInto(t, filepath.Join(dir, "options.go"), "corner", true)
	buildGoModule(t, dir)
}

func TestGetOptionSpec_StringDefaultWithSpecialCharactersCompiles(t *testing.T) {
	t.Parallel()

	dir := writeOptionsModule(t, map[string]string{
		"options.go": "package corner\n\ntype Options struct {\n" +
			"\tgreeting string `default:\"say \\\"hi\\\"\\\\n\"`\n" +
			"\tunicode  string `default:\"héllo 世界\"`\n}\n",
	})

	res, err := GetOptionSpec(filepath.Join(dir, "options.go"), "Options", "default", false, nil)
	require.NoError(t, err)
	require.Equal(t, `say "hi"\n`, res.Spec.Options[0].TagOption.Default)

	rendered := generateInto(t, filepath.Join(dir, "options.go"), "corner", false)
	require.Contains(t, rendered, `o.greeting = "say \"hi\"\\n"`)
	require.Contains(t, rendered, `o.unicode = "héllo 世界"`)

	buildGoModule(t, dir)
}

func TestGetOptionSpec_EmbeddedGenericCompiles(t *testing.T) {
	t.Parallel()

	dir := writeOptionsModule(t, map[string]string{
		"options.go": "package corner\n\ntype Box[T any] struct{ V T }\n\ntype Options struct {\n\tBox[int]\n}\n",
	})

	rendered := generateInto(t, filepath.Join(dir, "options.go"), "corner", false)
	require.Contains(t, rendered, "func WithBox(opt Box[int]) OptOptionsSetter")
	require.Contains(t, rendered, "o.Box = opt")

	buildGoModule(t, dir)
}

func TestGetOptionSpec_ImportedStructWithGenericFieldsCompiles(t *testing.T) {
	t.Parallel()

	dir := writeOptionsModule(t, map[string]string{
		"optionspkg/options.go": `package optionspkg

type Local struct{}

type Box[T any] struct{ V T }

type Pair[K comparable, V any] struct {
	Key K
	Val V
}

type Options struct {
	Boxed  Box[Local]
	Pairs  Pair[string, Local]
	Nested Box[[]*Local]
	Fn     func(...Local) (Local, error)
	Paren  (Local)
	Plain  string
}
`,
		"consumer/options.go": "package consumer\n\nimport alias \"example.com/corner/optionspkg\"\n\n" +
			"type Options alias.Options\n",
	})

	inFile := filepath.Join(dir, "consumer", "options.go")
	res, err := GetOptionSpec(inFile, "Options", "default", false, nil)
	require.NoError(t, err)

	gotTypes := make([]string, 0, len(res.Spec.Options))
	for _, option := range res.Spec.Options {
		gotTypes = append(gotTypes, option.Type)
	}

	require.Equal(t, []string{
		"alias.Box[alias.Local]",
		"alias.Pair[string, alias.Local]",
		"alias.Box[[]*alias.Local]",
		"func(...alias.Local) (alias.Local, error)",
		"(alias.Local)",
		"string",
	}, gotTypes)

	generateInto(t, inFile, "consumer", false)
	buildGoModule(t, dir)
}

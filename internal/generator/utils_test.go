//nolint:exhaustruct
package generator //nolint:testpackage

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/kazhuravlev/options-gen/internal/ctype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	benchmarkFormatCommentSink        string
	benchmarkApplyExcludesSink        []OptionMeta
	benchmarkFindImportPathPathSink   string
	benchmarkFindImportPathAliasSink  string
	benchmarkTypeParamsStrSpecSink    string
	benchmarkTypeParamsStrNamesSink   string
	benchmarkParseTagOptionSink       TagOption
	benchmarkParseTagWarningsSink     []string
	benchmarkImportPathBaseSink       string
	benchmarkExtractSliceElemTypeSink string
)

func Test_checkDefaultValue_Negative(t *testing.T) {
	cases := []struct {
		t   string
		val string
	}{
		{t: "int", val: "a"},
		{t: "int8", val: "b"},
		{t: "int16", val: "c"},
		{t: "int32", val: "d"},
		{t: "int64", val: "e"},

		{t: "uint", val: "aa"},
		{t: "uint8", val: "bb"},
		{t: "uint16", val: "cc"},
		{t: "uint32", val: "dd"},
		{t: "uint64", val: "ee"},

		{t: "float32", val: "aaa"},
		{t: "float64", val: "bbb"},

		{t: "bool", val: "a"},
		{t: "bool", val: "1"},
		{t: "bool", val: "t"},
		{t: "bool", val: "T"},
		{t: "bool", val: "TRUE"},
		{t: "bool", val: "True"},
		{t: "bool", val: "0"},
		{t: "bool", val: "f"},
		{t: "bool", val: "F"},
		{t: "bool", val: "FALSE"},
		{t: "bool", val: "False"},

		{t: "time.Duration", val: "1year"},

		{t: "fmt.Stringer", val: "nil"},
		{t: "Number", val: "nil"},
		{t: "localIterface", val: "nil"},
		{t: "*T", val: "nil"},
	}

	for _, tt := range cases {
		t.Run(tt.t, func(t *testing.T) {
			err := checkDefaultValue(tt.t, tt.val)
			assert.Error(t, err)
		})
	}
}

func Test_checkDefaultValue(t *testing.T) {
	cases := []struct {
		t        string
		val      string
		expected string
	}{
		{t: "int", val: "1", expected: "1"},
		{t: "int", val: "-1", expected: "-1"},
		{t: "int8", val: "1", expected: "1"},
		{t: "int8", val: "-1", expected: "-1"},
		{t: "int16", val: "1", expected: "1"},
		{t: "int16", val: "-1", expected: "-1"},
		{t: "int32", val: "1", expected: "1"},
		{t: "int32", val: "-1", expected: "-1"},
		{t: "int64", val: "1", expected: "1"},
		{t: "int64", val: "-1", expected: "-1"},

		{t: "uint", val: "1", expected: "1"},
		{t: "uint8", val: "1", expected: "1"},
		{t: "uint16", val: "1", expected: "1"},
		{t: "uint32", val: "1", expected: "1"},
		{t: "uint64", val: "1", expected: "1"},

		{t: "float32", val: "3.14", expected: "3.14"},
		{t: "float32", val: "-3.14", expected: "-3.14"},
		{t: "float64", val: "3.14", expected: "3.14"},
		{t: "float64", val: "-3.14", expected: "-3.14"},

		{t: "bool", val: "true", expected: "true"},
		{t: "bool", val: "false", expected: "false"},

		{t: "time.Duration", val: "1h", expected: "1h"},
	}

	for _, tt := range cases {
		t.Run(tt.t, func(t *testing.T) {
			err := checkDefaultValue(tt.t, tt.val)
			assert.Nil(t, err)
		})
	}
}

func Test_normalizeTypeName(t *testing.T) {
	cases := []struct {
		name     string
		val      string
		expected string
	}{
		{name: "int", val: "int", expected: "int"},
		{name: "*int", val: "*int", expected: "int"},
		{name: "[]int", val: "[]int", expected: "int"},
		{name: "[]*int", val: "[]*int", expected: "int"},
		{name: "some.Struct", val: "some.Struct", expected: "Struct"},
		{name: "*some.Struct", val: "*some.Struct", expected: "Struct"},
		{name: "[]some.Struct", val: "[]some.Struct", expected: "Struct"},
		{name: "[]*some.Struct", val: "[]*some.Struct", expected: "Struct"},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, normalizeTypeName(tt.val))
		})
	}
}

func Test_typeParamsStr(t *testing.T) {
	testCases := []struct {
		name       string
		params     []*ast.Field
		wantSpec   string
		wantParams string
		wantErr    string
	}{
		{
			name:       "empty",
			params:     nil,
			wantSpec:   "",
			wantParams: "",
		},
		{
			name: "single_param",
			params: []*ast.Field{
				{
					Names: []*ast.Ident{{Name: "T"}},
					Type:  &ast.Ident{Name: "any"},
				},
			},
			wantSpec:   "[T any]",
			wantParams: "[T]",
		},
		{
			name: "multiple_params",
			params: []*ast.Field{
				{
					Names: []*ast.Ident{{Name: "T"}},
					Type:  &ast.Ident{Name: "any"},
				},
				{
					Names: []*ast.Ident{{Name: "K"}},
					Type:  &ast.Ident{Name: "comparable"},
				},
			},
			wantSpec:   "[T any, K comparable]",
			wantParams: "[T, K]",
		},
		{
			name: "multiple_names_in_one_field",
			params: []*ast.Field{
				{
					Names: []*ast.Ident{{Name: "T"}, {Name: "K"}},
					Type:  &ast.Ident{Name: "comparable"},
				},
			},
			wantSpec:   "[T, K comparable]",
			wantParams: "[T, K]",
		},
		{
			name: "unnamed_param",
			params: []*ast.Field{
				{
					Type: &ast.Ident{Name: "any"},
				},
			},
			wantErr: "unnamed param any",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			gotSpec, gotParams, err := typeParamsStr(testCase.params)
			if testCase.wantErr != "" {
				require.EqualError(t, err, testCase.wantErr)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, testCase.wantSpec, gotSpec)
			assert.Equal(t, testCase.wantParams, gotParams)
		})
	}
}

func Test_optimizeGeneratedSource_PrunesOnlyUnusedNamedImports(t *testing.T) {
	src := []byte(`package testcase

import (
	"fmt"
	"io"
	. "math"
	_ "net/http/pprof"
	alias "strings"
)

var _ = fmt.Sprintf
var _ = alias.Builder{}
var _ = Pi
`)

	got, err := optimizeGeneratedSource(src)
	require.NoError(t, err)

	gotStr := string(got)
	require.Contains(t, gotStr, `"fmt"`)
	require.Contains(t, gotStr, `alias "strings"`)
	require.Contains(t, gotStr, `. "math"`)
	require.Contains(t, gotStr, `_ "net/http/pprof"`)
	require.NotContains(t, gotStr, `"io"`)
}

func Test_parseModulePath(t *testing.T) {
	tests := []struct {
		name    string
		goMod   string
		want    string
		wantErr string
	}{
		{
			name: "plain",
			goMod: `module example.com/project

go 1.26
`,
			want: "example.com/project",
		},
		{
			name: "trailing_comment",
			goMod: `// leading comment
module example.com/project // human note

go 1.26
`,
			want: "example.com/project",
		},
		{
			name:    "missing",
			goMod:   "go 1.26\n",
			wantErr: "module path not found",
		},
		{
			name:    "empty",
			goMod:   "module \n",
			wantErr: "module path not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseModulePath(tt.goMod)
			if tt.wantErr != "" {
				require.EqualError(t, err, tt.wantErr)

				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestGetOptionSpec_LocalAliasStructMergesCallerImportsAndCompiles(t *testing.T) {
	tmpDir := t.TempDir()
	writeTestFile(t, filepath.Join(tmpDir, "go.mod"), `module example.com/safety // valid trailing comment

go 1.26
`)
	writeTestFile(t, filepath.Join(tmpDir, "optionspkg", "options.go"), `package optionspkg

type Local struct{}

type Options struct {
	Required string `+"`option:\"mandatory\"`"+`
	Optional Local
	Ptr *Local
	Values []Local
	Mapping map[string]Local
	Callback func(Local) *Local
	private string
}
`)
	consumerDir := filepath.Join(tmpDir, "consumer")
	inFilename := filepath.Join(consumerDir, "options.go")
	writeTestFile(t, inFilename, `package consumer

import alias "example.com/safety/optionspkg"

type Options alias.Options
`)

	spec, err := GetOptionSpec(inFilename, "Options", "default", false, nil)
	require.NoError(t, err)
	require.Len(t, spec.Spec.Options, 6)
	require.Contains(t, spec.Imports, Import{
		Path: `"example.com/safety/optionspkg"`,
		Alias: func() *string {
			alias := "alias"

			return &alias
		}(),
	})

	rendered, err := Render(NewOptions(
		WithVersion("test"),
		WithPackageName("consumer"),
		WithOptionsStructName("Options"),
		WithFileImports(spec.Imports),
		WithSpec(&spec.Spec),
		WithTagName("default"),
		WithConstructorTypeRender("public"),
		WithOptionTypeName("OptOptionsSetter"),
	))
	require.NoError(t, err)

	renderedStr := string(rendered)
	require.Contains(t, renderedStr, `alias "example.com/safety/optionspkg"`)
	require.Contains(t, renderedStr, `func WithOptional(opt alias.Local) OptOptionsSetter`)
	require.Contains(t, renderedStr, `func WithPtr(opt *alias.Local) OptOptionsSetter`)
	require.Contains(t, renderedStr, `func WithValues(opt []alias.Local) OptOptionsSetter`)
	require.Contains(t, renderedStr, `func WithMapping(opt map[string]alias.Local) OptOptionsSetter`)
	require.Contains(t, renderedStr, `func WithCallback(opt func(alias.Local) *alias.Local) OptOptionsSetter`)
	require.NotContains(t, renderedStr, "private")

	writeTestFile(t, filepath.Join(consumerDir, "options_generated.go"), renderedStr)
	cmd := exec.Command("go", "test", "./consumer")
	cmd.Dir = tmpDir
	cmd.Env = append(os.Environ(), "GOCACHE="+filepath.Join(tmpDir, "gocache"))
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, string(out))
}

func Test_findLocalStructTypeParamsAndFields(t *testing.T) {
	tmpDir := t.TempDir()
	writeTestFile(t, filepath.Join(tmpDir, "go.mod"), `module example.com/local

go 1.26
`)
	writeTestFile(t, filepath.Join(tmpDir, "pkg", "options.go"), `package pkg

type Local struct{}

type Options[T any] struct {
	Required string `+"`option:\"mandatory\"`"+`
	Optional Local
	Nested map[string][]*Local
	private string
}
`)

	file, typeParams, fields, err := findLocalStructTypeParamsAndFields(
		token.NewFileSet(),
		"example.com/local/pkg",
		"Options",
		filepath.Join(tmpDir, "consumer"),
		"alias",
	)
	require.NoError(t, err)
	require.Equal(t, "pkg", file.Name.Name)

	typeParamsSpec, typeParamsNames, err := typeParamsStr(typeParams)
	require.NoError(t, err)
	require.Equal(t, "[T any]", typeParamsSpec)
	require.Equal(t, "[T]", typeParamsNames)

	require.Len(t, fields, 3)
	require.Equal(t, "Required", fields[0].Names[0].Name)
	require.Equal(t, "string", renderExprString(fields[0].Type))
	require.Equal(t, "Optional", fields[1].Names[0].Name)
	require.Equal(t, "alias.Local", renderExprString(fields[1].Type))
	require.Equal(t, "Nested", fields[2].Names[0].Name)
	require.Equal(t, "map[string][]*alias.Local", renderExprString(fields[2].Type))
}

func Test_findLocalStructTypeParamsAndFields_RejectsOutsideModuleImport(t *testing.T) {
	tmpDir := t.TempDir()
	writeTestFile(t, filepath.Join(tmpDir, "go.mod"), `module example.com/local

go 1.26
`)

	_, _, _, err := findLocalStructTypeParamsAndFields(
		token.NewFileSet(),
		"example.com/other/pkg",
		"Options",
		filepath.Join(tmpDir, "consumer"),
		"other",
	)
	require.EqualError(t, err, "import is outside current module")
}

func Test_findStructTypeParamsAndFields2_FallbackPackagesLoad(t *testing.T) {
	tmpDir := t.TempDir()
	pkgFile := filepath.Join(tmpDir, "external.go")
	writeTestFile(t, pkgFile, `package external

type Local struct{}

type Options[T any] struct {
	Required string `+"`option:\"mandatory\"`"+`
	Optional Local
	Ptr *Local
	Values []Local
	Mapping map[string]Local
	Callback func(Local) *Local
	private string
}
`)

	file, typeParams, fields, err := findStructTypeParamsAndFields2(
		token.NewFileSet(),
		pkgFile,
		"Options",
		tmpDir,
		"external",
	)
	require.NoError(t, err)
	require.Equal(t, "external", file.Name.Name)

	typeParamsSpec, typeParamsNames, err := typeParamsStr(typeParams)
	require.NoError(t, err)
	require.Equal(t, "[T any]", typeParamsSpec)
	require.Equal(t, "[T]", typeParamsNames)

	require.Len(t, fields, 6)
	require.Equal(t, "Required", fields[0].Names[0].Name)
	require.Equal(t, "string", renderExprString(fields[0].Type))
	require.Equal(t, "Optional", fields[1].Names[0].Name)
	require.Equal(t, "external.Local", renderExprString(fields[1].Type))
	require.Equal(t, "Ptr", fields[2].Names[0].Name)
	require.Equal(t, "*external.Local", renderExprString(fields[2].Type))
	require.Equal(t, "Values", fields[3].Names[0].Name)
	require.Equal(t, "[]external.Local", renderExprString(fields[3].Type))
	require.Equal(t, "Mapping", fields[4].Names[0].Name)
	require.Equal(t, "map[string]external.Local", renderExprString(fields[4].Type))
	require.Equal(t, "Callback", fields[5].Names[0].Name)
	require.Equal(t, "func(external.Local) *external.Local", renderExprString(fields[5].Type))
}

func writeTestFile(t *testing.T, filename, content string) {
	t.Helper()

	require.NoError(t, os.MkdirAll(filepath.Dir(filename), 0o755))
	require.NoError(t, os.WriteFile(filename, []byte(content), ctype.DefaultPermission))
}

func Test_parseTag(t *testing.T) {
	testCases := []struct {
		name         string
		tag          *ast.BasicLit
		fieldName    string
		tagName      string
		wantOption   TagOption
		wantWarnings []string
	}{
		{
			name:      "nil_tag",
			tag:       nil,
			fieldName: "fieldName",
			tagName:   "default",
			wantOption: TagOption{
				IsRequired:    false,
				GoValidator:   "",
				Default:       "",
				Variadic:      false,
				VariadicIsSet: false,
				Skip:          false,
			},
			wantWarnings: nil,
		},
		{
			name:      "validate_and_default",
			tag:       &ast.BasicLit{Value: "`validate:\"required,email\" default:\"42\"`"},
			fieldName: "fieldName",
			tagName:   "default",
			wantOption: TagOption{
				IsRequired:    false,
				GoValidator:   "required,email",
				Default:       "42",
				Variadic:      false,
				VariadicIsSet: false,
				Skip:          false,
			},
			wantWarnings: nil,
		},
		{
			name:      "mandatory_variadic_and_skip",
			tag:       &ast.BasicLit{Value: "`option:\"mandatory,variadic=true,-\"`"},
			fieldName: "fieldName",
			tagName:   "default",
			wantOption: TagOption{
				IsRequired:    true,
				GoValidator:   "",
				Default:       "",
				Variadic:      true,
				VariadicIsSet: true,
				Skip:          true,
			},
			wantWarnings: nil,
		},
		{
			name:      "deprecated_required_and_not_empty",
			tag:       &ast.BasicLit{Value: "`option:\"required,not-empty\" validate:\"min=10\"`"},
			fieldName: "fieldName",
			tagName:   "default",
			wantOption: TagOption{
				IsRequired:    true,
				GoValidator:   "min=10,required",
				Default:       "",
				Variadic:      false,
				VariadicIsSet: false,
				Skip:          false,
			},
			wantWarnings: []string{
				"Deprecated: use `option:\"mandatory\"` instead for field `fieldName` " +
					"to force the passing option in the constructor argument\n",
				"Deprecated: use github.com/go-playground/validator `validate` tag to check the field `fieldName` content\n",
			},
		},
		{
			name:      "invalid_variadic_value",
			tag:       &ast.BasicLit{Value: "`option:\"variadic=bad\"`"},
			fieldName: "fieldName",
			tagName:   "default",
			wantOption: TagOption{
				IsRequired:    false,
				GoValidator:   "",
				Default:       "",
				Variadic:      false,
				VariadicIsSet: true,
				Skip:          false,
			},
			wantWarnings: []string{
				"Error: parse variadic for the field fieldName failed: strconv.ParseBool: parsing \"bad\": invalid syntax\n",
			},
		},
		{
			name:      "name replace",
			tag:       &ast.BasicLit{Value: "`option:\"name=Some\"`"},
			fieldName: "fieldName",
			tagName:   "default",
			wantOption: TagOption{
				IsRequired:    false,
				GoValidator:   "",
				Default:       "",
				Variadic:      false,
				VariadicIsSet: false,
				Skip:          false,
				Name:          "Some",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gotOption, gotWarnings := parseTag(tc.tag, tc.fieldName, tc.tagName)
			assert.Equal(t, tc.wantOption, gotOption)
			assert.Equal(t, tc.wantWarnings, gotWarnings)
		})
	}
}

func BenchmarkFormatComment(b *testing.B) {
	b.Run("short", func(b *testing.B) {
		comment := "short comment"

		b.ReportAllocs()
		for b.Loop() {
			benchmarkFormatCommentSink = formatComment(comment)
		}
	})

	b.Run("multiline", func(b *testing.B) {
		comment := strings.Join([]string{
			"client config comment",
			"another line with details",
			"",
			"last line",
		}, "\n")

		b.ReportAllocs()
		for b.Loop() {
			benchmarkFormatCommentSink = formatComment(comment)
		}
	})

	b.Run("large", func(b *testing.B) {
		lines := make([]string, 64)
		for i := range lines {
			lines[i] = "benchmark line for comment formatting"
		}

		comment := strings.Join(lines, "\n")

		b.ReportAllocs()
		for b.Loop() {
			benchmarkFormatCommentSink = formatComment(comment)
		}
	})
}

func BenchmarkApplyExcludes(b *testing.B) {
	buildOptions := func(n int) []OptionMeta {
		options := make([]OptionMeta, n)
		for i := range options {
			options[i] = OptionMeta{
				Name:  "Field" + strconv.Itoa(i),
				Field: "field" + strconv.Itoa(i),
				Type:  "string",
			}
		}

		return options
	}

	buildPatterns := func(patterns ...string) []*regexp.Regexp {
		res := make([]*regexp.Regexp, len(patterns))
		for i, pattern := range patterns {
			res[i] = regexp.MustCompile(pattern)
		}

		return res
	}

	b.Run("small", func(b *testing.B) {
		options := buildOptions(32)
		excludes := buildPatterns("^Field1$", "^Field2$", "^Field3$")

		b.ReportAllocs()
		for b.Loop() {
			benchmarkApplyExcludesSink = ApplyExcludes(options, excludes)
		}
	})

	b.Run("medium", func(b *testing.B) {
		options := buildOptions(256)
		excludes := buildPatterns("^Field1[0-9]$", "^Field2[0-9]$", "^Field3[0-9]$", "^Field4[0-9]$")

		b.ReportAllocs()
		for b.Loop() {
			benchmarkApplyExcludesSink = ApplyExcludes(options, excludes)
		}
	})

	b.Run("large", func(b *testing.B) {
		options := buildOptions(1024)
		excludes := buildPatterns(
			"^Field1[0-9]{2}$",
			"^Field2[0-9]{2}$",
			"^Field3[0-9]{2}$",
			"^Field4[0-9]{2}$",
			"^Field5[0-9]{2}$",
		)

		b.ReportAllocs()
		for b.Loop() {
			benchmarkApplyExcludesSink = ApplyExcludes(options, excludes)
		}
	})
}

func Test_findImportPath(t *testing.T) {
	imports := []*ast.ImportSpec{
		{
			Path: &ast.BasicLit{Value: `"fmt"`},
		},
		{
			Name: &ast.Ident{Name: "aliaspkg"},
			Path: &ast.BasicLit{Value: `"github.com/example/project/pkg"`},
		},
		{
			Path: &ast.BasicLit{Value: `"github.com/org/lib/v2"`},
		},
		{
			Path: &ast.BasicLit{Value: `"github.com/company/service/internal/transport/httpapi"`},
		},
		{
			Path: &ast.BasicLit{Value: `broken`},
		},
	}

	testCases := []struct {
		name      string
		pkgName   string
		wantPath  string
		wantAlias string
	}{
		{
			name:      "standard_library_package",
			pkgName:   "fmt",
			wantPath:  "fmt",
			wantAlias: "fmt",
		},
		{
			name:      "aliased_import",
			pkgName:   "aliaspkg",
			wantPath:  "github.com/example/project/pkg",
			wantAlias: "aliaspkg",
		},
		{
			name:      "versioned_import_uses_previous_path_segment",
			pkgName:   "lib",
			wantPath:  "github.com/org/lib/v2",
			wantAlias: "lib",
		},
		{
			name:      "late_match",
			pkgName:   "httpapi",
			wantPath:  "github.com/company/service/internal/transport/httpapi",
			wantAlias: "httpapi",
		},
		{
			name:      "not_found",
			pkgName:   "missingpkg",
			wantPath:  "",
			wantAlias: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gotPath, gotAlias := findImportPath(imports, tc.pkgName)
			assert.Equal(t, tc.wantPath, gotPath)
			assert.Equal(t, tc.wantAlias, gotAlias)
		})
	}
}

func Test_importPathBase(t *testing.T) {
	testCases := []struct {
		name       string
		importPath string
		want       string
	}{
		{
			name:       "single_segment",
			importPath: "fmt",
			want:       "fmt",
		},
		{
			name:       "multi_segment",
			importPath: "github.com/company/project/pkg",
			want:       "pkg",
		},
		{
			name:       "version_suffix",
			importPath: "github.com/org/lib/v2",
			want:       "lib",
		},
		{
			name:       "non_version_v_prefix",
			importPath: "github.com/org/value",
			want:       "value",
		},
		{
			name:       "root_version_segment",
			importPath: "v2",
			want:       "v2",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, importPathBase(tc.importPath))
		})
	}
}

func BenchmarkFindImportPath(b *testing.B) {
	imports := []*ast.ImportSpec{
		{
			Path: &ast.BasicLit{Value: `"fmt"`},
		},
		{
			Path: &ast.BasicLit{Value: `"strings"`},
		},
		{
			Name: &ast.Ident{Name: "aliaspkg"},
			Path: &ast.BasicLit{Value: `"github.com/example/project/pkg"`},
		},
		{
			Path: &ast.BasicLit{Value: `"github.com/org/lib/v2"`},
		},
		{
			Path: &ast.BasicLit{Value: `"github.com/company/service/internal/domain"`},
		},
		{
			Path: &ast.BasicLit{Value: `"github.com/company/service/internal/repository"`},
		},
		{
			Path: &ast.BasicLit{Value: `"github.com/company/service/internal/transport/httpapi"`},
		},
	}

	b.Run("base_package_match_late", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			benchmarkFindImportPathPathSink, benchmarkFindImportPathAliasSink = findImportPath(imports, "httpapi")
		}
	})

	b.Run("alias_match", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			benchmarkFindImportPathPathSink, benchmarkFindImportPathAliasSink = findImportPath(imports, "aliaspkg")
		}
	})

	b.Run("not_found", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			benchmarkFindImportPathPathSink, benchmarkFindImportPathAliasSink = findImportPath(imports, "missingpkg")
		}
	})
}

func BenchmarkTypeParamsStr(b *testing.B) {
	buildParams := func(paramsCount, namesPerParam int) []*ast.Field {
		params := make([]*ast.Field, paramsCount)
		for i := range params {
			names := make([]*ast.Ident, namesPerParam)
			for j := range names {
				names[j] = &ast.Ident{Name: "T" + strconv.Itoa(i) + "_" + strconv.Itoa(j)}
			}

			params[i] = &ast.Field{
				Names: names,
				Type:  &ast.Ident{Name: "any"},
			}
		}

		return params
	}

	b.Run("small", func(b *testing.B) {
		params := buildParams(3, 1)

		b.ReportAllocs()
		for b.Loop() {
			var err error
			benchmarkTypeParamsStrSpecSink, benchmarkTypeParamsStrNamesSink, err = typeParamsStr(params)
			require.NoError(b, err)
		}
	})

	b.Run("multi_name_fields", func(b *testing.B) {
		params := buildParams(4, 3)

		b.ReportAllocs()
		for b.Loop() {
			var err error
			benchmarkTypeParamsStrSpecSink, benchmarkTypeParamsStrNamesSink, err = typeParamsStr(params)
			require.NoError(b, err)
		}
	})

	b.Run("large", func(b *testing.B) {
		params := buildParams(32, 2)

		b.ReportAllocs()
		for b.Loop() {
			var err error
			benchmarkTypeParamsStrSpecSink, benchmarkTypeParamsStrNamesSink, err = typeParamsStr(params)
			require.NoError(b, err)
		}
	})
}

func BenchmarkParseTag(b *testing.B) {
	b.Run("simple", func(b *testing.B) {
		tag := &ast.BasicLit{Value: "`validate:\"required\" default:\"42\"`"}

		b.ReportAllocs()
		for b.Loop() {
			benchmarkParseTagOptionSink, benchmarkParseTagWarningsSink = parseTag(tag, "fieldName", "default")
		}
	})

	b.Run("option_flags", func(b *testing.B) {
		tag := &ast.BasicLit{Value: "`option:\"mandatory,variadic=true\" validate:\"min=10\" default:\"1m\"`"}

		b.ReportAllocs()
		for b.Loop() {
			benchmarkParseTagOptionSink, benchmarkParseTagWarningsSink = parseTag(tag, "fieldName", "default")
		}
	})

	b.Run("deprecated_options", func(b *testing.B) {
		tag := &ast.BasicLit{Value: "`option:\"required,not-empty,variadic=bad\" validate:\"email\"`"}

		b.ReportAllocs()
		for b.Loop() {
			benchmarkParseTagOptionSink, benchmarkParseTagWarningsSink = parseTag(tag, "fieldName", "default")
		}
	})

	b.Run("nil_tag", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			benchmarkParseTagOptionSink, benchmarkParseTagWarningsSink = parseTag(nil, "fieldName", "default")
		}
	})
}

func BenchmarkImportPathBase(b *testing.B) {
	b.Run("single_segment", func(b *testing.B) {
		importPath := "fmt"

		b.ReportAllocs()
		for b.Loop() {
			benchmarkImportPathBaseSink = importPathBase(importPath)
		}
	})

	b.Run("deep_path", func(b *testing.B) {
		importPath := "github.com/company/service/internal/transport/httpapi"

		b.ReportAllocs()
		for b.Loop() {
			benchmarkImportPathBaseSink = importPathBase(importPath)
		}
	})

	b.Run("version_suffix", func(b *testing.B) {
		importPath := "github.com/org/lib/v2"

		b.ReportAllocs()
		for b.Loop() {
			benchmarkImportPathBaseSink = importPathBase(importPath)
		}
	})

	b.Run("non_version_v_prefix", func(b *testing.B) {
		importPath := "github.com/org/value"

		b.ReportAllocs()
		for b.Loop() {
			benchmarkImportPathBaseSink = importPathBase(importPath)
		}
	})
}

func BenchmarkExtractSliceElemType(b *testing.B) {
	tempDir := b.TempDir()

	somepkgDir := tempDir + "/somepkg"
	require.NoError(b, os.MkdirAll(somepkgDir, 0o755))

	somepkgContent := `package somepkg

type SliceInt []int
type User struct {
	ID string
}
`
	require.NoError(b, os.WriteFile(somepkgDir+"/somepkg.go", []byte(somepkgContent), ctype.DefaultPermission))
	require.NoError(b, os.WriteFile(tempDir+"/go.mod", []byte("module xxx\ngo 1.18"), ctype.DefaultPermission))
	require.NoError(
		b,
		os.WriteFile(
			tempDir+"/main.go",
			[]byte("package main\n\nimport \"./somepkg\"\n"),
			ctype.DefaultPermission,
		),
	)

	fset := token.NewFileSet()
	mainFile, err := parser.ParseFile(fset, tempDir+"/main.go", nil, parser.ParseComments)
	require.NoError(b, err)

	b.Run("local_slice", func(b *testing.B) {
		expr := &ast.ArrayType{
			Elt: &ast.Ident{Name: "int"},
		}
		store := NewPackageStore(fset, tempDir)

		b.ReportAllocs()
		for b.Loop() {
			benchmarkExtractSliceElemTypeSink, err = extractSliceElemType(mainFile, expr, store)
			require.NoError(b, err)
		}
	})

	b.Run("imported_named_slice", func(b *testing.B) {
		expr := &ast.SelectorExpr{
			X:   &ast.Ident{Name: "somepkg"},
			Sel: &ast.Ident{Name: "SliceInt"},
		}
		store := NewPackageStore(fset, tempDir)

		b.ReportAllocs()
		for b.Loop() {
			benchmarkExtractSliceElemTypeSink, err = extractSliceElemType(mainFile, expr, store)
			require.NoError(b, err)
		}
	})

	b.Run("imported_named_type", func(b *testing.B) {
		expr := &ast.ArrayType{
			Elt: &ast.SelectorExpr{
				X:   &ast.Ident{Name: "somepkg"},
				Sel: &ast.Ident{Name: "User"},
			},
		}
		store := NewPackageStore(fset, tempDir)

		b.ReportAllocs()
		for b.Loop() {
			benchmarkExtractSliceElemTypeSink, err = extractSliceElemType(mainFile, expr, store)
			require.NoError(b, err)
		}
	})
}

func TestExtractSliceElemType(t *testing.T) {
	tempDir := t.TempDir()

	somepkgDir := tempDir + "/somepkg"
	require.NoError(t, os.MkdirAll(somepkgDir, 0o755))

	somepkgContent := `package somepkg

type SliceInt []int
type Ints []int
type Users []User
type User struct {
	ID   string
	Name string
}
type CustomType struct{}
type CustomSlice []CustomType
`
	require.NoError(t, os.WriteFile(somepkgDir+"/somepkg.go", []byte(somepkgContent), ctype.DefaultPermission))

	require.NoError(t, os.WriteFile(tempDir+"/go.mod", []byte("module xxx\ngo 1.18"), ctype.DefaultPermission))

	mainContent := `package main

import "./somepkg"
`

	require.NoError(t, os.WriteFile(tempDir+"/main.go", []byte(mainContent), ctype.DefaultPermission))

	fset := token.NewFileSet()

	mainFile, err := parser.ParseFile(fset, tempDir+"/main.go", nil, parser.ParseComments)
	require.NoError(t, err)

	tests := []struct {
		name string
		expr ast.Expr
		want string
	}{
		{
			name: "slice_of_int",
			expr: &ast.ArrayType{
				Elt: &ast.Ident{Name: "int"},
			},
			want: "int",
		},
		{
			name: "slice_of_slice",
			expr: &ast.ArrayType{
				Elt: &ast.ArrayType{
					Elt: &ast.Ident{Name: "string"},
				},
			},
			want: "[]string",
		},
		{
			name: "slice_of_map",
			expr: &ast.ArrayType{
				Elt: &ast.MapType{
					Key:   &ast.Ident{Name: "string"},
					Value: &ast.Ident{Name: "int"},
				},
			},
			want: "map[string]int",
		},
		{
			name: "slice_of_chan",
			expr: &ast.ArrayType{
				Elt: &ast.ChanType{
					Dir:   ast.SEND | ast.RECV,
					Value: &ast.Ident{Name: "bool"},
				},
			},
			want: "chan bool",
		},
		{
			name: "slice_of_interface",
			expr: &ast.ArrayType{
				Elt: &ast.InterfaceType{
					Methods: &ast.FieldList{},
				},
			},
			want: "interface{}",
		},
		{
			name: "slice_of_external_type",
			expr: &ast.ArrayType{
				Elt: &ast.SelectorExpr{
					X:   &ast.Ident{Name: "somepkg"},
					Sel: &ast.Ident{Name: "User"},
				},
			},
			want: "somepkg.User",
		},
		{
			name: "slice_of_external_type2",
			expr: &ast.SelectorExpr{
				X:   &ast.Ident{Name: "somepkg"},
				Sel: &ast.Ident{Name: "SliceInt"},
			},
			want: "int",
		},
		{
			name: "local_intslice",
			expr: &ast.Ident{
				Name: "IntSlice",
				Obj: &ast.Object{ //nolint:staticcheck
					Kind: ast.Typ,
					Decl: &ast.TypeSpec{
						Name: &ast.Ident{Name: "IntSlice"},
						Type: &ast.ArrayType{
							Elt: &ast.Ident{Name: "int"},
						},
					},
				},
			},
			want: "int",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := extractSliceElemType(mainFile, tt.expr, NewPackageStore(fset, tempDir))
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}

	t.Run("not_a_slice", func(t *testing.T) {
		res, err := extractSliceElemType(
			mainFile,
			&ast.Ident{Name: "string"}, //nolint:exhaustruct
			NewPackageStore(fset, tempDir),
		)
		require.Error(t, err)
		require.Equal(t, "", res)
	})
}

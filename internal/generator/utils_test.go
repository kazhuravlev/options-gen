//nolint:exhaustruct
package generator //nolint:testpackage

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/kazhuravlev/options-gen/internal/ctype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var benchmarkFormatCommentSink string
var benchmarkApplyExcludesSink []OptionMeta
var benchmarkFindImportPathPathSink string
var benchmarkFindImportPathAliasSink string

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
		excludes := buildPatterns("^Field1[0-9]{2}$", "^Field2[0-9]{2}$", "^Field3[0-9]{2}$", "^Field4[0-9]{2}$", "^Field5[0-9]{2}$")

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
			got, err := extractSliceElemType(tempDir, fset, mainFile, tt.expr, NewPackageStore(fset, tempDir))
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}

	t.Run("not_a_slice", func(t *testing.T) {
		res, err := extractSliceElemType(tempDir, fset, mainFile, &ast.Ident{Name: "string"}, NewPackageStore(fset, tempDir)) //nolint:exhaustruct
		require.Error(t, err)
		require.Equal(t, "", res)
	})
}

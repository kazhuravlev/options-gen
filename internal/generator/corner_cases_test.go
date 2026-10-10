//nolint:exhaustruct
package generator //nolint:testpackage

import (
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_parseTag_CornerCases(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name         string
		tag          string
		wantOption   TagOption
		wantWarnings []string
	}{
		{
			name:       "interpreted_string_literal",
			tag:        `"option:\"mandatory\" default:\"42\""`,
			wantOption: TagOption{IsRequired: true, Default: "42"},
		},
		{
			name:       "variadic_without_value",
			tag:        "`option:\"variadic\"`",
			wantOption: TagOption{VariadicIsSet: true},
			wantWarnings: []string{
				"Error: parse variadic for the field f failed: strconv.ParseBool: parsing \"\": invalid syntax\n",
			},
		},
		{
			name:       "variadic_numeric_value",
			tag:        "`option:\"variadic=1\"`",
			wantOption: TagOption{Variadic: true, VariadicIsSet: true},
		},
		{
			name:       "variadic_false_is_still_set",
			tag:        "`option:\"variadic=false\"`",
			wantOption: TagOption{VariadicIsSet: true},
		},
		{
			name:       "skip_combined_with_other_flags",
			tag:        "`option:\"-,mandatory\" default:\"1\"`",
			wantOption: TagOption{IsRequired: true, Default: "1", Skip: true},
		},
		{
			name:       "empty_name_keeps_default_naming",
			tag:        "`option:\"name=\"`",
			wantOption: TagOption{},
		},
		{
			name:       "unknown_options_are_ignored",
			tag:        "`option:\"foo,bar=1,,mandatory\"`",
			wantOption: TagOption{IsRequired: true},
		},
		{
			name:       "space_after_comma_is_part_of_the_option_name",
			tag:        "`option:\"mandatory, variadic=true\"`",
			wantOption: TagOption{IsRequired: true},
		},
		{
			name:         "not_empty_appends_required",
			tag:          "`option:\"not-empty\" validate:\"email\"`",
			wantOption:   TagOption{GoValidator: "email,required"},
			wantWarnings: []string{deprecatedNotEmptyWarning("f")},
		},
		{
			// Documents the substring check: `required_if` already "contains" required.
			name:         "not_empty_with_required_if_does_not_add_required",
			tag:          "`option:\"not-empty\" validate:\"required_if=Other 1\"`",
			wantOption:   TagOption{GoValidator: "required_if=Other 1"},
			wantWarnings: []string{deprecatedNotEmptyWarning("f")},
		},
		{
			name:       "malformed_tag_is_ignored",
			tag:        "`option:\"mandatory`",
			wantOption: TagOption{},
		},
		{
			name:       "unquoted_literal_falls_back_to_raw_value",
			tag:        `option:"mandatory"`,
			wantOption: TagOption{IsRequired: true},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			tag := &ast.BasicLit{Kind: token.STRING, Value: testCase.tag}
			gotOption, gotWarnings := parseTag(tag, "f", "default")
			require.Equal(t, testCase.wantOption, gotOption)
			require.Equal(t, testCase.wantWarnings, gotWarnings)
		})
	}
}

func Test_checkDefaultValue_Bounds(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		fieldType string
		value     string
		wantErr   string // Substring of the error, empty for a valid value.
	}{
		{fieldType: "int8", value: "127"},
		{fieldType: "int8", value: "-128"},
		{fieldType: "int8", value: "128", wantErr: "value out of range"},
		{fieldType: "int8", value: "-129", wantErr: "value out of range"},
		{fieldType: "int16", value: "32767"},
		{fieldType: "int16", value: "32768", wantErr: "value out of range"},
		{fieldType: "int32", value: "2147483647"},
		{fieldType: "int32", value: "2147483648", wantErr: "value out of range"},
		{fieldType: "int64", value: "9223372036854775807"},
		{fieldType: "int64", value: "9223372036854775808", wantErr: "value out of range"},
		{fieldType: "int", value: "9223372036854775808", wantErr: "value out of range"},
		{fieldType: "uint8", value: "255"},
		{fieldType: "uint8", value: "256", wantErr: "value out of range"},
		{fieldType: "uint16", value: "65535"},
		{fieldType: "uint16", value: "65536", wantErr: "value out of range"},
		{fieldType: "uint32", value: "4294967295"},
		{fieldType: "uint32", value: "4294967296", wantErr: "value out of range"},
		{fieldType: "uint64", value: "18446744073709551615"},
		{fieldType: "uint64", value: "18446744073709551616", wantErr: "value out of range"},
		{fieldType: "uint", value: "18446744073709551616", wantErr: "value out of range"},
		{fieldType: "float32", value: "3.4e38"},
		{fieldType: "float32", value: "1e40", wantErr: "value out of range"},
		{fieldType: "float64", value: "1e308"},
		{fieldType: "float64", value: "1e400", wantErr: "value out of range"},
		{fieldType: "float64", value: "NaN", wantErr: "NaN and Inf"},
		{fieldType: "float64", value: "Inf", wantErr: "NaN and Inf"},
		{fieldType: "float64", value: "-inf", wantErr: "NaN and Inf"},
		{fieldType: "float32", value: "infinity", wantErr: "NaN and Inf"},
		{fieldType: "float64", value: "0x1p-2"},
		{fieldType: "float64", value: ".5"},
		{fieldType: "float64", value: "5."},
		{fieldType: "float64", value: "1e5"},
		{fieldType: "float64", value: "+1.5"},
		{fieldType: "float64", value: "1e-400"},
		{fieldType: "int", value: "0"},
		{fieldType: "int", value: "-0"},
		{fieldType: "int", value: "+5"},
		{fieldType: "int", value: "08", wantErr: "leading zeros"},
		{fieldType: "int", value: "010", wantErr: "leading zeros"},
		{fieldType: "int", value: "00", wantErr: "leading zeros"},
		{fieldType: "int", value: "-010", wantErr: "leading zeros"},
		{fieldType: "uint", value: "007", wantErr: "leading zeros"},
		{fieldType: "uint", value: "+5", wantErr: "invalid syntax"},
		{fieldType: "int", value: "1_000", wantErr: "invalid syntax"},
		{fieldType: "int", value: "0x10", wantErr: "invalid syntax"},
		{fieldType: "int", value: " 1", wantErr: "invalid syntax"},
		{fieldType: "int", value: "", wantErr: "invalid syntax"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.fieldType+"_"+testCase.value, func(t *testing.T) {
			t.Parallel()

			err := checkDefaultValue(testCase.fieldType, testCase.value)
			if testCase.wantErr == "" {
				require.NoError(t, err)

				return
			}

			require.ErrorContains(t, err, testCase.wantErr)
			require.ErrorContains(t, err, "bad default value")
		})
	}
}

func Test_normalizeTypeName_Generics(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		typeName string
		want     string
	}{
		{typeName: "pkg.Box[int]", want: "Box"},
		{typeName: "Box[pkg.T]", want: "Box"},
		{typeName: "*pkg.Box[map[string]int]", want: "Box"},
		{typeName: "[]pkg.Box[int]", want: "Box"},
		{typeName: "Box[[]int]", want: "Box"},
		{typeName: "Pair[int, string]", want: "Pair"},
		{typeName: "map[string]int", want: "map[string]int"},
		{typeName: "chan int", want: "chan int"},
		{typeName: "[3]int", want: "[3]int"},
		{typeName: "[]", want: ""},
		{typeName: "[int]", want: "[int]"},
		{typeName: "Box[int", want: "Box[int"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.typeName, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, testCase.want, normalizeTypeName(testCase.typeName))
		})
	}
}

func Test_importPathBase_Heuristics(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		importPath string
		want       string
	}{
		{importPath: "gopkg.in/yaml.v3", want: "yaml"},
		{importPath: "gopkg.in/foo/bar.v2", want: "bar"},
		{importPath: "github.com/mattn/go-isatty", want: "isatty"},
		{importPath: "github.com/foo/go-bar/v2", want: "bar"},
		{importPath: "github.com/foo/bar-baz", want: "bar"},
		{importPath: "github.com/foo/go-", want: ""},
		{importPath: "foo/v2", want: "foo"},
		{importPath: "a/b/v10", want: "b"},
		{importPath: "example.com/v2/sub", want: "sub"},
		{importPath: "github.com/foo/v2x", want: "v2x"},
		{importPath: "example.com/pkg_v2", want: "pkg_v2"},
		{importPath: "golang.org/x/tools/go/packages", want: "packages"},
		{importPath: "v", want: "v"},
		{importPath: "", want: ""},
		{importPath: "foo/bar/", want: ""},
	}

	for _, testCase := range testCases {
		t.Run(testCase.importPath, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, testCase.want, importPathBase(testCase.importPath))
		})
	}
}

func Test_importSpecName(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		spec *ast.ImportSpec
		want string
	}{
		{
			name: "alias",
			spec: &ast.ImportSpec{Name: &ast.Ident{Name: "alias"}, Path: &ast.BasicLit{Value: `"github.com/x/y"`}},
			want: "alias",
		},
		{
			name: "dot_import",
			spec: &ast.ImportSpec{Name: &ast.Ident{Name: "."}, Path: &ast.BasicLit{Value: `"math"`}},
			want: ".",
		},
		{
			name: "blank_import",
			spec: &ast.ImportSpec{Name: &ast.Ident{Name: "_"}, Path: &ast.BasicLit{Value: `"embed"`}},
			want: "_",
		},
		{
			name: "invalid_path_literal",
			spec: &ast.ImportSpec{Path: &ast.BasicLit{Value: `broken`}},
			want: "",
		},
		{
			name: "gopkg_in_style",
			spec: &ast.ImportSpec{Path: &ast.BasicLit{Value: `"gopkg.in/yaml.v3"`}},
			want: "yaml",
		},
		{
			name: "major_version_suffix",
			spec: &ast.ImportSpec{Path: &ast.BasicLit{Value: `"github.com/org/lib/v2"`}},
			want: "lib",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, testCase.want, importSpecName(testCase.spec))
		})
	}
}

func Test_findImportPath_SpecialImports(t *testing.T) {
	t.Parallel()

	imports := []*ast.ImportSpec{
		{Path: &ast.BasicLit{Value: `"gopkg.in/yaml.v3"`}},
		{Path: &ast.BasicLit{Value: `"github.com/mattn/go-isatty"`}},
		{Name: &ast.Ident{Name: "."}, Path: &ast.BasicLit{Value: `"math"`}},
		{Name: &ast.Ident{Name: "_"}, Path: &ast.BasicLit{Value: `"embed"`}},
	}

	testCases := []struct {
		pkgName  string
		wantPath string
	}{
		{pkgName: "yaml", wantPath: "gopkg.in/yaml.v3"},
		{pkgName: "isatty", wantPath: "github.com/mattn/go-isatty"},
		{pkgName: ".", wantPath: "math"},
		{pkgName: "_", wantPath: "embed"},
		// A dot import hides the package name: `math.Pi` is not valid in such a file.
		{pkgName: "math", wantPath: ""},
		{pkgName: "", wantPath: ""},
	}

	for _, testCase := range testCases {
		t.Run(testCase.pkgName, func(t *testing.T) {
			t.Parallel()

			gotPath, gotAlias := findImportPath(imports, testCase.pkgName)
			require.Equal(t, testCase.wantPath, gotPath)
			if testCase.wantPath == "" {
				require.Empty(t, gotAlias)
			} else {
				require.Equal(t, testCase.pkgName, gotAlias)
			}
		})
	}
}

// Test_formatGeneratedSource_ImportCornerCases runs the import corner cases
// through the production formatter and through the legacy pipeline that the
// parity tests keep as the reference.
func Test_formatGeneratedSource_ImportCornerCases(t *testing.T) {
	t.Parallel()

	formatters := map[string]func([]byte) ([]byte, error){
		"current": formatGeneratedSource,
		"legacy":  optimizeGeneratedSource,
	}

	testCases := []struct {
		name        string
		src         string
		wantErr     string
		contains    []string
		notContains []string
	}{
		{
			name: "non_identifier_import_paths_are_kept_when_used",
			src: `package testcase

import (
	"gopkg.in/yaml.v3"
	"github.com/mattn/go-isatty"
)

var _ yaml.Node
var _ = isatty.IsTerminal
`,
			contains: []string{`"gopkg.in/yaml.v3"`, `"github.com/mattn/go-isatty"`},
		},
		{
			name: "import_referenced_only_in_a_comment_is_pruned",
			src: `package testcase

import "io"

// io.Reader is only mentioned in this comment.
var _ = 1
`,
			notContains: []string{`"io"`, "import"},
		},
		{
			name: "fully_unused_import_decl_is_removed",
			src: `package testcase

import "fmt"

import (
	"strings"
)

var _ = strings.ToUpper
`,
			contains:    []string{`"strings"`},
			notContains: []string{`"fmt"`},
		},
		{
			name: "single_line_import_is_kept_when_used",
			src: `package testcase

import "fmt"

var _ = fmt.Sprint
`,
			contains: []string{`"fmt"`},
		},
		{
			name:    "unparsable_source",
			src:     "package testcase\n\nvar x = \n",
			wantErr: "parse generated source",
		},
	}

	for formatterName, formatter := range formatters {
		for _, testCase := range testCases {
			t.Run(formatterName+"/"+testCase.name, func(t *testing.T) {
				t.Parallel()

				got, err := formatter([]byte(testCase.src))
				if testCase.wantErr != "" {
					require.ErrorContains(t, err, testCase.wantErr)

					return
				}

				require.NoError(t, err)
				for _, substr := range testCase.contains {
					require.Contains(t, string(got), substr)
				}

				for _, substr := range testCase.notContains {
					require.NotContains(t, string(got), substr)
				}
			})
		}
	}
}

func Test_mergeImportSpecs(t *testing.T) {
	t.Parallel()

	fmtImport := &ast.ImportSpec{Path: &ast.BasicLit{Value: `"fmt"`}}
	fmtImportAgain := &ast.ImportSpec{Path: &ast.BasicLit{Value: `"fmt"`}}
	aliasedFmtImport := &ast.ImportSpec{Name: &ast.Ident{Name: "f"}, Path: &ast.BasicLit{Value: `"fmt"`}}
	osImport := &ast.ImportSpec{Path: &ast.BasicLit{Value: `"os"`}}

	merged := mergeImportSpecs(
		[]*ast.ImportSpec{fmtImport, osImport},
		[]*ast.ImportSpec{fmtImportAgain, aliasedFmtImport, osImport},
		nil,
	)

	// Duplicates are dropped, the same path under another alias is a distinct import,
	// and the order of the first occurrence is preserved.
	require.Equal(t, []*ast.ImportSpec{fmtImport, osImport, aliasedFmtImport}, merged)
	require.Empty(t, mergeImportSpecs())
}

const qualifyTypesSrc = `package pkg

type Local struct{}

type Box[T any] struct{ V T }

type Pair[K comparable, V any] struct{}

type Options struct {
	Plain    Local
	Ptr      *Local
	Slice    []Local
	Array    [2]Local
	Map      map[Local]*Local
	Chan     chan Local
	Func     func(Local, ...Local) (Local, error)
	Iface    interface{ M(Local) Local }
	Struct   struct{ X Local }
	Generic  Box[Local]
	Generics Pair[Local, int]
	Nested   Box[[]Local]
	Paren    (Local)
	Builtin  string
}
`

var qualifiedTypes = map[string]string{
	"Plain":    "pkg.Local",
	"Ptr":      "*pkg.Local",
	"Slice":    "[]pkg.Local",
	"Array":    "[2]pkg.Local",
	"Map":      "map[pkg.Local]*pkg.Local",
	"Chan":     "chan pkg.Local",
	"Func":     "func(pkg.Local, ...pkg.Local) (pkg.Local, error)",
	"Iface":    "interface{M(pkg.Local) pkg.Local}",
	"Struct":   "struct{X pkg.Local}",
	"Generic":  "pkg.Box[pkg.Local]",
	"Generics": "pkg.Pair[pkg.Local, int]",
	"Nested":   "pkg.Box[[]pkg.Local]",
	"Paren":    "(pkg.Local)",
	"Builtin":  "string",
}

func parseOptionsStruct(t *testing.T, fset *token.FileSet, src string) (*ast.File, []*ast.Field) {
	t.Helper()

	file, err := parser.ParseFile(fset, "options.go", src, parser.ParseComments)
	require.NoError(t, err)

	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}

		for _, spec := range genDecl.Specs {
			typeSpec, isTypeSpec := spec.(*ast.TypeSpec)
			if !isTypeSpec || typeSpec.Name.Name != "Options" {
				continue
			}

			structType, isStruct := typeSpec.Type.(*ast.StructType)
			require.True(t, isStruct, "Options must be a struct")

			return file, structType.Fields.List
		}
	}

	require.Fail(t, "Options struct not found")

	return nil, nil
}

func Test_addPackageToLocalType_AllExprKinds(t *testing.T) {
	t.Parallel()

	fset := token.NewFileSet()
	file, fields := parseOptionsStruct(t, fset, qualifyTypesSrc)
	localTypes := packageTypeNames(&ast.Package{Files: map[string]*ast.File{"options.go": file}}) //nolint:staticcheck

	require.Len(t, fields, len(qualifiedTypes))
	for _, field := range fields {
		name := field.Names[0].Name
		got := types.ExprString(addPackageToLocalType(field.Type, "pkg", localTypes))
		assert.Equal(t, qualifiedTypes[name], got, "field %s", name)
	}

	// A slice expression is not a type, but it is traversed as well.
	sliceExpr := addPackageToLocalType(&ast.SliceExpr{X: &ast.Ident{Name: "Local"}}, "pkg", localTypes)
	assert.Equal(t, "pkg.Local[:]", types.ExprString(sliceExpr))
}

func Test_addPackageToType_AllExprKinds(t *testing.T) {
	t.Parallel()

	fset := token.NewFileSet()
	file, fields := parseOptionsStruct(t, fset, qualifyTypesSrc)

	conf := types.Config{Error: func(error) {}}
	typesPkg, err := conf.Check("pkg", fset, []*ast.File{file}, nil)
	require.NoError(t, err)

	scope := typesPkg.Scope()

	require.Len(t, fields, len(qualifiedTypes))
	for _, field := range fields {
		name := field.Names[0].Name
		got := types.ExprString(addPackageToType(field.Type, "pkg", scope))
		assert.Equal(t, qualifiedTypes[name], got, "field %s", name)
	}

	sliceExpr := addPackageToType(&ast.SliceExpr{X: &ast.Ident{Name: "Local"}}, "pkg", scope)
	assert.Equal(t, "pkg.Local[:]", types.ExprString(sliceExpr))
}

func Test_typeParamsStr_Constraints(t *testing.T) {
	t.Parallel()

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "generic.go", `package pkg

import "fmt"

type Options[T interface{ ~int | ~string }, K comparable, V any, X, Y fmt.Stringer] struct{}
`, parser.ParseComments)
	require.NoError(t, err)

	typeSpec, ok := file.Decls[1].(*ast.GenDecl).Specs[0].(*ast.TypeSpec)
	require.True(t, ok)

	spec, names, err := typeParamsStr(extractFields(typeSpec.TypeParams))
	require.NoError(t, err)
	require.Equal(t, "[T interface{~int | ~string}, K comparable, V any, X, Y fmt.Stringer]", spec)
	require.Equal(t, "[T, K, V, X, Y]", names)
}

// writeSliceFixture creates a module with packages that cover the corner cases
// of resolving the element type of an imported slice type. It returns the
// module root, the file set and the parsed root file (whose imports are used
// for the lookups).
func writeSliceFixture(t *testing.T) (string, *token.FileSet, *ast.File) {
	t.Helper()

	dir := t.TempDir()
	writeTestFile(t, filepath.Join(dir, "go.mod"), "module example.com/fx\n\ngo 1.25\n")
	writeTestFile(t, filepath.Join(dir, "other", "other.go"), `package other

type User struct{ ID string }

type (
	Alias    = []int
	Named    []int
	Ptrs     []*User
	Maps     []map[string]User
	Errs     []error
	Funcs    []func(User) error
	NotSlice struct{}
	Str      = string
)
`)
	writeTestFile(t, filepath.Join(dir, "other2", "other2.go"), `package other2

import (
	"example.com/fx/other"
	"example.com/fx/other3"
)

type Users []other.User

type Things []other3.Thing
`)
	writeTestFile(t, filepath.Join(dir, "other3", "other3.go"), "package other3\n\ntype Thing struct{}\n")

	mainFile := filepath.Join(dir, "main.go")
	writeTestFile(t, mainFile, `package fx

import (
	sp "example.com/fx/other"
	"example.com/fx/other2"
)

type Matrix [3]int
`)

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, mainFile, nil, parser.ParseComments)
	require.NoError(t, err)

	return dir, fset, file
}

func selector(pkgName, typeName string) ast.Expr {
	return &ast.SelectorExpr{X: &ast.Ident{Name: pkgName}, Sel: &ast.Ident{Name: typeName}}
}

func Test_extractSliceElemType_CornerCases(t *testing.T) {
	dir, fset, file := writeSliceFixture(t)

	// The store is shared between the sub-tests (it is not safe for concurrent
	// use), so they run sequentially.
	store := NewPackageStore(fset, dir)

	typeParam := &ast.Ident{
		Name: "T",
		Obj:  &ast.Object{Kind: ast.Typ, Decl: &ast.Field{Names: []*ast.Ident{{Name: "T"}}}}, //nolint:staticcheck
	}
	variable := &ast.Ident{
		Name: "x",
		Obj:  &ast.Object{Kind: ast.Var, Decl: &ast.ValueSpec{}}, //nolint:staticcheck
	}
	localArray := &ast.Ident{
		Name: "Matrix",
		Obj: &ast.Object{ //nolint:staticcheck
			Kind: ast.Typ,
			Decl: &ast.TypeSpec{Type: &ast.ArrayType{Len: &ast.BasicLit{Value: "3"}, Elt: &ast.Ident{Name: "int"}}},
		},
	}

	testCases := []struct {
		name    string
		expr    ast.Expr
		want    string
		wantErr string
	}{
		{name: "imported_alias_of_slice", expr: selector("sp", "Alias"), want: "int"},
		{name: "imported_named_slice", expr: selector("sp", "Named"), want: "int"},
		{name: "imported_slice_of_pointers", expr: selector("sp", "Ptrs"), want: "*sp.User"},
		{name: "imported_slice_of_maps", expr: selector("sp", "Maps"), want: "map[string]sp.User"},
		{name: "imported_slice_of_universe_type", expr: selector("sp", "Errs"), want: "error"},
		{name: "imported_slice_of_funcs", expr: selector("sp", "Funcs"), want: "func(sp.User) error"},
		{name: "imported_struct", expr: selector("sp", "NotSlice"), wantErr: errIsNotSlice.Error()},
		{name: "imported_alias_of_basic", expr: selector("sp", "Str"), wantErr: errIsNotSlice.Error()},
		{name: "imported_missing_type", expr: selector("sp", "Missing"), wantErr: "lookup type not found"},
		{name: "element_from_aliased_third_package", expr: selector("other2", "Users"), want: "sp.User"},
		// The current file does not import other3, so the package name is the best guess.
		{name: "element_from_not_imported_package", expr: selector("other2", "Things"), want: "other3.Thing"},
		{name: "unknown_package", expr: selector("unknown", "X"), wantErr: "import path not found"},
		{
			name:    "nested_selector",
			expr:    &ast.SelectorExpr{X: selector("a", "b"), Sel: &ast.Ident{Name: "C"}},
			wantErr: "unsupported selector",
		},
		{
			name:    "fixed_size_array",
			expr:    &ast.ArrayType{Len: &ast.BasicLit{Value: "3"}, Elt: &ast.Ident{Name: "int"}},
			wantErr: errIsNotSlice.Error(),
		},
		{
			name:    "map",
			expr:    &ast.MapType{Key: &ast.Ident{Name: "string"}, Value: &ast.Ident{Name: "int"}},
			wantErr: errIsNotSlice.Error(),
		},
		{name: "pointer", expr: &ast.StarExpr{X: &ast.Ident{Name: "int"}}, wantErr: errIsNotSlice.Error()},
		{name: "type_parameter", expr: typeParam, wantErr: errIsNotSlice.Error()},
		{name: "ident_of_a_variable", expr: variable, wantErr: "unsupported ident expression"},
		{name: "local_fixed_size_array", expr: localArray, wantErr: errIsNotSlice.Error()},
		{
			name: "slice_of_send_chan",
			expr: &ast.ArrayType{Elt: &ast.ChanType{Dir: ast.SEND, Value: &ast.Ident{Name: "int"}}},
			want: "chan<- int",
		},
		{
			name: "slice_of_recv_chan",
			expr: &ast.ArrayType{Elt: &ast.ChanType{Dir: ast.RECV, Value: &ast.Ident{Name: "int"}}},
			want: "<-chan int",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			got, err := extractSliceElemType(file, testCase.expr, store)
			if testCase.wantErr != "" {
				require.EqualError(t, err, testCase.wantErr)
				require.Empty(t, got)

				return
			}

			require.NoError(t, err)
			require.Equal(t, testCase.want, got)
		})
	}

	t.Run("package_that_cannot_be_loaded", func(t *testing.T) {
		brokenStore := NewPackageStore(fset, writeBrokenModule(t))

		got, err := extractSliceElemType(file, selector("sp", "Alias"), brokenStore)
		require.EqualError(t, err, "unable to load package")
		require.Empty(t, got)
	})
}

// writeBrokenModule creates a module whose go.mod cannot be parsed, so that
// every `go list` invocation inside it fails.
func writeBrokenModule(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	writeTestFile(t, filepath.Join(dir, "go.mod"), "module example.com/broken\n\ngo invalid-version\n")

	return dir
}

func Test_PackageStore_Load(t *testing.T) {
	t.Parallel()

	t.Run("caches_loaded_packages", func(t *testing.T) {
		t.Parallel()

		dir, fset, _ := writeSliceFixture(t)
		store := NewPackageStore(fset, dir)

		first, err := store.Load("example.com/fx/other")
		require.NoError(t, err)
		require.NotNil(t, first.Types.Scope().Lookup("User"))

		second, err := store.Load("example.com/fx/other")
		require.NoError(t, err)
		require.Same(t, first, second, "the package must be served from the cache")
		require.Len(t, store.pkgs, 1)
	})

	t.Run("pattern_without_packages", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		writeTestFile(t, filepath.Join(dir, "go.mod"), "module example.com/empty\n\ngo 1.25\n")

		_, err := NewPackageStore(token.NewFileSet(), dir).Load("./...")
		require.EqualError(t, err, "no packages found")
	})

	t.Run("go_list_failure", func(t *testing.T) {
		t.Parallel()

		_, err := NewPackageStore(token.NewFileSet(), writeBrokenModule(t)).Load("example.com/broken")
		require.ErrorContains(t, err, "loading package")
	})
}

func Test_findModule_CornerCases(t *testing.T) {
	t.Parallel()

	t.Run("go_mod_in_parent_directory", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		writeTestFile(t, filepath.Join(dir, "go.mod"), "module example.com/parent\n\ngo 1.25\n")

		root, modulePath, err := findModule(filepath.Join(dir, "a", "b"))
		require.NoError(t, err)
		require.Equal(t, dir, root)
		require.Equal(t, "example.com/parent", modulePath)
	})

	t.Run("quoted_module_path_with_crlf", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		writeTestFile(t, filepath.Join(dir, "go.mod"), "module \"example.com/quoted\" // note\r\n\r\ngo 1.25\r\n")

		_, modulePath, err := findModule(dir)
		require.NoError(t, err)
		require.Equal(t, "example.com/quoted", modulePath)
	})

	t.Run("go_mod_without_module_directive", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		writeTestFile(t, filepath.Join(dir, "go.mod"), "go 1.25\n")

		_, _, err := findModule(dir)
		require.EqualError(t, err, "module path not found")
	})

	t.Run("go_mod_is_a_directory", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		require.NoError(t, os.Mkdir(filepath.Join(dir, "go.mod"), 0o755))

		_, _, err := findModule(dir)
		require.Error(t, err)
		require.NotEqual(t, "go.mod not found", err.Error())
	})

	t.Run("no_go_mod_up_the_tree", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		for parent := dir; ; parent = filepath.Dir(parent) {
			if _, err := os.Stat(filepath.Join(parent, "go.mod")); err == nil {
				t.Skipf("%s is inside a module, cannot test a missing go.mod", dir)
			}

			if filepath.Dir(parent) == parent {
				break
			}
		}

		_, _, err := findModule(dir)
		require.EqualError(t, err, "go.mod not found")
	})
}

func Test_parseModulePath_CornerCases(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		goMod   string
		want    string
		wantErr string
	}{
		{name: "quoted", goMod: "module \"example.com/x\"\n", want: "example.com/x"},
		{name: "crlf", goMod: "module example.com/x\r\ngo 1.25\r\n", want: "example.com/x"},
		{name: "comment_before_module", goMod: "// Deprecated: no\n\nmodule example.com/x\n", want: "example.com/x"},
		{name: "module_as_a_path_segment_only", goMod: "modules example.com/x\n", wantErr: "module path not found"},
		// Documents the current limitation: only a single space is accepted after the keyword.
		{name: "tab_after_keyword", goMod: "module\texample.com/x\n", wantErr: "module path not found"},
		{name: "empty_file", goMod: "", wantErr: "module path not found"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got, err := parseModulePath(testCase.goMod)
			if testCase.wantErr != "" {
				require.EqualError(t, err, testCase.wantErr)

				return
			}

			require.NoError(t, err)
			require.Equal(t, testCase.want, got)
		})
	}
}

func Test_findLocalStructTypeParamsAndFields_CornerCases(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeTestFile(t, filepath.Join(dir, "go.mod"), "module example.com/local\n\ngo 1.25\n")
	writeTestFile(t, filepath.Join(dir, "root.go"), `package local

import "strings"

var Version = strings.ToUpper("v1")

const Answer = 42

func Helper() {}

type NotStruct int

type Root struct {
	Name  string
	Other NotStruct
}
`)
	writeTestFile(t, filepath.Join(dir, "file.go"), "package local\n")
	writeTestFile(t, filepath.Join(dir, "broken", "broken.go"), "package broken\n\ntype Options struct {\n")

	workDir := filepath.Join(dir, "consumer")

	t.Run("struct_in_module_root", func(t *testing.T) {
		t.Parallel()

		file, typeParams, fields, err := findLocalStructTypeParamsAndFields(
			token.NewFileSet(), "example.com/local", "Root", workDir, "local",
		)
		require.NoError(t, err)
		require.Equal(t, "local", file.Name.Name)
		require.Empty(t, typeParams)
		require.Len(t, fields, 2)
		require.Equal(t, "string", types.ExprString(fields[0].Type))
		require.Equal(t, "local.NotStruct", types.ExprString(fields[1].Type))
	})

	t.Run("type_is_not_a_struct", func(t *testing.T) {
		t.Parallel()

		_, _, _, err := findLocalStructTypeParamsAndFields(
			token.NewFileSet(), "example.com/local", "NotStruct", workDir, "local",
		)
		require.EqualError(t, err, "cannot find target struct")
	})

	t.Run("missing_package_directory", func(t *testing.T) {
		t.Parallel()

		_, _, _, err := findLocalStructTypeParamsAndFields(
			token.NewFileSet(), "example.com/local/missing", "Options", workDir, "missing",
		)
		require.ErrorIs(t, err, os.ErrNotExist)
	})

	t.Run("import_path_points_to_a_file", func(t *testing.T) {
		t.Parallel()

		_, _, _, err := findLocalStructTypeParamsAndFields(
			token.NewFileSet(), "example.com/local/file.go", "Options", workDir, "file",
		)
		require.EqualError(t, err, "local import is not a directory")
	})

	t.Run("package_with_syntax_error", func(t *testing.T) {
		t.Parallel()

		_, _, _, err := findLocalStructTypeParamsAndFields(
			token.NewFileSet(), "example.com/local/broken", "Options", workDir, "broken",
		)
		require.ErrorContains(t, err, "cannot parse local package")
	})
}

func Test_findStructTypeParamsAndFields2_CornerCases(t *testing.T) {
	t.Parallel()

	t.Run("empty_work_dir_uses_the_file_directory", func(t *testing.T) {
		t.Parallel()

		pkgFile := filepath.Join(t.TempDir(), "external.go")
		writeTestFile(t, pkgFile, `package external

import "strings"

var upper = strings.ToUpper

func Helper() {}

type NotStruct int

type Options struct {
	Value NotStruct
	private string
}
`)

		file, typeParams, fields, err := findStructTypeParamsAndFields2(token.NewFileSet(), pkgFile, "Options", "", "ext")
		require.NoError(t, err)
		require.Equal(t, "external", file.Name.Name)
		require.Empty(t, typeParams)
		require.Len(t, fields, 1)
		require.Equal(t, "ext.NotStruct", types.ExprString(fields[0].Type))
	})

	t.Run("go_list_failure", func(t *testing.T) {
		t.Parallel()

		_, _, _, err := findStructTypeParamsAndFields2(
			token.NewFileSet(), "example.com/elsewhere/pkg", "Options", writeBrokenModule(t), "pkg",
		)
		require.ErrorContains(t, err, "load package")
	})

	t.Run("pattern_without_packages", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		writeTestFile(t, filepath.Join(dir, "go.mod"), "module example.com/empty\n\ngo 1.25\n")

		_, _, _, err := findStructTypeParamsAndFields2(token.NewFileSet(), "./...", "Options", dir, "pkg")
		require.EqualError(t, err, "no packages found")
	})

	t.Run("type_is_not_a_struct", func(t *testing.T) {
		t.Parallel()

		pkgFile := filepath.Join(t.TempDir(), "external.go")
		writeTestFile(t, pkgFile, "package external\n\ntype Options int\n")

		_, _, _, err := findStructTypeParamsAndFields2(token.NewFileSet(), pkgFile, "Options", "", "ext")
		require.EqualError(t, err, "cannot find target struct")
	})

	t.Run("package_with_type_errors", func(t *testing.T) {
		t.Parallel()

		pkgFile := filepath.Join(t.TempDir(), "external.go")
		writeTestFile(t, pkgFile, "package external\n\nvar x int = \"not an int\"\n\ntype Options struct{}\n")

		_, _, _, err := findStructTypeParamsAndFields2(token.NewFileSet(), pkgFile, "Options", "", "ext")
		require.ErrorContains(t, err, "package contains errors")
	})
}

func simpleOption(name, field, typ string, tagOption TagOption) OptionMeta {
	return OptionMeta{Name: name, Docstring: "", Field: field, Type: typ, TagOption: tagOption}
}

func renderCornerSpec(t *testing.T, spec *OptionSpec, imports []Import, extra ...OptOptionsSetter) (string, error) {
	t.Helper()

	setters := []OptOptionsSetter{
		WithVersion("test"),
		WithPackageName("testpkg"),
		WithOptionsStructName("Options"),
		WithFileImports(imports),
		WithSpec(spec),
		WithTagName("default"),
		WithConstructorTypeRender("public"),
		WithOptionTypeName("OptOptionsSetter"),
	}
	setters = append(setters, extra...)

	rendered, err := Render(NewOptions(setters...))

	return string(rendered), err
}

func Test_makeTemplateOptions(t *testing.T) {
	t.Parallel()

	options := []OptionMeta{
		simpleOption("Field", "field", "string", TagOption{}),
		simpleOption("Field", "field", "string", TagOption{Name: "Custom"}),
	}

	got := makeTemplateOptions(options)
	require.Len(t, got, 2)
	require.Equal(t, "Field", got[0].TargetName)
	require.Equal(t, "field", got[0].TargetField)
	require.Equal(t, "Custom", got[1].TargetName)
	require.Equal(t, "Custom", got[1].TargetField)
	require.Empty(t, makeTemplateOptions(nil))
}

func Test_Render_CornerCases(t *testing.T) {
	t.Parallel()

	plain := simpleOption("Field", "field", "string", TagOption{})

	testCases := []struct {
		name        string
		spec        *OptionSpec
		imports     []Import
		extra       []OptOptionsSetter
		wantErr     string
		contains    []string
		notContains []string
	}{
		{
			name:    "missing_required_configuration",
			spec:    &OptionSpec{Options: []OptionMeta{plain}},
			extra:   []OptOptionsSetter{WithOptionTypeName("")},
			wantErr: "bad configuration",
		},
		{
			name:        "private_constructor",
			spec:        &OptionSpec{Options: []OptionMeta{plain}},
			extra:       []OptOptionsSetter{WithConstructorTypeRender("private")},
			contains:    []string{"func newOptions("},
			notContains: []string{"func NewOptions("},
		},
		{
			name:        "no_constructor",
			spec:        &OptionSpec{Options: []OptionMeta{plain}},
			extra:       []OptOptionsSetter{WithConstructorTypeRender("no")},
			notContains: []string{"func newOptions(", "func NewOptions("},
		},
		{
			name:     "isset_without_options",
			spec:     &OptionSpec{Options: nil},
			extra:    []OptOptionsSetter{WithWithIsset(true)},
			contains: []string{"[0]bool", "func (o *Options) IsSet("},
		},
		{
			name: "name_override_for_mandatory_option",
			spec: &OptionSpec{Options: []OptionMeta{
				simpleOption("Field", "field", "string", TagOption{IsRequired: true, Name: "Custom"}),
				simpleOption("Other", "other", "int", TagOption{Name: "Renamed"}),
			}},
			contains:    []string{"Custom string,", "o.field = Custom", "func WithRenamed(opt int)"},
			notContains: []string{"WithOther", "WithField"},
		},
		{
			name: "string_default_with_special_characters",
			spec: &OptionSpec{Options: []OptionMeta{
				simpleOption("Field", "field", "string", TagOption{Default: "a\"b\\c\n"}),
			}},
			contains: []string{`o.field = "a\"b\\c\n"`},
		},
		{
			name: "validator_with_quote",
			spec: &OptionSpec{Options: []OptionMeta{
				simpleOption("Field", "field", "string", TagOption{GoValidator: `excludesall="`}),
			}},
			contains: []string{`.Var(o.field, "excludesall=\"")`},
		},
		{
			name: "multiline_docstring",
			spec: &OptionSpec{Options: []OptionMeta{
				{Name: "Field", Docstring: "// first\n// second", Field: "field", Type: "string"},
			}},
			contains: []string{"// first\n// second\nfunc WithField(opt string)"},
		},
		{
			name: "variadic_option",
			spec: &OptionSpec{Options: []OptionMeta{
				simpleOption("Items", "items", "string", TagOption{Variadic: true}),
			}},
			contains: []string{"func WithItems(opt ...string)", "o.items = append(o.items, opt...)"},
		},
		{
			name:     "non_identifier_import_is_kept_when_used",
			spec:     &OptionSpec{Options: []OptionMeta{simpleOption("Node", "node", "yaml.Node", TagOption{})}},
			imports:  []Import{{Path: `"gopkg.in/yaml.v3"`}, {Path: `"github.com/mattn/go-isatty"`}},
			contains: []string{`"gopkg.in/yaml.v3"`},
			// isatty is not referenced by any option.
			notContains: []string{"go-isatty"},
		},
		{
			name:    "invalid_identifier_breaks_the_generated_source",
			spec:    &OptionSpec{Options: []OptionMeta{simpleOption("Bad", "bad field", "string", TagOption{})}},
			wantErr: "cannot optimize generated source",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			rendered, err := renderCornerSpec(t, testCase.spec, testCase.imports, testCase.extra...)
			if testCase.wantErr != "" {
				require.ErrorContains(t, err, testCase.wantErr)

				return
			}

			require.NoError(t, err)
			for _, substr := range testCase.contains {
				require.Contains(t, rendered, substr)
			}

			for _, substr := range testCase.notContains {
				require.NotContains(t, rendered, substr)
			}

			require.True(t, strings.HasPrefix(rendered, "// Code generated by options-gen test. DO NOT EDIT."))
		})
	}
}

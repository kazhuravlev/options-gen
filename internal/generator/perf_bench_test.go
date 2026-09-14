//nolint:testpackage,varnamelen
package generator

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"regexp"
	"testing"
)

var (
	benchPruneUnusedImportsSink []byte
	benchOptimizeSourceSink     []byte
	benchNormalizeSink          string
	benchRenderExprSink         string
	benchIsPublicSink           bool
	benchCheckDefaultSink       error
	benchDeleteByIndexSink      []string
)

// BenchmarkOptimizeGeneratedSource benchmarks the optimization of generated source code
func BenchmarkOptimizeGeneratedSource(b *testing.B) {
	// Prepare a typical generated source with unused imports
	testSource := []byte(`package testcase

import (
	"fmt"
	"io"
	"strings"
	"time"
)

type Options struct {
	field1 string
	field2 int
}

func (o Options) String() string {
	return fmt.Sprintf("Options{field1=%s, field2=%d}", o.field1, o.field2)
}

func WithField1(v string) OptOptionsSetter {
	return func(o *Options) {
		o.field1 = v
	}
}
`)

	b.ReportAllocs()
	b.ResetTimer()

	var err error
	for b.Loop() {
		benchOptimizeSourceSink, err = optimizeGeneratedSource(testSource)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkPruneUnusedImports benchmarks the import pruning logic
func BenchmarkPruneUnusedImports(b *testing.B) {
	fset := token.NewFileSet()

	// Setup: parse a file with multiple unused imports
	source := `package testcase

import (
	"fmt"
	"io"
	"strings"
	"time"
	"bytes"
)

type Options struct {
	field string
}

func Test() string {
	return fmt.Sprintf("%s", "test")
}`

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		// Create a fresh copy of the file for each iteration
		fileCopy, _ := parser.ParseFile(fset, "", []byte(source), parser.ParseComments)
		pruneUnusedImports(fileCopy)
	}
}

// BenchmarkNormalizeTypeName benchmarks type name normalization
func BenchmarkNormalizeTypeName(b *testing.B) {
	testCases := []string{
		"int",
		"string",
		"*MyType",
		"[]string",
		"*[]string",
		"github.com/example/pkg.MyType",
		"*github.com/example/pkg.MyType",
		"[]github.com/example/pkg.MyType",
	}

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		for _, tc := range testCases {
			benchNormalizeSink = normalizeTypeName(tc)
		}
	}
}

// BenchmarkRenderExprString benchmarks AST expression rendering
func BenchmarkRenderExprString(b *testing.B) {
	fset := token.NewFileSet()

	// Parse to get sample expressions
	source := `package test
type M map[string]int
type C chan string
type F func(string) error
type I interface{ Read([]byte) (int, error) }
var x *int
var y []string`

	file, err := parser.ParseFile(fset, "", []byte(source), 0)
	if err != nil {
		b.Fatal(err)
	}

	// Extract various expressions from the parsed AST
	var exprs []ast.Expr
	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}
		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if ok {
				exprs = append(exprs, typeSpec.Type)
			}
		}
	}

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		for _, expr := range exprs {
			benchRenderExprSink = renderExprString(expr)
		}
	}
}

// BenchmarkIsPublic benchmarks the public/private field name check
func BenchmarkIsPublic(b *testing.B) {
	testCases := []string{
		"Field",
		"field",
		"Field1",
		"field1",
		"FIELD",
		"_private",
		"_Field",
		"FieldName",
		"a",
		"A",
		"日本語",
		"_",
	}

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		for _, tc := range testCases {
			benchIsPublicSink = isPublic(tc)
		}
	}
}

// BenchmarkCheckDefaultValue benchmarks default value validation
func BenchmarkCheckDefaultValue(b *testing.B) {
	testCases := []struct {
		fieldType string
		value     string
	}{
		{"int", "42"},
		{"int64", "-1"},
		{"uint", "100"},
		{"float32", "3.14"},
		{"float64", "2.718"},
		{"bool", "true"},
		{"bool", "false"},
		{"string", "hello"},
		{"time.Duration", "1s"},
		{"time.Duration", "1m30s"},
	}

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		for _, tc := range testCases {
			benchCheckDefaultSink = checkDefaultValue(tc.fieldType, tc.value)
		}
	}
}

// BenchmarkDeleteByIndex benchmarks slice element deletion
func BenchmarkDeleteByIndex(b *testing.B) {
	b.Run("small_slice", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()

		for b.Loop() {
			data := []string{"a", "b", "c", "d", "e"}
			benchDeleteByIndexSink = deleteByIndex(data, 2)
		}
	})

	b.Run("medium_slice", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()

		for b.Loop() {
			data := make([]string, 100)
			for i := range data {
				data[i] = fmt.Sprintf("item_%d", i)
			}
			benchDeleteByIndexSink = deleteByIndex(data, 50)
		}
	})

	b.Run("large_slice", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()

		for b.Loop() {
			data := make([]string, 10000)
			for i := range data {
				data[i] = fmt.Sprintf("item_%d", i)
			}
			benchDeleteByIndexSink = deleteByIndex(data, 5000)
		}
	})
}

// BenchmarkMergeImportSpecs benchmarks import spec merging
func BenchmarkMergeImportSpecs(b *testing.B) {
	fset := token.NewFileSet()
	source1 := `package test; import ("fmt"; "strings"; "bytes")`
	source2 := `package test; import ("fmt"; "io"; "os")`

	file1, _ := parser.ParseFile(fset, "", []byte(source1), 0)
	file2, _ := parser.ParseFile(fset, "", []byte(source2), 0)

	var imports1, imports2 []*ast.ImportSpec
	for _, decl := range file1.Decls {
		if genDecl, ok := decl.(*ast.GenDecl); ok && genDecl.Tok == token.IMPORT {
			for _, spec := range genDecl.Specs {
				imports1 = append(imports1, spec.(*ast.ImportSpec))
			}
		}
	}
	for _, decl := range file2.Decls {
		if genDecl, ok := decl.(*ast.GenDecl); ok && genDecl.Tok == token.IMPORT {
			for _, spec := range genDecl.Specs {
				imports2 = append(imports2, spec.(*ast.ImportSpec))
			}
		}
	}

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		_ = mergeImportSpecs(imports1, imports2)
	}
}

// BenchmarkApplyExcludesWithRegex benchmarks regex-based option exclusion
func BenchmarkApplyExcludesWithRegex(b *testing.B) {
	specSize := 50
	options := make([]OptionMeta, specSize)
	for i := range specSize {
		options[i] = OptionMeta{
			Name:  fmt.Sprintf("Option%d", i),
			Field: fmt.Sprintf("option%d", i),
			Type:  "string",
		}
	}

	excludePatterns := []*regexp.Regexp{
		regexp.MustCompile("^Option(1|2|3)$"),
		regexp.MustCompile(".*5$"),
	}

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		ApplyExcludes(options, excludePatterns)
	}
}

// BenchmarkImportSpecName benchmarks import spec name extraction
func BenchmarkImportSpecName(b *testing.B) {
	fset := token.NewFileSet()
	source := `package test
import (
	"fmt"
	f "fmt"
	. "math"
	_ "database/sql/driver"
	customAlias "github.com/kazhuravlev/options-gen/pkg"
)`

	file, _ := parser.ParseFile(fset, "", []byte(source), 0)

	var specs []*ast.ImportSpec
	for _, decl := range file.Decls {
		if genDecl, ok := decl.(*ast.GenDecl); ok && genDecl.Tok == token.IMPORT {
			for _, spec := range genDecl.Specs {
				specs = append(specs, spec.(*ast.ImportSpec))
			}
		}
	}

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		for _, spec := range specs {
			_ = importSpecName(spec)
		}
	}
}

// BenchmarkMakeTemplateOptions benchmarks template option preparation
func BenchmarkMakeTemplateOptions(b *testing.B) {
	options := make([]OptionMeta, 0, 50)
	for i := range 50 {
		options = append(options, OptionMeta{
			Name:      fmt.Sprintf("Option%d", i),
			Docstring: fmt.Sprintf("// Option %d", i),
			Field:     fmt.Sprintf("option%d", i),
			Type:      "string",
			TagOption: TagOption{
				Name:       fmt.Sprintf("opt%d", i),
				IsRequired: i%3 == 0,
			},
		})
	}

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		makeTemplateOptions(options)
	}
}

// BenchmarkRenderSmallSpec benchmarks rendering with small spec (adds coverage for edge cases)
func BenchmarkRenderSmallSpec(b *testing.B) {
	_ = filepath.Join("..", "..", "options-gen", "testdata", "case-02-builtin-types", "options.go")
	spec := &OptionSpec{
		TypeParamsSpec: "",
		TypeParams:     "",
		Options: []OptionMeta{
			{
				Name:      "Field1",
				Docstring: "// A single field",
				Field:     "field1",
				Type:      "string",
				TagOption: TagOption{
					IsRequired: false,
					Default:    "default",
				},
			},
		},
	}

	opts := NewOptions(
		WithVersion("bench"),
		WithPackageName("testcase"),
		WithOptionsStructName("Options"),
		WithFileImports(nil),
		WithSpec(spec),
		WithTagName("default"),
		WithConstructorTypeRender("public"),
		WithOptionTypeName("OptOptionsSetter"),
	)

	b.ReportAllocs()
	b.ResetTimer()

	var err error
	for b.Loop() {
		_, err = Render(opts)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkExtractFields benchmarks field list extraction
func BenchmarkExtractFields(b *testing.B) {
	fset := token.NewFileSet()
	source := `package test
	type S struct {
		Field1 string
		Field2 int
		Field3 bool
		Field4 []string
		Field5 map[string]int
		Field6 *MyType
		Field7 interface{}
		Field8 chan int
		Field9 func(string) error
		Field10 MyInterface
	}`

	file, _ := parser.ParseFile(fset, "", []byte(source), 0)
	var fieldLists []*ast.FieldList
	for _, decl := range file.Decls {
		if genDecl, ok := decl.(*ast.GenDecl); ok {
			for _, spec := range genDecl.Specs {
				if typeSpec, ok := spec.(*ast.TypeSpec); ok {
					if structType, ok := typeSpec.Type.(*ast.StructType); ok {
						fieldLists = append(fieldLists, structType.Fields)
					}
				}
			}
		}
	}

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		for _, fl := range fieldLists {
			extractFields(fl)
		}
	}
}

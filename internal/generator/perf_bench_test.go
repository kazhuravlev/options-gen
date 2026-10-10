//nolint:testpackage
package generator

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strconv"
	"testing"
)

var (
	benchFormatSourceSink    []byte
	benchNormalizeSink       string
	benchRenderExprSink      string
	benchIsPublicSink        bool
	errBenchCheckDefaultSink error
	benchDeleteByIndexSink   []string
	benchMergeImportsSink    []*ast.ImportSpec
	benchImportSpecNameSink  string
	benchTemplateOptionsSink []templateOptionMeta
	benchRenderSmallSink     []byte
)

// BenchmarkFormatGeneratedSource benchmarks the whole post-processing pipeline
// (parse, prune imports, render the import block, gofmt) on a tiny hand-written
// source. See BenchmarkRenderStages/format_source for realistic rendered inputs.
func BenchmarkFormatGeneratedSource(b *testing.B) {
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

	var err error
	for b.Loop() {
		benchFormatSourceSink, err = formatGeneratedSource(testSource)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkNormalizeTypeName benchmarks type name normalization (eight inputs per op).
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

	for b.Loop() {
		for _, tc := range testCases {
			benchNormalizeSink = normalizeTypeName(tc)
		}
	}
}

// BenchmarkRenderExprString benchmarks AST expression rendering (six expressions per op).
func BenchmarkRenderExprString(b *testing.B) {
	source := `package test
type M map[string]int
type C chan string
type F func(string) error
type I interface{ Read([]byte) (int, error) }
var x *int
var y []string`

	file, err := parser.ParseFile(token.NewFileSet(), "", []byte(source), 0)
	if err != nil {
		b.Fatal(err)
	}

	var exprs []ast.Expr
	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}

		for _, spec := range genDecl.Specs {
			if typeSpec, ok := spec.(*ast.TypeSpec); ok {
				exprs = append(exprs, typeSpec.Type)
			}
		}
	}

	b.ReportAllocs()

	for b.Loop() {
		for _, expr := range exprs {
			benchRenderExprSink = renderExprString(expr)
		}
	}
}

// BenchmarkIsPublic benchmarks the public/private field name check (twelve names per op).
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

	for b.Loop() {
		for _, tc := range testCases {
			benchIsPublicSink = isPublic(tc)
		}
	}
}

// BenchmarkCheckDefaultValue benchmarks default value validation (ten values per op).
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

	for b.Loop() {
		for _, tc := range testCases {
			errBenchCheckDefaultSink = checkDefaultValue(tc.fieldType, tc.value)
		}
	}
}

// BenchmarkDeleteByIndex benchmarks slice element deletion. deleteByIndex shifts the
// tail in place, so every iteration first restores the slice from a pristine copy;
// that copy (a memmove of the same order as the deletion itself) is part of the
// measured time, while the element strings are built once outside the loop.
func BenchmarkDeleteByIndex(b *testing.B) {
	run := func(name string, size, index int) {
		b.Run(name, func(b *testing.B) {
			pristine := make([]string, size)
			for i := range pristine {
				pristine[i] = "item_" + strconv.Itoa(i)
			}

			data := make([]string, size)

			b.ReportAllocs()

			for b.Loop() {
				copy(data, pristine)
				benchDeleteByIndexSink = deleteByIndex(data, index)
			}

			if len(benchDeleteByIndexSink) != size-1 {
				b.Fatalf("unexpected result length %d", len(benchDeleteByIndexSink))
			}
		})
	}

	run("small_slice", 5, 2)
	run("medium_slice", 100, 50)
	run("large_slice", 10000, 5000)
}

// BenchmarkMergeImportSpecs benchmarks import spec merging of two overlapping groups.
func BenchmarkMergeImportSpecs(b *testing.B) {
	source1 := `package test; import ("fmt"; "strings"; "bytes")`
	source2 := `package test; import ("fmt"; "io"; "os")`

	imports1 := parseImportSpecs(b, source1)
	imports2 := parseImportSpecs(b, source2)

	b.ReportAllocs()

	for b.Loop() {
		benchMergeImportsSink = mergeImportSpecs(imports1, imports2)
	}

	if len(benchMergeImportsSink) != 5 {
		b.Fatalf("unexpected merged imports count %d", len(benchMergeImportsSink))
	}
}

// BenchmarkImportSpecName benchmarks import name resolution (five specs per op:
// plain path, alias, dot import, blank import and a deep aliased path).
func BenchmarkImportSpecName(b *testing.B) {
	source := `package test
import (
	"fmt"
	f "fmt"
	. "math"
	_ "database/sql/driver"
	customAlias "github.com/kazhuravlev/options-gen/pkg"
)`

	specs := parseImportSpecs(b, source)

	b.ReportAllocs()

	for b.Loop() {
		for _, spec := range specs {
			benchImportSpecNameSink = importSpecName(spec)
		}
	}
}

// BenchmarkMakeTemplateOptions benchmarks template option preparation for 50 options.
func BenchmarkMakeTemplateOptions(b *testing.B) {
	options := make([]OptionMeta, 0, 50)
	for i := range 50 {
		options = append(options, OptionMeta{
			Name:      "Option" + strconv.Itoa(i),
			Docstring: "// Option " + strconv.Itoa(i),
			Field:     "option" + strconv.Itoa(i),
			Type:      "string",
			TagOption: TagOption{
				IsRequired:    i%3 == 0,
				GoValidator:   "",
				Default:       "",
				Variadic:      false,
				VariadicIsSet: false,
				Skip:          false,
				Name:          "opt" + strconv.Itoa(i),
			},
		})
	}

	b.ReportAllocs()

	for b.Loop() {
		benchTemplateOptionsSink = makeTemplateOptions(options)
	}
}

// BenchmarkRenderSmallSpec benchmarks Render for a one-field spec with a tag default
// and without validators, i.e. the smallest realistic output.
func BenchmarkRenderSmallSpec(b *testing.B) {
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
					IsRequired:    false,
					GoValidator:   "",
					Default:       "default",
					Variadic:      false,
					VariadicIsSet: false,
					Skip:          false,
					Name:          "",
				},
			},
		},
	}

	opts := benchmarkRenderOptions(spec)

	b.ReportAllocs()

	var err error
	for b.Loop() {
		benchRenderSmallSink, err = Render(opts)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func parseImportSpecs(b *testing.B, source string) []*ast.ImportSpec {
	b.Helper()

	file, err := parser.ParseFile(token.NewFileSet(), "", []byte(source), 0)
	if err != nil {
		b.Fatal(err)
	}

	var specs []*ast.ImportSpec
	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.IMPORT {
			continue
		}

		for _, spec := range genDecl.Specs {
			importSpec, ok := spec.(*ast.ImportSpec)
			if !ok {
				b.Fatalf("unexpected spec %T", spec)
			}

			specs = append(specs, importSpec)
		}
	}

	return specs
}

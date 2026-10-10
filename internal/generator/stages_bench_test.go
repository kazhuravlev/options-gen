//nolint:testpackage
package generator

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"testing"

	"golang.org/x/tools/imports"
)

var (
	benchStageTemplateSink []byte
	benchStageOptimizeSink []byte
	benchStageParseSink    *ast.File
	benchStageFormatSink   []byte
	benchStageImportsSink  []byte
)

// benchmarkExecuteTemplate mirrors the template stage of Render: it builds the
// template context and executes the template, returning the unformatted source.
// BenchmarkRenderStages verifies once per spec that optimizeGeneratedSource of this
// output equals Render's output, so a drift between this copy and Render is caught.
func benchmarkExecuteTemplate(opts Options) ([]byte, error) {
	optionsStructType := opts.optionsStructName
	optionsStructInstanceType := opts.optionsStructName

	if opts.spec.TypeParamsSpec != "" {
		optionsStructType += opts.spec.TypeParamsSpec
		optionsStructInstanceType += opts.spec.TypeParams
	}

	options := makeTemplateOptions(opts.spec.Options)
	tplContext := map[string]any{
		"version":       opts.version,
		"packageName":   opts.packageName,
		"imports":       opts.fileImports,
		"options":       options,
		"optionsLen":    len(options),
		"hasValidation": opts.spec.HasValidation(),

		"optionsTypeParamsSpec": opts.spec.TypeParamsSpec,
		"optionsTypeParams":     opts.spec.TypeParams,

		"optionsPrefix":             opts.prefix,
		"optionsStructName":         opts.optionsStructName,
		"optionsStructType":         optionsStructType,
		"optionsStructInstanceType": optionsStructInstanceType,
		"optionsTypeName":           opts.optionTypeName,
		"defaultsTagName":           opts.tagName,
		"defaultsVarName":           opts.varName,
		"defaultsFuncName":          opts.funcName,

		"withIsset": opts.withIsset,

		"constructorTypeRender": opts.constructorTypeRender,
	}

	buf := new(bytes.Buffer)
	if err := tmpl.Execute(buf, tplContext); err != nil {
		return nil, fmt.Errorf("cannot render template: %w", err)
	}

	return buf.Bytes(), nil
}

// benchmarkTemplateSource returns the raw template output for opts after checking
// that the stage split reproduces Render exactly.
func benchmarkTemplateSource(b *testing.B, opts Options) []byte {
	b.Helper()

	src, err := benchmarkExecuteTemplate(opts)
	if err != nil {
		b.Fatal(err)
	}

	optimized, err := optimizeGeneratedSource(src)
	if err != nil {
		b.Fatal(err)
	}

	rendered, err := Render(opts)
	if err != nil {
		b.Fatal(err)
	}

	if !bytes.Equal(optimized, rendered) {
		b.Fatal("benchmarkExecuteTemplate diverged from Render: update the template context copy")
	}

	return src
}

// BenchmarkRenderStages splits Render into its two stages for the same specs as
// BenchmarkRenderCriticalPath: template_execute is context building plus
// text/template execution, optimize_source is optimizeGeneratedSource (parse, prune
// imports, sort imports, format.Node, imports.Process) on that template output.
func BenchmarkRenderStages(b *testing.B) {
	for _, optionCount := range []int{10, 50, 100} {
		opts := benchmarkRenderOptions(benchmarkOptionSpec(optionCount))
		src := benchmarkTemplateSource(b, opts)

		b.Run(fmt.Sprintf("template_execute/%d_fields", optionCount), func(b *testing.B) {
			b.ReportAllocs()

			var err error
			for b.Loop() {
				benchStageTemplateSink, err = benchmarkExecuteTemplate(opts)
				if err != nil {
					b.Fatal(err)
				}
			}
			b.SetBytes(int64(len(benchStageTemplateSink)))
		})

		b.Run(fmt.Sprintf("optimize_source/%d_fields", optionCount), func(b *testing.B) {
			b.ReportAllocs()

			var err error
			for b.Loop() {
				benchStageOptimizeSink, err = optimizeGeneratedSource(src)
				if err != nil {
					b.Fatal(err)
				}
			}
			b.SetBytes(int64(len(src)))
		})
	}
}

// BenchmarkOptimizeStages measures each step of optimizeGeneratedSource separately on
// the raw template output of the 50-field spec: parse (parser.ParseFile with
// comments), prune_imports (pruneUnusedImports on the parsed file, restored before
// every iteration), sort_and_format (ast.SortImports plus format.Node) and
// imports_process (imports.Process in FormatOnly mode on the formatted source).
func BenchmarkOptimizeStages(b *testing.B) {
	const optionCount = 50

	src := benchmarkTemplateSource(b, benchmarkRenderOptions(benchmarkOptionSpec(optionCount)))

	b.Run("parse", func(b *testing.B) {
		b.ReportAllocs()
		b.SetBytes(int64(len(src)))

		var err error
		for b.Loop() {
			benchStageParseSink, err = parser.ParseFile(token.NewFileSet(), "", src, parser.ParseComments)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("prune_imports", func(b *testing.B) {
		file, err := parser.ParseFile(token.NewFileSet(), "", src, parser.ParseComments)
		if err != nil {
			b.Fatal(err)
		}

		snap := snapshotImports(file)

		b.ReportAllocs()

		for b.Loop() {
			snap.restore()
			pruneUnusedImports(file)
		}
	})

	b.Run("sort_and_format", func(b *testing.B) {
		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, "", src, parser.ParseComments)
		if err != nil {
			b.Fatal(err)
		}

		pruneUnusedImports(file)

		b.ReportAllocs()

		for b.Loop() {
			ast.SortImports(fset, file)

			var buf bytes.Buffer
			if err := format.Node(&buf, fset, file); err != nil {
				b.Fatal(err)
			}

			benchStageFormatSink = buf.Bytes()
		}
		b.SetBytes(int64(len(benchStageFormatSink)))
	})

	b.Run("imports_process", func(b *testing.B) {
		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, "", src, parser.ParseComments)
		if err != nil {
			b.Fatal(err)
		}

		pruneUnusedImports(file)
		ast.SortImports(fset, file)

		var buf bytes.Buffer
		if err := format.Node(&buf, fset, file); err != nil {
			b.Fatal(err)
		}

		formatted := buf.Bytes()

		b.ReportAllocs()
		b.SetBytes(int64(len(formatted)))

		for b.Loop() {
			benchStageImportsSink, err = imports.Process("", formatted, &imports.Options{
				Fragment:   false,
				AllErrors:  false,
				Comments:   true,
				TabIndent:  true,
				TabWidth:   generatedFormatTabWidth,
				FormatOnly: true,
			})
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

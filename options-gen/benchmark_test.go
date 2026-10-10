//nolint:testpackage,varnamelen
package optionsgen

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kazhuravlev/options-gen/internal/ctype"
)

// BenchmarkRunCriticalPath measures the complete Run on testdata cases: reading and
// parsing the case directory, building the spec, rendering and formatting the code
// and writing options_generated.go into a temp dir. The file write (a few KB) is
// part of every iteration on purpose, since that is what the CLI does.
func BenchmarkRunCriticalPath(b *testing.B) {
	benchmarks := []struct {
		name        string
		caseDir     string
		defaults    Defaults
		allVariadic bool
		withIsset   bool
	}{
		{
			name:        "builtin_fields",
			caseDir:     filepath.Join("testdata", "case-02-builtin-types"),
			defaults:    Defaults{From: DefaultsFromNone, Param: ""},
			allVariadic: false,
			withIsset:   false,
		},
		{
			name:        "all_variadic",
			caseDir:     filepath.Join("testdata", "case-02.1-builtin-types-all-variadic"),
			defaults:    Defaults{From: DefaultsFromNone, Param: ""},
			allVariadic: true,
			withIsset:   false,
		},
		{
			name:        "generics",
			caseDir:     filepath.Join("testdata", "case-05-generics-01"),
			defaults:    Defaults{From: DefaultsFromNone, Param: ""},
			allVariadic: false,
			withIsset:   false,
		},
		{
			name:    "defaults_from_tag",
			caseDir: filepath.Join("testdata", "case-12-defaults-tag-02"),
			defaults: Defaults{
				From:  DefaultsFromTag,
				Param: "",
			},
			allVariadic: false,
			withIsset:   false,
		},
		{
			name:        "with_isset",
			caseDir:     filepath.Join("testdata", "case-20-isset"),
			defaults:    Defaults{From: DefaultsFromNone, Param: ""},
			allVariadic: false,
			withIsset:   true,
		},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			outFilename := filepath.Join(b.TempDir(), "options_generated.go")
			opts := NewOptions(
				WithVersion("benchmark"),
				WithInFilename(filepath.Join(bm.caseDir, "options.go")),
				WithOutFilename(outFilename),
				WithStructName("Options"),
				WithPackageName("testcase"),
				WithDefaults(bm.defaults),
				WithAllVariadic(bm.allVariadic),
				WithWithIsset(bm.withIsset),
				WithConstructorTypeRender(ConstructorPublicRender),
			)
			b.ReportAllocs()

			for b.Loop() {
				if err := Run(opts); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkRunScaling measures the complete Run for generated structs of 10, 50 and
// 100 builtin-typed fields (validators on every third field, every fifth mandatory).
// The input lives in a temp dir and the output is written next to it, like a real
// `go generate` run, so from the second iteration on parser.ParseDir also parses
// the previously generated file.
func BenchmarkRunScaling(b *testing.B) {
	for _, fieldCount := range []int{10, 50, 100} {
		b.Run(fmt.Sprintf("%d_fields", fieldCount), func(b *testing.B) {
			dir := b.TempDir()
			inFilename := filepath.Join(dir, "options.go")
			if err := os.WriteFile(inFilename, []byte(benchmarkStructSource(fieldCount)), ctype.DefaultPermission); err != nil {
				b.Fatal(err)
			}

			opts := NewOptions(
				WithVersion("benchmark"),
				WithInFilename(inFilename),
				WithOutFilename(filepath.Join(dir, "options_generated.go")),
				WithStructName("Options"),
				WithPackageName("testcase"),
				WithDefaults(Defaults{From: DefaultsFromTag, Param: ""}),
				WithAllVariadic(false),
				WithWithIsset(false),
				WithConstructorTypeRender(ConstructorPublicRender),
			)
			b.ReportAllocs()

			for b.Loop() {
				if err := Run(opts); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// benchmarkStructSource renders a package with an Options struct of fieldCount fields
// of mixed builtin types; it mirrors the generator package's benchmark input.
func benchmarkStructSource(fieldCount int) string {
	fieldTypes := []string{"string", "int", "time.Duration", "bool", "float64", "int64"}

	var builder strings.Builder
	builder.WriteString("package testcase\n\nimport \"time\"\n\nvar _ time.Duration\n\ntype Options struct {\n")

	for i := range fieldCount {
		if i%2 == 0 {
			fmt.Fprintf(&builder, "\t// Field%d holds benchmark value number %d.\n", i, i)
		}

		fmt.Fprintf(&builder, "\tfield%d %s", i, fieldTypes[i%len(fieldTypes)])

		switch {
		case i%5 == 0:
			builder.WriteString(" `option:\"mandatory\"`")
		case i%3 == 0:
			builder.WriteString(" `validate:\"required\"`")
		}

		builder.WriteString("\n")
	}

	builder.WriteString("}\n")

	return builder.String()
}

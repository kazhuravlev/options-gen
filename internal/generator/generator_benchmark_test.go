//nolint:testpackage,varnamelen
package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/kazhuravlev/options-gen/internal/ctype"
)

var (
	benchmarkSpecSink   *GetOptionSpecRes
	benchmarkRenderSink []byte
)

const (
	benchmarkTypeInt           = "int"
	benchmarkValidatorRequired = "required"
	benchmarkVariantFieldCount = 50
)

// BenchmarkGetOptionSpecCriticalPath measures GetOptionSpec on the testdata cases of
// the options-gen package. Every iteration re-reads the case directory from disk
// (parser.ParseDir parses options.go and the previously generated
// options_generated.go), which is exactly what a `go generate` run does.
func BenchmarkGetOptionSpecCriticalPath(b *testing.B) {
	benchmarks := []struct {
		name        string
		filePath    string
		structName  string
		tagName     string
		allVariadic bool
	}{
		{
			name:        "builtin_fields",
			filePath:    filepath.Join("..", "..", "options-gen", "testdata", "case-02-builtin-types", "options.go"),
			structName:  "Options",
			tagName:     "default",
			allVariadic: false,
		},
		{
			name:        "generics",
			filePath:    filepath.Join("..", "..", "options-gen", "testdata", "case-05-generics-01", "options.go"),
			structName:  "Options",
			tagName:     "default",
			allVariadic: false,
		},
		{
			name: "all_variadic",
			filePath: filepath.Join(
				"..", "..", "options-gen", "testdata", "case-02.1-builtin-types-all-variadic", "options.go",
			),
			structName:  "Options",
			tagName:     "default",
			allVariadic: true,
		},
		{
			name:        "imported_alias_struct",
			filePath:    filepath.Join("..", "..", "options-gen", "testdata", "case-05.2-generics-01-alias", "options.go"),
			structName:  "Options",
			tagName:     "default",
			allVariadic: false,
		},
		{
			name:        "defaults_duration",
			filePath:    filepath.Join("..", "..", "options-gen", "testdata", "case-12-defaults-tag-02", "options.go"),
			structName:  "Options",
			tagName:     "default",
			allVariadic: false,
		},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			b.ReportAllocs()

			var err error
			for b.Loop() {
				benchmarkSpecSink, err = GetOptionSpec(
					bm.filePath,
					bm.structName,
					bm.tagName,
					bm.allVariadic,
					nil,
				)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkRenderCriticalPath measures Render (template execution plus
// formatGeneratedSource) for specs of a growing number of string options, all of
// them with a `required` validator and every fourth one mandatory.
func BenchmarkRenderCriticalPath(b *testing.B) {
	for _, optionCount := range []int{1, 10, 50, 100} {
		b.Run(fmt.Sprintf("%d_fields", optionCount), func(b *testing.B) {
			spec := benchmarkOptionSpec(optionCount)
			opts := benchmarkRenderOptions(spec)
			b.ReportAllocs()

			var err error
			for b.Loop() {
				benchmarkRenderSink, err = Render(opts)
				if err != nil {
					b.Fatal(err)
				}
			}
			b.SetBytes(int64(len(benchmarkRenderSink)))
		})
	}
}

// BenchmarkGenerateCriticalPath measures GetOptionSpec followed by Render for the
// builtin-types case, i.e. everything optionsgen.Run does except writing the file.
func BenchmarkGenerateCriticalPath(b *testing.B) {
	filePath := filepath.Join("..", "..", "options-gen", "testdata", "case-02-builtin-types", "options.go")
	b.ReportAllocs()

	var err error
	for b.Loop() {
		benchmarkSpecSink, err = GetOptionSpec(filePath, "Options", "default", false, nil)
		if err != nil {
			b.Fatal(err)
		}

		benchmarkRenderSink, err = Render(benchmarkRenderOptions(&benchmarkSpecSink.Spec))
		if err != nil {
			b.Fatal(err)
		}
	}
	b.SetBytes(int64(len(benchmarkRenderSink)))
}

// BenchmarkGetOptionSpecFields measures how GetOptionSpec scales with the number of
// struct fields. The options file is generated into a temp dir that contains nothing
// else, so the cost is parsing one file plus the per-field work.
func BenchmarkGetOptionSpecFields(b *testing.B) {
	for _, fieldCount := range []int{10, 50, 100} {
		b.Run(fmt.Sprintf("%d_fields", fieldCount), func(b *testing.B) {
			dir := b.TempDir()
			filePath := filepath.Join(dir, "options.go")
			benchmarkWriteFile(b, filePath, benchmarkStructSource(fieldCount, benchmarkMixedTag))

			b.ReportAllocs()

			var err error
			for b.Loop() {
				benchmarkSpecSink, err = GetOptionSpec(filePath, "Options", "default", false, nil)
				if err != nil {
					b.Fatal(err)
				}
			}

			if len(benchmarkSpecSink.Spec.Options) != fieldCount {
				b.Fatalf("expected %d options, got %d", fieldCount, len(benchmarkSpecSink.Spec.Options))
			}
		})
	}
}

// BenchmarkGetOptionSpecDirectory measures the cost of parser.ParseDir parsing the
// whole package directory. Every variant has a 50-field options.go next to the
// options_generated.go rendered from it (as in a real package after the first
// `go generate`), plus N filler files of roughly 3 KB each.
func BenchmarkGetOptionSpecDirectory(b *testing.B) {
	const fieldCount = 50

	for _, extraFiles := range []int{0, 5, 20} {
		b.Run(fmt.Sprintf("%d_extra_files", extraFiles), func(b *testing.B) {
			dir := b.TempDir()
			filePath := filepath.Join(dir, "options.go")
			benchmarkWriteFile(b, filePath, benchmarkStructSource(fieldCount, benchmarkMixedTag))

			spec, err := GetOptionSpec(filePath, "Options", "default", false, nil)
			if err != nil {
				b.Fatal(err)
			}

			generated, err := Render(benchmarkRenderOptions(&spec.Spec))
			if err != nil {
				b.Fatal(err)
			}

			benchmarkWriteFile(b, filepath.Join(dir, "options_generated.go"), string(generated))

			for i := range extraFiles {
				benchmarkWriteFile(b, filepath.Join(dir, fmt.Sprintf("file_%02d.go", i)), benchmarkFillerSource(i))
			}

			b.ReportAllocs()

			for b.Loop() {
				benchmarkSpecSink, err = GetOptionSpec(filePath, "Options", "default", false, nil)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkGetOptionSpecVariants measures GetOptionSpec on 50-field structs that
// exercise specific code paths: exclude patterns, a validate/default/option tag on
// every field, all-variadic over builtin types, and all-variadic over a struct that
// imports a std package (which makes extractSliceElemType call packages.Load, i.e.
// run `go list`, on every call because the PackageStore is created per call).
func BenchmarkGetOptionSpecVariants(b *testing.B) {
	const fieldCount = 50

	b.Run("excludes", func(b *testing.B) {
		dir := b.TempDir()
		filePath := filepath.Join(dir, "options.go")
		benchmarkWriteFile(b, filePath, benchmarkStructSource(fieldCount, benchmarkMixedTag))

		excludes := []*regexp.Regexp{
			regexp.MustCompile(`^Field1\d$`),
			regexp.MustCompile(`^Field2[0-4]$`),
			regexp.MustCompile(`^Internal`),
		}

		b.ReportAllocs()

		var err error
		for b.Loop() {
			benchmarkSpecSink, err = GetOptionSpec(filePath, "Options", "default", false, excludes)
			if err != nil {
				b.Fatal(err)
			}
		}

		if got := len(benchmarkSpecSink.Spec.Options); got != fieldCount-15 {
			b.Fatalf("expected %d options after excludes, got %d", fieldCount-15, got)
		}
	})

	b.Run("all_tags", func(b *testing.B) {
		dir := b.TempDir()
		filePath := filepath.Join(dir, "options.go")
		benchmarkWriteFile(b, filePath, benchmarkStructSource(fieldCount, benchmarkFullTag))

		b.ReportAllocs()

		var err error
		for b.Loop() {
			benchmarkSpecSink, err = GetOptionSpec(filePath, "Options", "default", false, nil)
			if err != nil {
				b.Fatal(err)
			}
		}

		for _, opt := range benchmarkSpecSink.Spec.Options {
			if opt.TagOption.Default == "" || opt.TagOption.GoValidator == "" || opt.TagOption.Name == "" {
				b.Fatalf("tags of %s were not parsed: %+v", opt.Field, opt.TagOption)
			}
		}
	})

	b.Run("all_variadic_builtin", func(b *testing.B) {
		dir := b.TempDir()
		filePath := filepath.Join(dir, "options.go")
		benchmarkWriteFile(b, filePath, benchmarkStructSourceOf(fieldCount, benchmarkVariadicTypes, benchmarkMixedTag))

		b.ReportAllocs()

		var err error
		for b.Loop() {
			benchmarkSpecSink, err = GetOptionSpec(filePath, "Options", "default", true, nil)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("all_variadic_std_import", func(b *testing.B) {
		if testing.Short() {
			b.Skip("runs `go list` through packages.Load on every iteration (hundreds of ms)")
		}

		filePath := filepath.Join("..", "..", "options-gen", "testdata", "case-19.2-all-variadic", "options.go")

		b.ReportAllocs()

		var err error
		for b.Loop() {
			benchmarkSpecSink, err = GetOptionSpec(filePath, "Options", "default", true, nil)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// BenchmarkRenderVariants measures Render for 50-option specs that enable one feature
// of the template each, so the cost of every feature can be compared against the
// no_validators and validators baselines.
func BenchmarkRenderVariants(b *testing.B) {
	plain := benchmarkVariant{validators: false, generics: false, variadic: false, tagDefaults: false}

	variants := []struct {
		name string
		opts Options
	}{
		{
			name: "no_validators",
			opts: benchmarkVariantOptions(benchmarkVariantSpec(plain)),
		},
		{
			name: "validators",
			opts: benchmarkVariantOptions(benchmarkVariantSpec(benchmarkVariant{
				validators: true, generics: false, variadic: false, tagDefaults: false,
			})),
		},
		{
			name: "generics",
			opts: benchmarkVariantOptions(benchmarkVariantSpec(benchmarkVariant{
				validators: false, generics: true, variadic: false, tagDefaults: false,
			})),
		},
		{
			name: "with_isset",
			opts: benchmarkVariantOptions(benchmarkVariantSpec(plain), WithWithIsset(true)),
		},
		{
			name: "all_variadic",
			opts: benchmarkVariantOptions(benchmarkVariantSpec(benchmarkVariant{
				validators: false, generics: false, variadic: true, tagDefaults: false,
			})),
		},
		{
			name: "defaults_tag",
			opts: benchmarkVariantOptions(benchmarkVariantSpec(benchmarkVariant{
				validators: false, generics: false, variadic: false, tagDefaults: true,
			})),
		},
		{
			name: "defaults_var",
			opts: benchmarkVariantOptions(
				benchmarkVariantSpec(plain),
				WithTagName(""),
				WithVarName("defaultOptions"),
			),
		},
		{
			name: "defaults_func",
			opts: benchmarkVariantOptions(
				benchmarkVariantSpec(plain),
				WithTagName(""),
				WithFuncName("getDefaultOptions"),
			),
		},
	}

	for _, variant := range variants {
		b.Run(variant.name, func(b *testing.B) {
			b.ReportAllocs()

			var err error
			for b.Loop() {
				benchmarkRenderSink, err = Render(variant.opts)
				if err != nil {
					b.Fatal(err)
				}
			}
			b.SetBytes(int64(len(benchmarkRenderSink)))
		})
	}
}

func benchmarkRenderOptions(spec *OptionSpec) Options {
	return NewOptions(
		WithVersion("benchmark"),
		WithPackageName("testcase"),
		WithOptionsStructName("Options"),
		WithFileImports(nil),
		WithSpec(spec),
		WithTagName("default"),
		WithConstructorTypeRender("public"),
		WithOptionTypeName("OptOptionsSetter"),
	)
}

// benchmarkVariantOptions is benchmarkRenderOptions with the "time" import of the
// input file (required by the time.Duration fields of benchmarkVariantSpec) and the
// given extra setters applied last.
func benchmarkVariantOptions(spec *OptionSpec, extra ...OptOptionsSetter) Options {
	setters := []OptOptionsSetter{
		WithVersion("benchmark"),
		WithPackageName("testcase"),
		WithOptionsStructName("Options"),
		WithFileImports([]Import{{Path: `"time"`, Alias: nil}}),
		WithSpec(spec),
		WithTagName("default"),
		WithConstructorTypeRender("public"),
		WithOptionTypeName("OptOptionsSetter"),
	}

	return NewOptions(append(setters, extra...)...)
}

func benchmarkOptionSpec(optionCount int) *OptionSpec {
	options := make([]OptionMeta, 0, optionCount)
	for i := range optionCount {
		field := fmt.Sprintf("field%d", i)
		opt := OptionMeta{
			Name:      fmt.Sprintf("Field%d", i),
			Docstring: fmt.Sprintf("// Field%d configures benchmark field %d.", i, i),
			Field:     field,
			Type:      "string",
			TagOption: TagOption{
				IsRequired:    false,
				GoValidator:   benchmarkValidatorRequired,
				Default:       "",
				Variadic:      false,
				VariadicIsSet: false,
				Skip:          false,
				Name:          "",
			},
		}
		if i%4 == 0 {
			opt.TagOption.IsRequired = true
		}
		options = append(options, opt)
	}

	return &OptionSpec{
		TypeParamsSpec: "",
		TypeParams:     "",
		Options:        options,
	}
}

// benchmarkVariant selects which template features a spec built by
// benchmarkVariantSpec exercises.
type benchmarkVariant struct {
	validators  bool
	generics    bool
	variadic    bool
	tagDefaults bool
}

// benchmarkFieldTypes are the field types used by generated benchmark structs; the
// list cycles so every struct of ten or more fields contains each of them.
// benchmarkVariadicTypes contains no selector types, so an all-variadic
// GetOptionSpec never has to load a package to resolve a slice element type.
var (
	benchmarkFieldTypes    = []string{"string", benchmarkTypeInt, "time.Duration", "bool", "float64", "int64"}
	benchmarkVariadicTypes = []string{"string", "[]string", benchmarkTypeInt, "[]int", "bool", "[]byte"}
)

// benchmarkDefaultFor returns a default tag value accepted by checkDefaultValue for
// the given builtin type.
func benchmarkDefaultFor(typ string) string {
	switch typ {
	case "string":
		return "value"
	case benchmarkTypeInt, "int64":
		return "42"
	case "time.Duration":
		return "3s"
	case "bool":
		return "true"
	case "float64":
		return "1.5"
	default:
		return ""
	}
}

// benchmarkVariantSpec builds a spec of benchmarkVariantFieldCount options of mixed
// builtin types. Every fourth option is mandatory and every second one has a
// docstring, like real-world option structs.
func benchmarkVariantSpec(variant benchmarkVariant) *OptionSpec {
	options := make([]OptionMeta, 0, benchmarkVariantFieldCount)
	for i := range benchmarkVariantFieldCount {
		typ := benchmarkFieldTypes[i%len(benchmarkFieldTypes)]
		if variant.generics && i%3 == 0 {
			typ = "T"
		}

		opt := OptionMeta{
			Name:      "Field" + strconv.Itoa(i),
			Docstring: "",
			Field:     "field" + strconv.Itoa(i),
			Type:      typ,
			TagOption: TagOption{
				IsRequired:    i%4 == 0,
				GoValidator:   "",
				Default:       "",
				Variadic:      variant.variadic && i%4 != 0,
				VariadicIsSet: variant.variadic,
				Skip:          false,
				Name:          "",
			},
		}

		if i%2 == 0 {
			opt.Docstring = "// Field" + strconv.Itoa(i) + " configures benchmark field " + strconv.Itoa(i) + "."
		}

		if variant.validators {
			opt.TagOption.GoValidator = benchmarkValidatorRequired
		}

		if variant.tagDefaults && !opt.TagOption.IsRequired && typ != "T" {
			opt.TagOption.Default = benchmarkDefaultFor(typ)
		}

		options = append(options, opt)
	}

	spec := &OptionSpec{
		TypeParamsSpec: "",
		TypeParams:     "",
		Options:        options,
	}

	if variant.generics {
		spec.TypeParamsSpec = "[T any]"
		spec.TypeParams = "[T]"
	}

	return spec
}

// benchmarkMixedTag returns the struct tag used by the scaling benchmarks: a
// validator on every third field, mandatory on every fifth and no tag otherwise.
func benchmarkMixedTag(i int, _ string) string {
	switch {
	case i%5 == 0:
		return `option:"mandatory"`
	case i%3 == 0:
		return `validate:"required"`
	default:
		return ""
	}
}

// benchmarkFullTag returns a tag with validate, default and option keys for every field.
func benchmarkFullTag(i int, typ string) string {
	return fmt.Sprintf(`validate:"required" default:"%s" option:"name=Custom%d"`, benchmarkDefaultFor(typ), i)
}

// benchmarkStructSource renders the source of a package with an Options struct of
// fieldCount fields cycling through benchmarkFieldTypes. tagFor returns the struct
// tag (without backticks) for field i of the given type; every second field has a
// doc comment.
func benchmarkStructSource(fieldCount int, tagFor func(i int, typ string) string) string {
	return benchmarkStructSourceOf(fieldCount, benchmarkFieldTypes, tagFor)
}

func benchmarkStructSourceOf(fieldCount int, fieldTypes []string, tagFor func(i int, typ string) string) string {
	var builder strings.Builder
	builder.WriteString("package testcase\n\nimport \"time\"\n\nvar _ time.Duration\n\ntype Options struct {\n")

	for i := range fieldCount {
		typ := fieldTypes[i%len(fieldTypes)]
		if i%2 == 0 {
			fmt.Fprintf(&builder, "\t// Field%d holds benchmark value number %d.\n", i, i)
		}

		fmt.Fprintf(&builder, "\tfield%d %s", i, typ)
		if tag := tagFor(i, typ); tag != "" {
			fmt.Fprintf(&builder, " `%s`", tag)
		}

		builder.WriteString("\n")
	}

	builder.WriteString("}\n")

	return builder.String()
}

// benchmarkFillerSource renders an unrelated ~3 KB Go file (types, constants and
// documented functions) to populate a package directory.
func benchmarkFillerSource(idx int) string {
	var builder strings.Builder
	fmt.Fprintf(&builder, "package testcase\n\nimport \"strconv\"\n\n")
	fmt.Fprintf(&builder, "// Record%d is a filler type used by directory benchmarks.\n", idx)
	fmt.Fprintf(&builder, "type Record%d struct {\n\tID   int\n\tName string\n\tTags []string\n}\n\n", idx)
	fmt.Fprintf(&builder, "const (\n\tlimit%d  = %d\n\tprefix%d = \"record-%d-\"\n)\n\n", idx, idx*10, idx, idx)

	for fn := range 24 {
		fmt.Fprintf(&builder, "// helper%d_%d formats the record id for filler file %d.\n", idx, fn, idx)
		fmt.Fprintf(&builder, "func helper%d_%d(r Record%d) string {\n", idx, fn, idx)
		fmt.Fprintf(&builder, "\tif r.ID > limit%d {\n\t\treturn prefix%d + strconv.Itoa(r.ID)\n\t}\n\n", idx, idx)
		fmt.Fprintf(&builder, "\treturn r.Name + strconv.Itoa(%d)\n}\n\n", fn)
	}

	return builder.String()
}

func benchmarkWriteFile(b *testing.B, filePath, content string) {
	b.Helper()

	if err := os.WriteFile(filePath, []byte(content), ctype.DefaultPermission); err != nil {
		b.Fatal(err)
	}
}

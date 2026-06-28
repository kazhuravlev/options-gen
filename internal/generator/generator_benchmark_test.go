package generator

import (
	"fmt"
	"path/filepath"
	"testing"
)

var (
	benchmarkSpecSink   *GetOptionSpecRes
	benchmarkRenderSink []byte
)

func BenchmarkGetOptionSpecCriticalPath(b *testing.B) {
	benchmarks := []struct {
		name        string
		filePath    string
		structName  string
		tagName     string
		allVariadic bool
	}{
		{
			name:       "builtin_fields",
			filePath:   filepath.Join("..", "..", "options-gen", "testdata", "case-02-builtin-types", "options.go"),
			structName: "Options",
			tagName:    "default",
		},
		{
			name:       "generics",
			filePath:   filepath.Join("..", "..", "options-gen", "testdata", "case-05-generics-01", "options.go"),
			structName: "Options",
			tagName:    "default",
		},
		{
			name:        "all_variadic",
			filePath:    filepath.Join("..", "..", "options-gen", "testdata", "case-02.1-builtin-types-all-variadic", "options.go"),
			structName:  "Options",
			tagName:     "default",
			allVariadic: true,
		},
		{
			name:       "imported_alias_struct",
			filePath:   filepath.Join("..", "..", "options-gen", "testdata", "case-05.2-generics-01-alias", "options.go"),
			structName: "Options",
			tagName:    "default",
		},
		{
			name:       "defaults_duration",
			filePath:   filepath.Join("..", "..", "options-gen", "testdata", "case-12-defaults-tag-02", "options.go"),
			structName: "Options",
			tagName:    "default",
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
				GoValidator: "required",
			},
		}
		if i%4 == 0 {
			opt.TagOption.IsRequired = true
		}
		options = append(options, opt)
	}

	return &OptionSpec{
		Options: options,
	}
}

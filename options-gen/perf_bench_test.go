//nolint:testpackage
package optionsgen

import (
	"testing"
)

var (
	benchOptionsSink     Options
	errBenchValidateSink error
)

// BenchmarkNewOptions benchmarks creating Options with various configurations.
func BenchmarkNewOptions(b *testing.B) {
	b.Run("minimal_options", func(b *testing.B) {
		b.ReportAllocs()

		for b.Loop() {
			benchOptionsSink = NewOptions(
				WithVersion("test"),
				WithInFilename("test.go"),
				WithOutFilename("out.go"),
				WithStructName("Options"),
				WithPackageName("pkg"),
			)
		}
	})

	b.Run("full_options", func(b *testing.B) {
		b.ReportAllocs()

		for b.Loop() {
			benchOptionsSink = NewOptions(
				WithVersion("test"),
				WithInFilename("test.go"),
				WithOutFilename("out.go"),
				WithStructName("Options"),
				WithPackageName("pkg"),
				WithDefaults(Defaults{From: DefaultsFromNone, Param: ""}),
				WithAllVariadic(false),
				WithWithIsset(false),
				WithConstructorTypeRender(ConstructorPublicRender),
				WithShowWarnings(false),
				WithOutOptionTypeName("OptSetter"),
			)
		}
	})
}

// BenchmarkOptionsValidation benchmarks Options.Validate(), which Run calls first.
func BenchmarkOptionsValidation(b *testing.B) {
	opts := NewOptions(
		WithVersion("test"),
		WithInFilename("test.go"),
		WithOutFilename("out.go"),
		WithStructName("Options"),
		WithPackageName("pkg"),
		WithDefaults(Defaults{From: DefaultsFromNone, Param: ""}),
		WithConstructorTypeRender(ConstructorPublicRender),
	)

	b.ReportAllocs()

	for b.Loop() {
		errBenchValidateSink = opts.Validate()
	}

	if errBenchValidateSink != nil {
		b.Fatal(errBenchValidateSink)
	}
}

// BenchmarkWithOptionals benchmarks chaining optional parameters.
func BenchmarkWithOptionals(b *testing.B) {
	b.Run("10_options", func(b *testing.B) {
		b.ReportAllocs()

		for b.Loop() {
			benchOptionsSink = NewOptions(
				WithVersion("v1"),
				WithInFilename("in.go"),
				WithOutFilename("out.go"),
				WithStructName("Opts"),
				WithPackageName("pkg"),
				WithDefaults(Defaults{From: DefaultsFromNone, Param: ""}),
				WithAllVariadic(false),
				WithWithIsset(false),
				WithConstructorTypeRender(ConstructorPublicRender),
				WithShowWarnings(true),
			)
		}
	})
}

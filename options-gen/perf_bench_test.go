//nolint:testpackage
package optionsgen

import (
	"testing"
)

var benchOptionsSink Options

// BenchmarkNewOptions benchmarks creating Options with various configurations.
func BenchmarkNewOptions(b *testing.B) {
	b.Run("minimal_options", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()

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
		b.ResetTimer()

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

// BenchmarkOptionsValidation benchmarks Options.Validate() method.
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
	b.ResetTimer()

	for b.Loop() {
		_ = opts.Validate()
	}
}

// BenchmarkWithOptionals benchmarks chaining optional parameters.
func BenchmarkWithOptionals(b *testing.B) {
	b.Run("10_options", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()

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

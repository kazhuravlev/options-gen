//nolint:testpackage
package generator

import (
	"fmt"
	"testing"
)

var (
	benchStageTemplateSink []byte
	benchStageFormatSink   []byte
)

// BenchmarkRenderStages splits Render into its two stages for the same specs as
// BenchmarkRenderCriticalPath: template_execute is context building plus
// text/template execution (renderTemplate), format_source is
// formatGeneratedSource (parse, prune imports, render the import block, gofmt)
// on that template output.
func BenchmarkRenderStages(b *testing.B) {
	for _, optionCount := range []int{10, 50, 100} {
		opts := benchmarkRenderOptions(benchmarkOptionSpec(optionCount))

		src, err := renderTemplate(opts)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(fmt.Sprintf("template_execute/%d_fields", optionCount), func(b *testing.B) {
			b.ReportAllocs()

			var err error
			for b.Loop() {
				benchStageTemplateSink, err = renderTemplate(opts)
				if err != nil {
					b.Fatal(err)
				}
			}
			b.SetBytes(int64(len(benchStageTemplateSink)))
		})

		b.Run(fmt.Sprintf("format_source/%d_fields", optionCount), func(b *testing.B) {
			b.ReportAllocs()

			var err error
			for b.Loop() {
				benchStageFormatSink, err = formatGeneratedSource(src)
				if err != nil {
					b.Fatal(err)
				}
			}
			b.SetBytes(int64(len(src)))
		})
	}
}

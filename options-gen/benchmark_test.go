package optionsgen

import (
	"path/filepath"
	"testing"
)

func BenchmarkRunCriticalPath(b *testing.B) {
	benchmarks := []struct {
		name        string
		caseDir     string
		defaults    Defaults
		allVariadic bool
		withIsset   bool
	}{
		{
			name:    "builtin_fields",
			caseDir: filepath.Join("testdata", "case-02-builtin-types"),
		},
		{
			name:        "all_variadic",
			caseDir:     filepath.Join("testdata", "case-02.1-builtin-types-all-variadic"),
			allVariadic: true,
		},
		{
			name:    "generics",
			caseDir: filepath.Join("testdata", "case-05-generics-01"),
		},
		{
			name:    "defaults_from_tag",
			caseDir: filepath.Join("testdata", "case-12-defaults-tag-02"),
			defaults: Defaults{
				From: DefaultsFromTag,
			},
		},
		{
			name:      "with_isset",
			caseDir:   filepath.Join("testdata", "case-20-isset"),
			withIsset: true,
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

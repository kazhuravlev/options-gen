//nolint:testpackage
package generator

import (
	"go/token"
	"os"
	"path/filepath"
	"testing"

	"github.com/kazhuravlev/options-gen/internal/ctype"
)

var benchStructFileSink string

// BenchmarkFindStructTypeParamsAndFields2Cold measures the packages.Load
// fallback used for structs declared outside the current module (a `go list`
// subprocess plus type-checking on every iteration). The cold PackageStore
// path is covered by BenchmarkPackageStoreLoadCold in utils_test.go.
func BenchmarkFindStructTypeParamsAndFields2Cold(b *testing.B) {
	if testing.Short() {
		b.Skip("runs `go list` on every iteration")
	}

	dir := b.TempDir()
	pkgFile := filepath.Join(dir, "external.go")
	src := `package external

import "time"

type Local struct{}

type Options struct {
	Timeout  time.Duration
	Optional Local
	Values   []Local
}
`
	if err := os.WriteFile(pkgFile, []byte(src), ctype.DefaultPermission); err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()

	for b.Loop() {
		file, _, fields, err := findStructTypeParamsAndFields2(token.NewFileSet(), pkgFile, "Options", dir, "external")
		if err != nil {
			b.Fatal(err)
		}

		if len(fields) != 3 {
			b.Fatalf("unexpected fields count: %d", len(fields))
		}

		benchStructFileSink = file.Name.Name
	}
}

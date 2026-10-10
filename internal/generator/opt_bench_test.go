//nolint:testpackage
package generator

import (
	"fmt"
	"go/token"
	"os"
	"path/filepath"
	"testing"

	"github.com/kazhuravlev/options-gen/internal/ctype"
)

var (
	benchCrowdedSpecSink *GetOptionSpecRes
	benchStructFileSink  string
)

// BenchmarkGetOptionSpecCrowdedDir measures GetOptionSpec when the options
// file sits in a directory with many sibling files: in a real package the
// directory holds the previously generated options_generated.go, tests and
// other sources that have nothing to do with the target struct.
func BenchmarkGetOptionSpecCrowdedDir(b *testing.B) {
	const siblingFiles = 20

	caseDir := filepath.Join("..", "..", "options-gen", "testdata", "case-02-builtin-types")
	optionsSrc, err := os.ReadFile(filepath.Join(caseDir, "options.go"))
	if err != nil {
		b.Fatal(err)
	}

	siblingSrc, err := os.ReadFile(filepath.Join(caseDir, "options_generated.go"))
	if err != nil {
		b.Fatal(err)
	}

	dir := b.TempDir()
	filePath := filepath.Join(dir, "options.go")
	if err := os.WriteFile(filePath, optionsSrc, ctype.DefaultPermission); err != nil {
		b.Fatal(err)
	}

	for i := range siblingFiles {
		name := filepath.Join(dir, fmt.Sprintf("sibling_%02d.go", i))
		if err := os.WriteFile(name, siblingSrc, ctype.DefaultPermission); err != nil {
			b.Fatal(err)
		}
	}

	b.ReportAllocs()

	for b.Loop() {
		benchCrowdedSpecSink, err = GetOptionSpec(filePath, "Options", "default", false, nil)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkPackageStoreLoadCold measures a cold packages.Load through a fresh
// PackageStore on every iteration. validator/v10 is a direct dependency with
// a sizeable transitive dependency graph, so the cost of loading (or not
// loading) dependencies is visible.
func BenchmarkPackageStoreLoadCold(b *testing.B) {
	if testing.Short() {
		b.Skip("runs `go list` on every iteration")
	}

	const pkgPath = "github.com/go-playground/validator/v10"

	b.ReportAllocs()

	for b.Loop() {
		store := NewPackageStore(token.NewFileSet(), ".")
		pkg, err := store.Load(pkgPath)
		if err != nil {
			b.Fatal(err)
		}

		if pkg.Types == nil || pkg.Types.Scope().Lookup("Validate") == nil {
			b.Fatal("package types are not loaded")
		}
	}
}

// BenchmarkFindStructTypeParamsAndFields2Cold measures the packages.Load
// fallback used for structs declared outside the current module.
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

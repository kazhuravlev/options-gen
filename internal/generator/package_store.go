package generator

import (
	"fmt"
	"go/token"

	"golang.org/x/tools/go/packages"
)

type PackageStore struct {
	fset    *token.FileSet
	dirPath string
	pkgs    map[string]*packages.Package
}

func NewPackageStore(fset *token.FileSet, dirPath string) *PackageStore {
	return &PackageStore{
		fset:    fset,
		dirPath: dirPath,
		pkgs:    make(map[string]*packages.Package),
	}
}

// Load returns a cached package or loads it by full import path.
func (s *PackageStore) Load(pkgName string) (*packages.Package, error) {
	if pkg, ok := s.pkgs[pkgName]; ok {
		return pkg, nil
	}

	cfg := &packages.Config{ //nolint:exhaustruct
		Mode: packages.NeedTypes | packages.NeedDeps,
		Dir:  s.dirPath,
		Fset: s.fset,
	}

	pkgs, err := packages.Load(cfg, pkgName)
	if err != nil {
		return nil, fmt.Errorf("loading package: %w", err)
	}

	if len(pkgs) == 0 {
		return nil, fmt.Errorf("no packages found")
	}

	s.pkgs[pkgName] = pkgs[0]

	return pkgs[0], nil
}

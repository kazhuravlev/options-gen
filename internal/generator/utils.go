package generator

import (
	"fmt"
	"go/ast"

	"github.com/pkg/errors"
)

func findStructFields(packages map[string]*ast.Package, typeName string) []*ast.Field {
	var methods []*ast.Field

	decls := getDecls(packages)
	for _, decl := range decls {
		x, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}

		for _, spec := range x.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec) //nolint:varnamelen
			if !ok {
				continue
			}

			if typeSpec.Name.Name != typeName {
				continue
			}

			structType, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				continue
			}

			methods = append(methods, structType.Fields.List...)
		}
	}

	return methods
}

func getDecls(packages map[string]*ast.Package) []ast.Decl {
	var res []ast.Decl

	for _, pkg := range packages {
		for _, fileObj := range pkg.Files {
			res = append(res, fileObj.Decls...)
		}
	}

	return res
}

func makeTypeName(expr ast.Expr) (string, error) {
	switch concreteType := expr.(type) {
	case *ast.SelectorExpr:
		ident, ok := concreteType.X.(*ast.Ident)
		if !ok {
			return "", errors.New("cast to *ast.Ident")
		}

		return ident.Name + "." + concreteType.Sel.Name, nil
	case *ast.Ident:
		return concreteType.Name, nil
	case *ast.ArrayType:
		eltName, err := makeTypeName(concreteType.Elt)
		if err != nil {
			return "", err
		}

		return "[]" + eltName, nil
	case *ast.StarExpr:
		tName, err := makeTypeName(concreteType.X)
		if err != nil {
			return "", errors.Wrap(err, "cannot make type name for star expr")
		}

		return "*" + tName, nil
	case *ast.MapType:
		tName := fmt.Sprintf("map[%s]%s", concreteType.Key, concreteType.Value)

		return tName, nil
	default:
		return "", errors.Errorf("unknown field type (%T)", expr)
	}
}

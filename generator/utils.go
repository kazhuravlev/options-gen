package generator

import (
	"fmt"
	"github.com/pkg/errors"
	"go/ast"
)

//nolint:gocognit,nestif
func findStructFields(packages map[string]*ast.Package, typeName string) []*ast.Field {
	var methods []*ast.Field

	for _, pkg := range packages {
		for _, fileObj := range pkg.Files {
			for _, decl := range fileObj.Decls {
				if x, ok := decl.(*ast.GenDecl); ok {
					for _, spec := range x.Specs {
						if typ, ok := spec.(*ast.TypeSpec); ok {
							if xType, ok := typ.Type.(*ast.StructType); ok {
								if typ.Name.Name == typeName {
									methods = append(
										methods,
										xType.Fields.List...,
									)
								}
							}
						}
					}
				}
			}
		}
	}

	return methods
}

func makeTypeName(expr ast.Expr) (string, error) {
	var typeName string
	switch t := expr.(type) {
	case *ast.SelectorExpr:
		typeName = t.X.(*ast.Ident).Name + "." + t.Sel.Name
	case *ast.Ident:
		typeName = t.Name
	case *ast.ArrayType:
		eltName, err := makeTypeName(t.Elt)
		if err != nil {
			return "", err
		}

		typeName = "[]" + eltName
	case *ast.StarExpr:
		tName, err := makeTypeName(t.X)
		if err != nil {
			return "", errors.Wrap(err, "cannot make type name for star expr")
		}

		return "*" + tName, nil
	case *ast.MapType:
		tName := fmt.Sprintf("map[%s]%s", t.Key, t.Value)

		return tName, nil
	default:
		return "", errors.Errorf("unknown field type (%T)", expr)
	}

	return typeName, nil
}

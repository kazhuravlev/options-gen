package generator

import (
	"fmt"
	"go/ast"

	"github.com/pkg/errors"
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

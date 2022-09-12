package generator

import (
	"fmt"
	"go/ast"
	"go/types"
)

func findStructTypeParamsAndFields(packages map[string]*ast.Package, typeName string) ([]*ast.Field, []*ast.Field) {
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

			return extractFields(typeSpec.TypeParams), extractFields(structType.Fields)
		}
	}
	return nil, nil
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

func extractFields(fl *ast.FieldList) []*ast.Field {
	if fl == nil {
		return nil
	}
	return fl.List
}

func makeTypeName(expr ast.Expr) (string, error) {
	switch expr.(type) {
	case *ast.SelectorExpr, *ast.Ident, *ast.ArrayType, *ast.StarExpr, *ast.MapType, *ast.FuncType, *ast.ChanType:
	default:
		return "", fmt.Errorf("unsupported field type (%T)", expr)
	}

	return types.ExprString(expr), nil
}

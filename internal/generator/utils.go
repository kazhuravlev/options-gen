package generator

import (
	"go/ast"
	"unicode"
	"unicode/utf8"
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

func isPublic(fieldName string) bool {
	char, _ := utf8.DecodeRuneInString(fieldName)
	if char == utf8.RuneError {
		return false
	}

	if unicode.IsLetter(char) && unicode.IsUpper(char) {
		return true
	}

	return false
}

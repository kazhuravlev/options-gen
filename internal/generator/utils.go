package generator

import (
	"bytes"
	"fmt"
	"go/ast"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

func prefix(str1, str2 string) string {
	if str1 == "" || str2 == "" {
		return ""
	}

	buf := bytes.NewBuffer(nil)
	if len(str1) > len(str2) {
		str1, str2 = str2, str1
	}

	for i := range str1 {
		if str1[i] != str2[i] {
			break
		}

		buf.WriteByte(str1[i])
	}

	return buf.String()
}

func formatComment(comment string) string {
	if comment == "" {
		return ""
	}

	buf := bytes.NewBuffer(nil)

	lines := strings.Split(comment, "\n")
	// this is a hack for go1.18.
	//nolint:godox
	// TODO: need help with go build constraints. failed to configure tags like // go:build go1.18
	commonPrefix := lines[0]
	for _, line := range lines {
		commonPrefix = prefix(commonPrefix, line)
	}

	for i := range lines {
		lines[i] = strings.TrimPrefix(lines[i], commonPrefix)

		// Last line contains an empty string.
		if lines[i] == "" && i == len(lines)-1 {
			continue
		}

		if i != 0 {
			buf.WriteString("\n")
		}

		buf.WriteString("// ")
		buf.WriteString(lines[i])
	}

	return buf.String()
}

func findStructTypeParamsAndFields(packages map[string]*ast.Package, typeName string) ([]*ast.Field, []*ast.Field) {
	decls := getDecls(packages)
	for _, decl := range decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}

		for _, spec := range genDecl.Specs {
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

	return char != utf8.RuneError && unicode.IsUpper(char)
}

func checkDefaultValue(fieldType string, tag string) error {
	var err error

	switch fieldType {
	case "int", "int8", "int16", "int32", "int64":
		_, err = strconv.ParseInt(tag, 10, 64)

	case "uint", "uint8", "uint16", "uint32", "uint64":
		_, err = strconv.ParseUint(tag, 10, 64)

	case "float32", "float64":
		_, err = strconv.ParseFloat(tag, 64)

	case "time.Duration":
		_, err = time.ParseDuration(tag)

	case "string":
		// As is.

	default:
		return fmt.Errorf("unsupported type `%s`", fieldType)
	}

	if err != nil {
		return fmt.Errorf("bad default value %w %s", err, tag)
	}

	return nil
}

package generator

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"golang.org/x/mod/modfile"
)

const typePartsSize = 2

func formatComment(comment string) string {
	if comment == "" {
		return ""
	}

	buf := bytes.NewBuffer(nil)

	lines := strings.Split(comment, "\n")
	for i := range lines {
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

	case "bool":
		if tag != "true" && tag != "false" {
			return fmt.Errorf("bool type only supports true/false")
		}

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

func extractTypesFromPackage(
	fset *token.FileSet,
	packages map[string]*ast.Package,
	packageName string,
	currentDir string,
) (map[string]*ast.Package, bool) {
	for _, pkg := range packages {
		for _, file := range pkg.Files {
			currentPackage := file.Name.Name

			for _, importNode := range file.Imports {
				var name string
				if importNode.Name != nil {
					v, err := strconv.Unquote(importNode.Name.Name)
					if err != nil {
						return nil, false
					}

					name = v
				} else {
					pth, err := strconv.Unquote(importNode.Path.Value)
					if err != nil {
						return nil, false
					}

					name = path.Base(pth)
				}

				if name != packageName {
					continue
				}

				pth, err := strconv.Unquote(importNode.Path.Value)
				if err != nil {
					return nil, false
				}

				node, ok := tryParsePackage(fset, pth, currentPackage, currentDir)
				if !ok {
					continue
				}

				return node, true
			}
		}
	}

	return nil, false
}

func tryParsePackage(
	fset *token.FileSet,
	packagePath string,
	currentPackage string,
	currentDir string,
) (map[string]*ast.Package, bool) {
	if currentDir == "." {
		path, err := os.Getwd()
		if err != nil {
			return nil, false
		}

		currentDir = path
	}
	currentPackage = "/" + strings.TrimSuffix(currentPackage, "_test") + "/"
	if idx := strings.Index(packagePath, currentPackage); idx > -1 {
		node, err := parser.ParseDir(
			fset,
			path.Join(currentDir, packagePath[idx+len(currentPackage):]),
			nil,
			parser.ParseComments,
		)
		if err == nil {
			return node, true
		}
	}

	var gomodPath string
	for dir, file := currentDir, ""; dir != "" && file != "."; dir, file = path.Split(dir) {
		dir = strings.TrimSuffix(dir, "/")

		node, err := parser.ParseDir(fset, path.Join(dir, "vendor", packagePath), nil, parser.ParseComments)
		if err == nil {
			return node, true
		}

		if gomodPath == "" {
			testPath := path.Join(dir, "go.mod")
			if _, err := os.Stat(testPath); err == nil {
				gomodPath = testPath
			}
		}
	}

	packagePath, version, err := resolvePackageVersion(gomodPath, packagePath)
	if err != nil {
		return nil, false
	}

	if version != "" {
		version = "@" + version
	}

	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		// last chance to detect go path
		cmd := exec.Command("go")
		cmd.Args = append(cmd.Args, "env", "GOPATH")

		out, err := cmd.Output()
		if err != nil {
			return nil, false
		}

		gopath = strings.TrimSpace(string(out))
	}

	node, err := parser.ParseDir(
		fset,
		path.Join(gopath, "pkg/mod", packagePath+version),
		nil,
		parser.ParseComments,
	)
	if err == nil {
		return node, true
	}

	return nil, false
}

func resolvePackageVersion(
	gomodPath string,
	packagePath string,
) (string, string, error) {
	var version string
	data, err := os.ReadFile(gomodPath)
	if err != nil {
		return "", "", err
	}

	parsedModFile, err := modfile.Parse(gomodPath, data, nil)
	if err != nil {
		return "", "", err
	}

	for _, repl := range parsedModFile.Replace {
		if repl.Old.Path != packagePath {
			continue
		}

		return repl.New.Path, repl.New.Version, nil
	}

	if version == "" {
		for _, req := range parsedModFile.Require {
			if req.Mod.Path != packagePath {
				continue
			}

			return packagePath, req.Mod.Version, nil
		}
	}

	return packagePath, "", nil
}

func extractSliceKind(
	fset *token.FileSet,
	packages map[string]*ast.Package,
	typeName string,
	currentDir string,
) (string, bool) {
	if strings.HasPrefix(typeName, "[]") {
		return typeName[2:], true
	}

	decls := getDecls(packages)
	nameParts := strings.SplitN(typeName, ".", typePartsSize)
	typeName = nameParts[0]
	if len(nameParts) > 1 {
		typeName = nameParts[1]

		newPkg, ok := extractTypesFromPackage(fset, packages, nameParts[0], currentDir)
		if !ok {
			return "", false
		}

		decls = getDecls(newPkg)
	}

	for _, decl := range decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}

		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}

			if typeSpec.Name.Name != typeName {
				continue
			}

			if arr, ok := typeSpec.Type.(*ast.ArrayType); ok {
				return types.ExprString(arr.Elt), true
			}
		}
	}

	return "", false
}

func normalizeTypeName(typeName string) string {
	if idx := strings.LastIndex(typeName, "."); idx > -1 {
		typeName = typeName[idx+1:]
	}

	return strings.TrimPrefix(strings.TrimPrefix(typeName, "[]"), "*")
}

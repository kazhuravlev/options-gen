package generator

import (
	"bytes"
	"errors"
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"path"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"golang.org/x/tools/go/packages"
)

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

func findStructTypeParamsAndFields(packages map[string]*ast.Package, typeName string) (*ast.Package, *ast.File, []*ast.Field, []*ast.Field, bool) {
	for _, pkgObj := range packages {
		for _, fileObj := range pkgObj.Files {
			for _, decl := range fileObj.Decls {
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

					return pkgObj, fileObj, extractFields(typeSpec.TypeParams), extractFields(structType.Fields), true
				}
			}
		}
	}

	return nil, nil, nil, nil, false
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

func normalizeTypeName(typeName string) string {
	if idx := strings.LastIndex(typeName, "."); idx > -1 {
		typeName = typeName[idx+1:]
	}

	return strings.TrimPrefix(strings.TrimPrefix(typeName, "[]"), "*")
}

// extractSliceElemType will find the element type for given slice.
func extractSliceElemType(workDir string, fset *token.FileSet, curPkg *ast.Package, curFile *ast.File, expr ast.Expr) (string, error) {
	switch expr := expr.(type) {
	default:
		return "", errors.New("unsupported expression")
	case *ast.SelectorExpr:
		// Extract package name and type name
		pkgIdent, ok := expr.X.(*ast.Ident)
		if !ok {
			return "", errors.New("unsupported selector")
		}

		pkgName := pkgIdent.Name
		typeName := expr.Sel.Name

		// FIXME(zhuravlev): use only go/packages anf go/types packages in whole project.
		importPath, alias := findImportPath(curFile.Imports, pkgName)
		if importPath == "" {
			return "", errors.New("import path not found")
		}

		pkg, err := loadPkg(fset, importPath, workDir)
		if err != nil {
			return "", errors.New("unable to load package")
		}

		lookupType := pkg.Types.Scope().Lookup(typeName)
		switch expr := lookupType.(type) {
		case *types.TypeName:
			switch expr := expr.Type().(type) {
			case *types.Named:
				switch expr := expr.Underlying().(type) {
				case *types.Slice:
					// FIXME(zhuravlev): use more gently way to extract the type name.
					fmt.Printf("%T: %s\n", expr.Elem(), expr.String())
					switch expr := expr.Elem().(type) {
					case *types.Named:
						typName := alias + "." + expr.Obj().Name()

						return typName, nil
					case *types.Basic:
						return expr.Name(), nil
					}
				}
			}
		}

		return "", errors.New("lookup type not found")
	case *ast.ArrayType:
		return types.ExprString(expr.Elt), nil
	case *ast.Ident:
		if expr.Obj == nil {
			return "", errors.New("id is empty")
		}

		switch expr := expr.Obj.Decl.(type) {
		default:
			return "", errors.New("unsupported ident expression")
		case *ast.TypeSpec:
			return extractSliceElemType(workDir, fset, curPkg, curFile, expr.Type)
		}
	}
}

// findImportPath return full package name and alias if presented.
func findImportPath(imports []*ast.ImportSpec, pkgName string) (string, string) {
	for _, imp := range imports {
		importPath, err := strconv.Unquote(imp.Path.Value)
		if err != nil {
			continue
		}

		// If the import has an alias, check that
		if imp.Name != nil {
			aliasName, err := strconv.Unquote(imp.Name.Name)
			if err == nil && aliasName == pkgName {
				return importPath, aliasName
			}
		} else {
			// Otherwise, check if the base package name matches
			baseName := path.Base(importPath)
			if baseName == pkgName {
				return importPath, baseName
			}
		}
	}

	return "", ""
}

// loadPkg loads a package by full import path.
func loadPkg(fset *token.FileSet, pkgName, dirPath string) (*packages.Package, error) {
	cfg := &packages.Config{
		Mode: packages.NeedName |
			packages.NeedSyntax |
			packages.NeedTypes |
			packages.NeedImports |
			packages.NeedDeps |
			packages.NeedTypesInfo,
		Dir:  dirPath,
		Fset: fset,
	}

	pkgs, err := packages.Load(cfg, pkgName)
	if err != nil {
		return nil, fmt.Errorf("loading package: %w", err)
	}

	if len(pkgs) == 0 {
		return nil, fmt.Errorf("no packages found")
	}

	return pkgs[0], nil
}

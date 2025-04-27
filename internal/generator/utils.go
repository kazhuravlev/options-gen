package generator

import (
	"bytes"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"path"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"golang.org/x/tools/go/packages"
)

var errIsNotSlice = errors.New("it is not slice")

// Named Capture Group support since Go 1.22.
// When we remove support for Go versions below 1.22, we will be able to use code like
//
// var (
//	importPackageMask             = regexp.MustCompile(`(?<pkgName>[\w_\-\.\d]+)(\/v\d+)?$`)
//	importPackageMaskPkgNameIndex = importPackageMask.SubexpIndex("pkgName")
// )

var importPackageMask = regexp.MustCompile(`([\w_\-\.\d]+)(\/v\d+)?$`)

const importPackageMaskPkgNameIndex = 1

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

func findStructTypeParamsAndFields(fset *token.FileSet, filePath, typeName string) (*ast.File, []*ast.Field, []*ast.Field, error) { //nolint:lll
	// TODO(kazhuravlev): should be replaces to findStructTypeParamsAndFields2 but after fixing the performance issues.
	workDir := path.Dir(filePath)

	node, err := parser.ParseDir(fset, workDir, nil, parser.ParseComments)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("cannot parse file: %w", err)
	}

	for _, pkgObj := range node {
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

					return fileObj, extractFields(typeSpec.TypeParams), extractFields(structType.Fields), nil
				}
			}
		}
	}

	return nil, nil, nil, errors.New("cannot find target struct")
}

func findStructTypeParamsAndFields2(fset *token.FileSet, filePath, optStructName string) (*ast.File, []*ast.Field, []*ast.Field, error) { //nolint:lll
	workDir := path.Dir(filePath)

	// Configure the loader to use types package instead of ParseDir
	cfg := &packages.Config{
		Mode: packages.NeedSyntax |
			packages.NeedTypes |
			packages.NeedDeps,
		Dir:   workDir,
		Tests: true,
		Fset:  fset,
	}

	// Load the package that contains the file we want to parse
	pkgs, err := packages.Load(cfg, "file="+filePath)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("load package: %w", err)
	}

	if len(pkgs) == 0 {
		return nil, nil, nil, errors.New("no packages found")
	}

	pkg := pkgs[0]
	if len(pkg.Errors) > 0 {
		return nil, nil, nil, fmt.Errorf("package contains errors: %v", pkg.Errors)
	}

	// Find the struct
	var file *ast.File
	var typeParams []*ast.Field
	var fields []*ast.Field
	var found bool

SearchLoop:
	for _, f := range pkg.Syntax {
		for _, decl := range f.Decls {
			genDecl, ok := decl.(*ast.GenDecl)
			if !ok {
				continue
			}

			for _, spec := range genDecl.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}

				if typeSpec.Name.Name != optStructName {
					continue
				}

				structType, ok := typeSpec.Type.(*ast.StructType)
				if !ok {
					continue
				}

				file = f
				typeParams = extractFields(typeSpec.TypeParams)
				fields = extractFields(structType.Fields)
				found = true
				break SearchLoop
			}
		}
	}
	if !found {
		return nil, nil, nil, errors.New("cannot find target struct")
	}

	return file, typeParams, fields, nil
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
func extractSliceElemType(
	workDir string,
	fset *token.FileSet,
	curFile *ast.File,
	expr ast.Expr,
) (string, error) {
	switch expr := expr.(type) {
	default:
		return "", errIsNotSlice
	case *ast.SelectorExpr:
		// Extract package name and type name
		pkgIdent, ok := expr.X.(*ast.Ident)
		if !ok {
			return "", errors.New("unsupported selector")
		}

		pkgName := pkgIdent.Name
		typeName := expr.Sel.Name

		importPath, alias := findImportPath(curFile.Imports, pkgName)
		if importPath == "" {
			return "", errors.New("import path not found")
		}

		pkg, err := loadPkg(fset, importPath, workDir)
		if err != nil {
			return "", errors.New("unable to load package")
		}

		lookupType := pkg.Types.Scope().Lookup(typeName)
		if expr, ok := lookupType.(*types.TypeName); ok { //nolint:nestif
			if expr, ok := expr.Type().(*types.Named); ok {
				if expr, ok := expr.Underlying().(*types.Slice); ok {
					switch expr := expr.Elem().(type) {
					case *types.Named:
						if importPath == expr.Obj().Pkg().Path() {
							return alias + "." + expr.Obj().Name(), nil
						}

						return expr.Obj().Pkg().Name() + "." + expr.Obj().Name(), nil
					case *types.Basic:
						return expr.Name(), nil
					}
				}

				return "", errIsNotSlice
			}
		}

		return "", errors.New("lookup type not found")
	case *ast.ArrayType:
		return types.ExprString(expr.Elt), nil
	case *ast.Ident:
		if expr.Obj == nil {
			return "", errIsNotSlice
		}

		switch expr := expr.Obj.Decl.(type) {
		default:
			return "", errors.New("unsupported ident expression")
		case *ast.TypeSpec:
			return extractSliceElemType(workDir, fset, curFile, expr.Type)
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
			match := importPackageMask.FindStringSubmatch(importPath)
			if len(match) > 0 && match[importPackageMaskPkgNameIndex] == pkgName {
				return importPath, pkgName
			}
		}
	}

	return "", ""
}

// loadPkg loads a package by full import path.
func loadPkg(fset *token.FileSet, pkgName, dirPath string) (*packages.Package, error) {
	cfg := &packages.Config{ //nolint:exhaustruct
		Mode: packages.NeedTypes | packages.NeedDeps,
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

func parseTag(tag *ast.BasicLit, fieldName string, tagName string) (TagOption, []string) {
	var tagOpt TagOption
	if tag == nil {
		return tagOpt, nil
	}

	value := tag.Value
	tagOpt.GoValidator = reflect.StructTag(strings.Trim(value, "`")).Get("validate")
	tagOpt.Default = reflect.StructTag(strings.Trim(value, "`")).Get(tagName)

	var warnings []string
	optionTag := reflect.StructTag(strings.Trim(value, "`")).Get("option")
	for _, opt := range strings.Split(optionTag, ",") {
		optParts := strings.SplitN(opt, "=", keyValueSliceSize)
		var optName, optValue string
		optName = optParts[0]

		if len(optParts) > 1 {
			optValue = optParts[1]
		}

		switch optName {
		case "mandatory":
			tagOpt.IsRequired = true

		case "required":
			// NOTE: remove the tag.
			warnings = append(warnings, fmt.Sprintf(
				"Deprecated: use `option:\"mandatory\"` "+
					"instead for field `%s` to force the passing "+
					"option in the constructor argument\n", fieldName))

			tagOpt.IsRequired = true

		case "not-empty":
			// NOTE: remove the tag.
			warnings = append(warnings, fmt.Sprintf(
				"Deprecated: use "+
					"github.com/go-playground/validator `validate` tag to check "+
					"the field `%s` content\n", fieldName))

			if !strings.Contains(tagOpt.GoValidator, "required") {
				if tagOpt.GoValidator == "" {
					tagOpt.GoValidator = "required"
				} else {
					tagOpt.GoValidator += ",required"
				}
			}

		case "variadic":
			val, err := strconv.ParseBool(optValue)
			if err != nil {
				warnings = append(warnings, fmt.Sprintf("Error: parse variadic for the field %s failed: %s\n",
					fieldName, err.Error()))
			}

			tagOpt.Variadic = val
			tagOpt.VariadicIsSet = true

		case "-":
			tagOpt.Skip = true
		}
	}

	return tagOpt, warnings
}

func typeParamsStr(params []*ast.Field) (string, string, error) {
	if len(params) == 0 {
		return "", "", nil
	}

	paramNames := make([]string, 0, len(params))
	paramNamesWithTypes := make([]string, len(params))
	for i, param := range params {
		if len(param.Names) == 0 {
			return "", "", fmt.Errorf("unnamed param %s", param.Type)
		}

		names := make([]string, len(param.Names))
		for i := range param.Names {
			names[i] = param.Names[i].Name
		}

		paramNames = append(paramNames, names...)

		typeName := types.ExprString(param.Type)
		paramNamesWithTypes[i] = fmt.Sprintf("%s %s", strings.Join(names, ", "), typeName)
	}

	paramNamesStr := fmt.Sprintf("[%s]", strings.Join(paramNames, ", "))
	paramExprStr := fmt.Sprintf("[%s]", strings.Join(paramNamesWithTypes, ", "))

	return paramExprStr, paramNamesStr, nil
}

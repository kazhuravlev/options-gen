package generator

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"golang.org/x/tools/go/packages"
)

var errIsNotSlice = errors.New("it is not slice")

func formatComment(comment string) string {
	if comment == "" {
		return ""
	}

	if !strings.Contains(comment, "\n") {
		return "// " + comment
	}

	buf := make([]byte, 0, len(comment)+strings.Count(comment, "\n")*3)
	lineStart := 0
	lineIndex := 0

	for i := 0; i <= len(comment); i++ {
		if i < len(comment) && comment[i] != '\n' {
			continue
		}

		// Last line contains an empty string.
		if i == len(comment) && lineStart == i {
			break
		}

		if lineIndex != 0 {
			buf = append(buf, '\n')
		}

		buf = append(buf, '/', '/', ' ')
		buf = append(buf, comment[lineStart:i]...)

		lineIndex++
		lineStart = i + 1
	}

	return string(buf)
}

func findStructTypeParamsAndFields(fset *token.FileSet, filePath, typeName string) (*ast.File, []*ast.Field, []*ast.Field, error) { //nolint:lll
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
					typeSpec, ok := spec.(*ast.TypeSpec)
					if !ok {
						continue
					}

					if typeSpec.Name.Name != typeName {
						continue
					}

					switch castedType := typeSpec.Type.(type) {
					case *ast.StructType:
						return fileObj, extractFields(typeSpec.TypeParams), extractFields(castedType.Fields), nil
					case *ast.SelectorExpr:
						pkgIdent, ok := castedType.X.(*ast.Ident)
						if !ok {
							continue
						}

						importPath, _ := findImportPath(fileObj.Imports, pkgIdent.Name)
						if importPath == "" {
							continue
						}

						file, typeParams, fields, err := findStructTypeParamsAndFields2(
							fset,
							importPath,
							castedType.Sel.Name,
							workDir,
							pkgIdent.Name,
						)
						if err != nil {
							return nil, nil, nil, err
						}

						file.Imports = mergeImportSpecs(fileObj.Imports, file.Imports)

						return file, typeParams, fields, nil
					}
				}
			}
		}
	}

	return nil, nil, nil, errors.New("cannot find target struct")
}

func findStructTypeParamsAndFields2(
	fset *token.FileSet,
	importPath, optStructName, workDir, pkgName string,
) (*ast.File, []*ast.Field, []*ast.Field, error) {
	if workDir == "" {
		workDir = path.Dir(importPath)
	}

	if file, typeParams, fields, err := findLocalStructTypeParamsAndFields(
		fset,
		importPath,
		optStructName,
		workDir,
		pkgName,
	); err == nil {
		return file, typeParams, fields, nil
	}

	// Configure the loader to use types package instead of ParseDir
	cfg := &packages.Config{ //nolint:exhaustruct
		Mode: packages.NeedSyntax |
			packages.NeedTypes |
			packages.NeedDeps,
		Dir:   workDir,
		Tests: true,
		Fset:  fset,
	}

	if stat, err := os.Stat(importPath); err == nil && !stat.IsDir() {
		importPath = "file=" + importPath
	}

	// Load the package that contains the file we want to parse
	pkgs, err := packages.Load(cfg, importPath)
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

	scope := pkg.Types.Scope()

	for _, astFile := range pkg.Syntax {
		for _, decl := range astFile.Decls {
			genDecl, ok := decl.(*ast.GenDecl)
			if !ok {
				continue
			}

			for _, spec := range genDecl.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec) //nolint:varnamelen
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

				typeParams := extractFields(typeSpec.TypeParams)
				fields := extractFields(structType.Fields)
				var toDelFields []int

				for index, field := range fields {
					public := true
					for _, name := range field.Names {
						public = public && isPublic(name.Name)
					}

					if !public {
						toDelFields = append(toDelFields, index)

						continue
					}

					field.Type = addPackageToType(field.Type, pkgName, scope)
				}

				for i := len(toDelFields) - 1; i >= 0; i-- {
					idx := toDelFields[i]
					fields = deleteByIndex(fields, idx)
				}

				return astFile, typeParams, fields, nil
			}
		}
	}

	return nil, nil, nil, errors.New("cannot find target struct")
}

func findLocalStructTypeParamsAndFields(
	fset *token.FileSet,
	importPath, optStructName, workDir, pkgName string,
) (*ast.File, []*ast.Field, []*ast.Field, error) {
	moduleRoot, modulePath, err := findModule(workDir)
	if err != nil {
		return nil, nil, nil, err
	}

	relPath, ok := strings.CutPrefix(importPath, modulePath+"/")
	if !ok {
		if importPath != modulePath {
			return nil, nil, nil, errors.New("import is outside current module")
		}

		relPath = ""
	}

	dirPath := filepath.Join(moduleRoot, filepath.FromSlash(relPath))
	stat, err := os.Stat(dirPath)
	if err != nil {
		return nil, nil, nil, err
	}

	if !stat.IsDir() {
		return nil, nil, nil, errors.New("local import is not a directory")
	}

	pkgs, err := parser.ParseDir(fset, dirPath, nil, parser.ParseComments)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("cannot parse local package: %w", err)
	}

	for _, pkgObj := range pkgs {
		localTypes := packageTypeNames(pkgObj)
		for _, fileObj := range pkgObj.Files {
			for _, decl := range fileObj.Decls {
				genDecl, ok := decl.(*ast.GenDecl)
				if !ok {
					continue
				}

				for _, spec := range genDecl.Specs {
					typeSpec, isTypeSpec := spec.(*ast.TypeSpec)
					if !isTypeSpec || typeSpec.Name.Name != optStructName {
						continue
					}

					structType, ok := typeSpec.Type.(*ast.StructType)
					if !ok {
						continue
					}

					fields := filterImportedFields(extractFields(structType.Fields), pkgName, localTypes)

					return fileObj, extractFields(typeSpec.TypeParams), fields, nil
				}
			}
		}
	}

	return nil, nil, nil, errors.New("cannot find target struct")
}

func findModule(workDir string) (root, modulePath string, err error) {
	dir, err := filepath.Abs(workDir)
	if err != nil {
		return "", "", err
	}

	for {
		goModPath := filepath.Join(dir, "go.mod")
		data, err := os.ReadFile(goModPath)
		if err == nil {
			modulePath, err := parseModulePath(string(data))
			if err != nil {
				return "", "", err
			}

			return dir, modulePath, nil
		}

		if !errors.Is(err, os.ErrNotExist) {
			return "", "", err
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", "", errors.New("go.mod not found")
		}

		dir = parent
	}
}

func parseModulePath(goMod string) (string, error) {
	for _, line := range strings.Split(goMod, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}

		if modulePath, ok := strings.CutPrefix(line, "module "); ok {
			fields := strings.Fields(modulePath)
			if len(fields) == 0 {
				return "", errors.New("empty module path")
			}

			return fields[0], nil
		}
	}

	return "", errors.New("module path not found")
}

func packageTypeNames(pkgObj *ast.Package) map[string]struct{} { //nolint:staticcheck
	typeNames := make(map[string]struct{})
	for _, fileObj := range pkgObj.Files {
		for _, decl := range fileObj.Decls {
			genDecl, ok := decl.(*ast.GenDecl)
			if !ok || genDecl.Tok != token.TYPE {
				continue
			}

			for _, spec := range genDecl.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}

				typeNames[typeSpec.Name.Name] = struct{}{}
			}
		}
	}

	return typeNames
}

func filterImportedFields(fields []*ast.Field, pkgName string, localTypes map[string]struct{}) []*ast.Field {
	filtered := make([]*ast.Field, 0, len(fields))
	for _, field := range fields {
		public := true
		for _, name := range field.Names {
			public = public && isPublic(name.Name)
		}

		if !public {
			continue
		}

		field.Type = addPackageToLocalType(field.Type, pkgName, localTypes)
		filtered = append(filtered, field)
	}

	return filtered
}

func addPackageToLocalType(inExpr ast.Expr, pkgName string, localTypes map[string]struct{}) ast.Expr {
	switch casted := inExpr.(type) {
	case *ast.Ident:
		if _, ok := localTypes[casted.Name]; ok {
			inExpr = &ast.SelectorExpr{
				Sel: casted,
				X: &ast.Ident{
					NamePos: token.NoPos,
					Name:    pkgName,
					Obj:     nil,
				},
			}
		}
	case *ast.StarExpr:
		casted.X = addPackageToLocalType(casted.X, pkgName, localTypes)
	case *ast.MapType:
		casted.Key = addPackageToLocalType(casted.Key, pkgName, localTypes)
		casted.Value = addPackageToLocalType(casted.Value, pkgName, localTypes)
	case *ast.SliceExpr:
		casted.X = addPackageToLocalType(casted.X, pkgName, localTypes)
	case *ast.ArrayType:
		casted.Elt = addPackageToLocalType(casted.Elt, pkgName, localTypes)
	case *ast.ChanType:
		casted.Value = addPackageToLocalType(casted.Value, pkgName, localTypes)
	case *ast.FuncType:
		for i := range extractFields(casted.Params) {
			casted.Params.List[i].Type = addPackageToLocalType(casted.Params.List[i].Type, pkgName, localTypes)
		}

		for i := range extractFields(casted.Results) {
			casted.Results.List[i].Type = addPackageToLocalType(casted.Results.List[i].Type, pkgName, localTypes)
		}
	case *ast.InterfaceType:
		for i := range extractFields(casted.Methods) {
			casted.Methods.List[i].Type = addPackageToLocalType(casted.Methods.List[i].Type, pkgName, localTypes)
		}
	case *ast.StructType:
		for i := range extractFields(casted.Fields) {
			casted.Fields.List[i].Type = addPackageToLocalType(casted.Fields.List[i].Type, pkgName, localTypes)
		}
	}

	return inExpr
}

func mergeImportSpecs(imports ...[]*ast.ImportSpec) []*ast.ImportSpec {
	seen := make(map[string]struct{})
	res := make([]*ast.ImportSpec, 0)
	for _, importGroup := range imports {
		for _, imp := range importGroup {
			key := imp.Path.Value
			if imp.Name != nil {
				key = imp.Name.Name + " " + key
			}

			if _, ok := seen[key]; ok {
				continue
			}

			seen[key] = struct{}{}
			res = append(res, imp)
		}
	}

	return res
}

func addPackageToType(inExpr ast.Expr, pkgName string, scope *types.Scope) ast.Expr {
	switch casted := inExpr.(type) {
	case *ast.Ident:
		if scope.Lookup(casted.Name) != nil {
			inExpr = &ast.SelectorExpr{
				Sel: casted,
				X: &ast.Ident{
					Name:    pkgName,
					NamePos: token.NoPos,
					Obj:     nil,
				},
			}
		}
	case *ast.StarExpr:
		casted.X = addPackageToType(casted.X, pkgName, scope)
	case *ast.MapType:
		casted.Key = addPackageToType(casted.Key, pkgName, scope)
		casted.Value = addPackageToType(casted.Value, pkgName, scope)
	case *ast.SliceExpr:
		casted.X = addPackageToType(casted.X, pkgName, scope)
	case *ast.ArrayType:
		casted.Elt = addPackageToType(casted.Elt, pkgName, scope)
	case *ast.ChanType:
		casted.Value = addPackageToType(casted.Value, pkgName, scope)
	case *ast.FuncType:
		for i := range extractFields(casted.Params) {
			casted.Params.List[i].Type = addPackageToType(casted.Params.List[i].Type, pkgName, scope)
		}

		for i := range extractFields(casted.Results) {
			casted.Results.List[i].Type = addPackageToType(casted.Results.List[i].Type, pkgName, scope)
		}
	case *ast.InterfaceType:
		for i := range extractFields(casted.Methods) {
			casted.Methods.List[i].Type = addPackageToType(casted.Methods.List[i].Type, pkgName, scope)
		}
	case *ast.StructType:
		for i := range extractFields(casted.Fields) {
			casted.Fields.List[i].Type = addPackageToType(casted.Fields.List[i].Type, pkgName, scope)
		}
	}

	return inExpr
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
	curFile *ast.File,
	expr ast.Expr,
	packageStore *PackageStore,
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

		pkg, err := packageStore.Load(importPath)
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
		return renderExprString(expr.Elt), nil
	case *ast.Ident:
		if expr.Obj == nil {
			return "", errIsNotSlice
		}

		switch expr := expr.Obj.Decl.(type) {
		default:
			return "", errors.New("unsupported ident expression")
		case *ast.TypeSpec:
			return extractSliceElemType(curFile, expr.Type, packageStore)
		}
	}
}

func renderExprString(expr ast.Expr) string {
	switch casted := expr.(type) {
	case *ast.Ident:
		return casted.Name
	case *ast.SelectorExpr:
		if pkgIdent, ok := casted.X.(*ast.Ident); ok {
			return pkgIdent.Name + "." + casted.Sel.Name
		}
	case *ast.StarExpr:
		return "*" + renderExprString(casted.X)
	case *ast.ArrayType:
		return "[]" + renderExprString(casted.Elt)
	case *ast.MapType:
		return "map[" + renderExprString(casted.Key) + "]" + renderExprString(casted.Value)
	case *ast.ChanType:
		switch casted.Dir {
		case ast.SEND:
			return "chan<- " + renderExprString(casted.Value)
		case ast.RECV:
			return "<-chan " + renderExprString(casted.Value)
		default:
			return "chan " + renderExprString(casted.Value)
		}
	case *ast.InterfaceType:
		if casted.Methods == nil || len(casted.Methods.List) == 0 {
			return "interface{}"
		}
	}

	return types.ExprString(expr)
}

// findImportPath return full package name and alias if presented.
func findImportPath(imports []*ast.ImportSpec, pkgName string) (string, string) {
	for _, imp := range imports {
		importPath, err := strconv.Unquote(imp.Path.Value)
		if err != nil {
			continue
		}

		if imp.Name != nil {
			if imp.Name.Name == pkgName {
				return importPath, pkgName
			}

			continue
		}

		if importPathBase(importPath) == pkgName {
			return importPath, pkgName
		}
	}

	return "", ""
}

func importPathBase(importPath string) string {
	slashIdx := strings.LastIndexByte(importPath, '/')
	base := importPath
	if slashIdx >= 0 {
		base = importPath[slashIdx+1:]
	}

	// Module major version suffixes like /v2 should map to the preceding path element.
	if len(base) > 1 && base[0] == 'v' {
		isVersion := true
		for i := 1; i < len(base); i++ {
			if base[i] < '0' || base[i] > '9' {
				isVersion = false

				break
			}
		}

		if isVersion && slashIdx > 0 {
			prev := importPath[:slashIdx]
			prevSlashIdx := strings.LastIndexByte(prev, '/')
			if prevSlashIdx >= 0 {
				return prev[prevSlashIdx+1:]
			}

			return prev
		}
	}

	return base
}

func parseTag(tag *ast.BasicLit, fieldName string, tagName string) (TagOption, []string) {
	var tagOpt TagOption
	if tag == nil {
		return tagOpt, nil
	}

	tagValue := reflect.StructTag(strings.Trim(tag.Value, "`"))
	tagOpt.GoValidator = tagValue.Get("validate")
	tagOpt.Default = tagValue.Get(tagName)

	var warnings []string
	optionTag := tagValue.Get("option")
	for len(optionTag) > 0 {
		nextComma := strings.IndexByte(optionTag, ',')
		opt := optionTag
		if nextComma >= 0 {
			opt = optionTag[:nextComma]
			optionTag = optionTag[nextComma+1:]
		} else {
			optionTag = ""
		}

		optName, optValue, _ := strings.Cut(opt, "=")

		switch optName {
		case "name":
			tagOpt.Name = optValue
		case "mandatory":
			tagOpt.IsRequired = true

		case "required":
			// NOTE: remove the tag.
			warnings = append(warnings, deprecatedRequiredWarning(fieldName))

			tagOpt.IsRequired = true

		case "not-empty":
			// NOTE: remove the tag.
			warnings = append(warnings, deprecatedNotEmptyWarning(fieldName))

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
				warnings = append(warnings, parseVariadicWarning(fieldName, err))
			}

			tagOpt.Variadic = val
			tagOpt.VariadicIsSet = true

		case "-":
			tagOpt.Skip = true
		}
	}

	return tagOpt, warnings
}

func deprecatedRequiredWarning(fieldName string) string {
	return "Deprecated: use `option:\"mandatory\"` instead for field `" + fieldName +
		"` to force the passing option in the constructor argument\n"
}

func deprecatedNotEmptyWarning(fieldName string) string {
	return "Deprecated: use github.com/go-playground/validator `validate` tag to check the field `" +
		fieldName + "` content\n"
}

func parseVariadicWarning(fieldName string, err error) string {
	return "Error: parse variadic for the field " + fieldName + " failed: " + err.Error() + "\n"
}

func typeParamsStr(params []*ast.Field) (string, string, error) {
	if len(params) == 0 {
		return "", "", nil
	}

	var namesBuilder strings.Builder
	var specBuilder strings.Builder
	namesBuilder.WriteByte('[')
	specBuilder.WriteByte('[')

	firstName := true
	firstField := true

	for _, param := range params {
		if len(param.Names) == 0 {
			return "", "", fmt.Errorf("unnamed param %s", param.Type)
		}

		if !firstField {
			specBuilder.WriteString(", ")
		}

		for idx, name := range param.Names {
			if !firstName {
				namesBuilder.WriteString(", ")
			}
			if idx != 0 {
				specBuilder.WriteString(", ")
			}

			namesBuilder.WriteString(name.Name)
			specBuilder.WriteString(name.Name)

			firstName = false
		}

		specBuilder.WriteByte(' ')
		specBuilder.WriteString(types.ExprString(param.Type))
		firstField = false
	}

	namesBuilder.WriteByte(']')
	specBuilder.WriteByte(']')

	return specBuilder.String(), namesBuilder.String(), nil
}

func deleteByIndex[T any](input []T, index int) []T {
	if index < 0 || index >= len(input) {
		return input
	}

	return slices.Delete(input, index, index+1)
}

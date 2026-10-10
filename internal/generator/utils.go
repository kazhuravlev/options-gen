package generator

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"math"
	"os"
	"path"
	"path/filepath"
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

	var buf strings.Builder
	buf.Grow(len(comment) + (strings.Count(comment, "\n")+1)*3)
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
			buf.WriteByte('\n')
		}

		buf.WriteString("// ")
		buf.WriteString(comment[lineStart:i])

		lineIndex++
		lineStart = i + 1
	}

	return buf.String()
}

// structDecl is a struct type declaration together with the file that
// declares it (or, for an alias of an imported struct, the file of the
// imported package with the caller's imports merged in).
type structDecl struct {
	file       *ast.File
	typeParams []*ast.Field
	fields     []*ast.Field
}

func findStructTypeParamsAndFields(fset *token.FileSet, filePath, typeName string) (*ast.File, []*ast.Field, []*ast.Field, error) { //nolint:lll
	workDir := path.Dir(filePath)

	// The target struct is almost always declared in filePath itself (the
	// GOFILE of the go:generate directive), so look there before parsing the
	// whole directory, which also holds the previously generated code and the
	// tests of the package.
	fileObj, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("cannot parse file: %w", err)
	}

	decl, found, err := findStructInFile(fset, fileObj, typeName, workDir)
	if err != nil {
		return nil, nil, nil, err
	}

	if found {
		return decl.file, decl.typeParams, decl.fields, nil
	}

	node, err := parser.ParseDir(fset, workDir, nil, parser.ParseComments)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("cannot parse file: %w", err)
	}

	for _, pkgObj := range node {
		for _, fileObj := range pkgObj.Files {
			decl, found, err := findStructInFile(fset, fileObj, typeName, workDir)
			if err != nil {
				return nil, nil, nil, err
			}

			if found {
				return decl.file, decl.typeParams, decl.fields, nil
			}
		}
	}

	return nil, nil, nil, errors.New("cannot find target struct")
}

// findStructInFile looks for the type declaration typeName in fileObj. It
// reports found=false when the file does not declare such a struct.
func findStructInFile(
	fset *token.FileSet,
	fileObj *ast.File,
	typeName, workDir string,
) (structDecl, bool, error) {
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
				return structDecl{
					file:       fileObj,
					typeParams: extractFields(typeSpec.TypeParams),
					fields:     extractFields(castedType.Fields),
				}, true, nil
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
					return structDecl{file: nil, typeParams: nil, fields: nil}, false, err
				}

				file.Imports = mergeImportSpecs(fileObj.Imports, file.Imports)

				return structDecl{file: file, typeParams: typeParams, fields: fields}, true, nil
			}
		}
	}

	return structDecl{file: nil, typeParams: nil, fields: nil}, false, nil
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

			// The module path may be written as a quoted string.
			if unquoted, err := strconv.Unquote(fields[0]); err == nil {
				return unquoted, nil
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
	case *ast.IndexExpr:
		// Generic instantiation with a single type argument: Box[Local].
		casted.X = addPackageToLocalType(casted.X, pkgName, localTypes)
		casted.Index = addPackageToLocalType(casted.Index, pkgName, localTypes)
	case *ast.IndexListExpr:
		// Generic instantiation with several type arguments: Pair[Local, int].
		casted.X = addPackageToLocalType(casted.X, pkgName, localTypes)
		for i := range casted.Indices {
			casted.Indices[i] = addPackageToLocalType(casted.Indices[i], pkgName, localTypes)
		}
	case *ast.Ellipsis:
		casted.Elt = addPackageToLocalType(casted.Elt, pkgName, localTypes)
	case *ast.ParenExpr:
		casted.X = addPackageToLocalType(casted.X, pkgName, localTypes)
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
	case *ast.IndexExpr:
		// Generic instantiation with a single type argument: Box[Local].
		casted.X = addPackageToType(casted.X, pkgName, scope)
		casted.Index = addPackageToType(casted.Index, pkgName, scope)
	case *ast.IndexListExpr:
		// Generic instantiation with several type arguments: Pair[Local, int].
		casted.X = addPackageToType(casted.X, pkgName, scope)
		for i := range casted.Indices {
			casted.Indices[i] = addPackageToType(casted.Indices[i], pkgName, scope)
		}
	case *ast.Ellipsis:
		casted.Elt = addPackageToType(casted.Elt, pkgName, scope)
	case *ast.ParenExpr:
		casted.X = addPackageToType(casted.X, pkgName, scope)
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

// Bit sizes of the numeric field types that support defaults. The generated
// code assigns the default literal to the field directly, so a value that does
// not fit into the field type would not compile.
const (
	bitSize8  = 8
	bitSize16 = 16
	bitSize32 = 32
	bitSize64 = 64
)

func checkDefaultValue(fieldType string, tag string) error {
	var err error
	switch fieldType {
	case "int", "int64":
		err = checkIntDefault(tag, bitSize64, false)
	case "int8":
		err = checkIntDefault(tag, bitSize8, false)
	case "int16":
		err = checkIntDefault(tag, bitSize16, false)
	case "int32":
		err = checkIntDefault(tag, bitSize32, false)

	case "uint", "uint64":
		err = checkIntDefault(tag, bitSize64, true)
	case "uint8":
		err = checkIntDefault(tag, bitSize8, true)
	case "uint16":
		err = checkIntDefault(tag, bitSize16, true)
	case "uint32":
		err = checkIntDefault(tag, bitSize32, true)

	case "float32":
		err = checkFloatDefault(tag, bitSize32)
	case "float64":
		err = checkFloatDefault(tag, bitSize64)

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

// checkIntDefault makes sure that tag is a decimal literal that fits into an
// integer type of the given bit size. Leading zeros are rejected because Go
// reads such literals as octal: `010` would silently become 8 and `08` would
// not compile at all.
func checkIntDefault(tag string, bitSize int, unsigned bool) error {
	var err error
	if unsigned {
		_, err = strconv.ParseUint(tag, 10, bitSize)
	} else {
		_, err = strconv.ParseInt(tag, 10, bitSize)
	}

	if err != nil {
		return err
	}

	if digits := strings.TrimLeft(tag, "+-"); len(digits) > 1 && digits[0] == '0' {
		return errors.New("leading zeros are not allowed")
	}

	return nil
}

// checkFloatDefault makes sure that tag is a float literal that fits into a
// float type of the given bit size. strconv accepts "NaN" and "Inf", but they
// are not valid Go literals, so they are rejected here.
func checkFloatDefault(tag string, bitSize int) error {
	val, err := strconv.ParseFloat(tag, bitSize)
	if err != nil {
		return err
	}

	if math.IsNaN(val) || math.IsInf(val, 0) {
		return errors.New("NaN and Inf are not valid float literals")
	}

	return nil
}

func normalizeTypeName(typeName string) string {
	typeName = trimTypeArgs(typeName)

	if idx := strings.LastIndex(typeName, "."); idx > -1 {
		typeName = typeName[idx+1:]
	}

	return strings.TrimPrefix(strings.TrimPrefix(typeName, "[]"), "*")
}

// trimTypeArgs strips a trailing type argument list, so that the embedded field
// name of an instantiated generic type is derived from the type name only:
// "pkg.Box[map[string]int]" -> "pkg.Box".
func trimTypeArgs(typeName string) string {
	if !strings.HasSuffix(typeName, "]") {
		return typeName
	}

	depth := 0
	for i := len(typeName) - 1; i > 0; i-- {
		switch typeName[i] {
		case ']':
			depth++
		case '[':
			depth--
			if depth == 0 {
				return typeName[:i]
			}
		}
	}

	// The brackets open at the very beginning ("[]"), so this is not a type argument list.
	return typeName
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
		pkgIdent, isIdent := expr.X.(*ast.Ident)
		if !isIdent {
			return "", errors.New("unsupported selector")
		}

		importPath, _ := findImportPath(curFile.Imports, pkgIdent.Name)
		if importPath == "" {
			return "", errors.New("import path not found")
		}

		pkg, err := packageStore.Load(importPath)
		if err != nil {
			return "", errors.New("unable to load package")
		}

		typeName, isTypeName := pkg.Types.Scope().Lookup(expr.Sel.Name).(*types.TypeName)
		if !isTypeName {
			return "", errors.New("lookup type not found")
		}

		// Aliases (`type Ints = []int`) are resolved to the type they stand for.
		sliceType, isSlice := types.Unalias(typeName.Type()).Underlying().(*types.Slice)
		if !isSlice {
			return "", errIsNotSlice
		}

		// Packages are referenced through the alias used by the current file,
		// if any, and through their own name otherwise.
		qualifier := func(elemPkg *types.Package) string {
			for _, imp := range curFile.Imports {
				if imp.Name == nil {
					continue
				}

				if impPath, err := strconv.Unquote(imp.Path.Value); err == nil && impPath == elemPkg.Path() {
					return imp.Name.Name
				}
			}

			return elemPkg.Name()
		}

		return types.TypeString(sliceType.Elem(), qualifier), nil
	case *ast.ArrayType:
		// [N]T is an array, not a slice: it cannot be appended to.
		if expr.Len != nil {
			return "", errIsNotSlice
		}

		return renderExprString(expr.Elt), nil
	case *ast.Ident:
		if expr.Obj == nil {
			return "", errIsNotSlice
		}

		switch decl := expr.Obj.Decl.(type) {
		default:
			return "", errors.New("unsupported ident expression")
		case *ast.Field:
			// Type parameters are declared as fields of the type parameter list.
			return "", errIsNotSlice
		case *ast.TypeSpec:
			return extractSliceElemType(curFile, decl.Type, packageStore)
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
			base = prev
			if prevSlashIdx := strings.LastIndexByte(prev, '/'); prevSlashIdx >= 0 {
				base = prev[prevSlashIdx+1:]
			}
		}
	}

	// Follow the same heuristic as goimports: a "go-" prefix is dropped and the
	// name ends at the first character that cannot be part of an identifier
	// (gopkg.in/yaml.v3 -> yaml, github.com/mattn/go-isatty -> isatty).
	base = strings.TrimPrefix(base, "go-")
	if idx := strings.IndexFunc(base, isNotIdentifierChar); idx >= 0 {
		base = base[:idx]
	}

	return base
}

func isNotIdentifierChar(char rune) bool {
	return char != '_' && !unicode.IsLetter(char) && !unicode.IsDigit(char)
}

func parseTag(tag *ast.BasicLit, fieldName string, tagName string) (TagOption, []string) {
	var tagOpt TagOption
	if tag == nil {
		return tagOpt, nil
	}

	// Struct tags are usually raw string literals, but interpreted string
	// literals ("option:\"mandatory\"") are legal as well.
	tagValue, err := strconv.Unquote(tag.Value)
	if err != nil {
		tagValue = strings.Trim(tag.Value, "`")
	}

	goValidator, defaultValue, optionTag := lookupTagValues(tagValue, tagName)
	tagOpt.GoValidator = goValidator
	tagOpt.Default = defaultValue

	var warnings []string
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

// lookupTagValues scans the struct tag once and returns the values of the
// `validate`, tagName and `option` keys. It follows the same parsing rules as
// reflect.StructTag.Lookup but avoids re-scanning the tag for every key.
func lookupTagValues(tag, tagName string) (validate, defaultValue, option string) { //nolint:cyclop
	var haveValidate, haveDefault, haveOption bool

	for tag != "" {
		// Skip leading space.
		i := 0
		for i < len(tag) && tag[i] == ' ' {
			i++
		}

		tag = tag[i:]
		if tag == "" {
			break
		}

		// Scan to colon. A space, a quote or a control character is a syntax error.
		i = 0
		for i < len(tag) && tag[i] > ' ' && tag[i] != ':' && tag[i] != '"' && tag[i] != 0x7f {
			i++
		}

		if i == 0 || i+1 >= len(tag) || tag[i] != ':' || tag[i+1] != '"' {
			break
		}

		name := tag[:i]
		tag = tag[i+1:]

		// Scan quoted string to find value.
		i = 1
		for i < len(tag) && tag[i] != '"' {
			if tag[i] == '\\' {
				i++
			}
			i++
		}

		if i >= len(tag) {
			break
		}

		qvalue := tag[:i+1]
		tag = tag[i+1:]

		if name != "validate" && name != tagName && name != "option" {
			continue
		}

		value, err := strconv.Unquote(qvalue)

		// The first occurrence of a key wins, malformed values resolve to an
		// empty string — exactly like reflect.StructTag.Lookup.
		if name == "validate" && !haveValidate {
			haveValidate = true
			if err == nil {
				validate = value
			}
		}

		if name == tagName && !haveDefault {
			haveDefault = true
			if err == nil {
				defaultValue = value
			}
		}

		if name == "option" && !haveOption {
			haveOption = true
			if err == nil {
				option = value
			}
		}
	}

	return validate, defaultValue, option
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

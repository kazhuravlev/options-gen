package generator

import (
	"bytes"
	"embed"
	"fmt"
	"github.com/pkg/errors"
	"go/ast"
	"go/parser"
	"go/token"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"golang.org/x/tools/imports"
	"io/ioutil"
	"path"
	"reflect"
	"strings"
	"text/template"
)

//go:embed templates/options.go.tpl
var templates embed.FS

type OptionMeta struct {
	Name      string
	Field     string
	Type      string
	TagOption TagOption
}

type TagOption struct {
	IsRequired bool
	IsNotEmpty bool
}

func RenderOptions(packageName, optionsStructName string, fileImports []string, data []OptionMeta) ([]byte, error) {
	tmpl := template.Must(template.ParseFS(templates, "templates/options.go.tpl"))

	tplContext := map[string]interface{}{
		"packageName":       packageName,
		"imports":           fileImports,
		"optionsStructName": optionsStructName,
		"options":           data,
	}
	buf := new(bytes.Buffer)
	if err := tmpl.Execute(buf, tplContext); err != nil {
		return nil, errors.Wrap(err, "cannot render template")
	}

	// reformat, remove unused and duplicate imports, sort them
	formatted, err := imports.Process("", buf.Bytes(), nil)
	if err != nil {
		return nil, errors.Wrap(err, "cannot optimize imports")
	}

	return formatted, nil
}

//nolint:gocognit,nestif
func findInterfaceMethods(packages map[string]*ast.Package, typeName string) []*ast.Field {
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

func GetOptionSpec(filePath, optionsStructName string) ([]OptionMeta, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseDir(fset, path.Dir(filePath), nil, parser.ParseComments)
	if err != nil {
		return nil, errors.Wrap(err, "cannot parse dir")
	}

	methods := findInterfaceMethods(node, optionsStructName)

	options := make([]OptionMeta, len(methods))
	for i := range methods {
		field := methods[i]
		typeName, err := makeTypeName(field.Type)
		if err != nil {
			return nil, errors.Wrap(err, "cannot make type name")
		}

		var tagOpt TagOption
		if field.Tag != nil {
			value := field.Tag.Value

			tag := reflect.StructTag(strings.Trim(value, "`")).Get("option")

			for _, opt := range strings.Split(tag, ",") {
				if opt == "required" {
					tagOpt.IsRequired = true
				}
				if opt == "not-empty" {
					tagOpt.IsNotEmpty = true
				}
			}
		}

		title := cases.Title(language.English, cases.NoLower)
		options[i] = OptionMeta{
			Name:      title.String(field.Names[0].Name),
			Field:     field.Names[0].Name,
			Type:      typeName,
			TagOption: tagOpt,
		}
	}

	return options, nil
}

func GetFileImports(filePath string) ([]string, error) {
	source, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read file %q", filePath)
	}

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filePath, source, 0)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot parse file %q", filePath)
	}

	fileImports := make([]string, 0, len(file.Imports))
	for _, importSpec := range file.Imports {
		imp := string(source[importSpec.Pos()-1 : importSpec.End()-1])
		fileImports = append(fileImports, imp)
	}

	return fileImports, nil
}

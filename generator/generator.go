package generator

import (
	"bytes"
	"fmt"
	"github.com/pkg/errors"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"path"
	"reflect"
	"strings"
	"text/template"
)

//go:generate go-assets-builder --package=generator --variable=templates --output=templates_generated.go templates/options.go.tpl

var (
	tplOption = mustLoadAsset("/templates/options.go.tpl")
)

type tagOption struct {
	IsRequired bool
	IsNotEmpty bool
}

type optionMeta struct {
	Name      string
	Field     string
	Type      string
	TagOption tagOption
}

func RenderOptions(packageName string, data []optionMeta) (string, error) {
	tmpl := template.Must(template.New("tpl").Parse(tplOption))

	tplContext := map[string]interface{}{
		"packageName": packageName,
		"options":     data,
	}
	buf := new(bytes.Buffer)
	if err := tmpl.Execute(buf, tplContext); err != nil {
		return "", errors.Wrap(err, "cannot RenderOptions template")
	}

	formatted, err := formatSource(buf.String())
	if err != nil {
		return "", errors.Wrap(err, "cannot format source")
	}

	return formatted, nil
}

func formatSource(s string) (string, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, "<inmem-file>", s, 0)
	if err != nil {
		return "", errors.Wrap(err, "cannot parse expresion")
	}

	buf2 := new(bytes.Buffer)
	if err := format.Node(buf2, fset, node); err != nil {
		log.Fatal(err)
	}

	return buf2.String(), nil
}

func mustLoadAsset(path string) string {
	file, ok := templates.Files[path]
	if !ok {
		panic(fmt.Sprintf("file %q not found", path))
	}
	blob, err := ioutil.ReadAll(file)
	if err != nil {
		panic(err)
	}
	return string(blob)
}

func findInterfaceMethods(packages map[string]*ast.Package, typeName string) ([]*ast.Field, error) {
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

	return methods, nil
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
	default:
		return "", errors.Errorf("unknown field type (%T). use only local-defined interfaces", expr)
	}

	return typeName, nil
}

func GetOptionSpec(filePath, optionsStructName string) ([]optionMeta, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseDir(fset, path.Dir(filePath), nil, parser.ParseComments)
	if err != nil {
		return nil, errors.Wrap(err, "cannot parse dir")
	}

	data, err := findInterfaceMethods(node, optionsStructName)
	if err != nil {
		return nil, errors.Wrap(err, "can")
	}

	options := make([]optionMeta, len(data))
	for i := range data {
		field := data[i]
		typeName, err := makeTypeName(field.Type)
		if err != nil {
			return nil, errors.Wrap(err, "cannot make type name")
		}

		var tagOpt tagOption
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

		options[i] = optionMeta{
			Name:      strings.Title(field.Names[0].Name),
			Field:     field.Names[0].Name,
			Type:      typeName,
			TagOption: tagOpt,
		}
	}

	return options, nil
}

package generator

import (
	"bytes"
	"embed"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"path"
	"reflect"
	"strings"
	"text/template"

	"github.com/pkg/errors"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"golang.org/x/tools/imports"
)

//go:embed templates/options.go.tpl
var templates embed.FS

var templateImports = []string{
	`"fmt"`,
	`"golang.org/x/sync/errgroup"`,
	`"github.com/kazhuravlev/options-gen/pkg/validator"`,
	`goplvalidator "github.com/go-playground/validator/v10"`,
}

type OptionSpec struct {
	Options []OptionMeta
}

func (s OptionSpec) HasValidation() bool {
	for _, o := range s.Options {
		if o.TagOption.GoValidator != "" {
			return true
		}
	}
	return false
}

type OptionMeta struct {
	Name      string
	Field     string
	Type      string
	TagOption TagOption
}

type TagOption struct {
	IsRequired  bool
	IsNotEmpty  bool
	GoValidator string
}

// RenderOptions will render file and out it's content.
func RenderOptions(packageName, optionsStructName string, fileImports []string, spec *OptionSpec) ([]byte, error) {
	tmpl := template.Must(template.ParseFS(templates, "templates/options.go.tpl"))

	tplContext := map[string]interface{}{
		"packageName":       packageName,
		"imports":           filterImports(fileImports, templateImports),
		"optionsStructName": optionsStructName,
		"options":           spec.Options,
		"hasValidation":     spec.HasValidation(),
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

func filterImports(srcImports []string, existentImports []string) []string {
	result := make([]string, 0, len(srcImports))
L:
	for _, i := range srcImports {
		for _, ii := range existentImports { // Small O(n^2).
			if i == ii {
				continue L
			}
			result = append(result, i)
		}
	}
	return result
}

// GetOptionSpec read the input filename by filePath, find optionsStructName
// and scan for options.
func GetOptionSpec(filePath, optionsStructName string) (*OptionSpec, error) {
	fset := token.NewFileSet()

	node, err := parser.ParseDir(fset, path.Dir(filePath), nil, parser.ParseComments)
	if err != nil {
		return nil, errors.Wrap(err, "cannot parse dir")
	}

	fields := findStructFields(node, optionsStructName)
	options := make([]OptionMeta, len(fields))

	for idx := range fields {
		field := fields[idx]

		fieldName := field.Names[0].Name

		typeName, err := makeTypeName(field.Type)
		if err != nil {
			return nil, errors.Wrap(err, "cannot make type name")
		}

		tagOpt := parseTag(field.Tag, fieldName)

		title := cases.Title(language.English, cases.NoLower)
		options[idx] = OptionMeta{
			Name:      title.String(fieldName),
			Field:     fieldName,
			Type:      typeName,
			TagOption: tagOpt,
		}
	}

	return &OptionSpec{Options: options}, nil
}

func parseTag(tag *ast.BasicLit, fieldName string) TagOption {
	tagOpt := TagOption{
		IsRequired:  false,
		IsNotEmpty:  false,
		GoValidator: "",
	}

	if tag == nil {
		return tagOpt
	}

	value := tag.Value
	tagOpt.GoValidator = reflect.StructTag(strings.Trim(value, "`")).Get("validate")

	optionTag := reflect.StructTag(strings.Trim(value, "`")).Get("option")
	for _, opt := range strings.Split(optionTag, ",") {
		if opt == "required" {
			log.Printf(
				"Deprecated: use `option:\"mandatory\"` "+
					"instead for field `%s` to force the passing "+
					"option in the constructor argument\n", fieldName)

			tagOpt.IsRequired = true
		}

		if opt == "mandatory" {
			tagOpt.IsRequired = true
		}

		// NOTE: remove the tag
		if opt == "not-empty" {
			log.Printf(
				"Deprecated: use "+
					"github.com/go-playground/validator tag to check "+
					"the field `%s` content\n", fieldName)

			tagOpt.IsNotEmpty = true
		}
	}

	return tagOpt
}

// GetFileImports read the file and parse the imports section. Return all found
// imports with aliases.
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

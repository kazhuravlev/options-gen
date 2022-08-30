package generator

import (
	"bytes"
	"embed"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path"
	"reflect"
	"strings"
	"text/template"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"golang.org/x/tools/imports"
)

//go:embed templates/options.go.tpl
var templates embed.FS

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
		"imports":           fileImports,
		"optionsStructName": optionsStructName,
		"options":           spec.Options,
		"hasValidation":     spec.HasValidation(),
	}
	buf := new(bytes.Buffer)

	if err := tmpl.Execute(buf, tplContext); err != nil {
		return nil, fmt.Errorf("cannot render template: %w", err)
	}

	// reformat, remove unused and duplicate imports, sort them
	formatted, err := imports.Process("", buf.Bytes(), nil)
	if err != nil {
		return nil, fmt.Errorf("cannot optimize imports: %w", err)
	}

	return formatted, nil
}

// GetOptionSpec read the input filename by filePath, find optionsStructName
// and scan for options.
func GetOptionSpec(filePath, optionsStructName string) (*OptionSpec, error) {
	fset := token.NewFileSet()

	node, err := parser.ParseDir(fset, path.Dir(filePath), nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("cannot parse dir: %w", err)
	}

	fields := findStructFields(node, optionsStructName)
	options := make([]OptionMeta, len(fields))

	for idx := range fields {
		field := fields[idx]

		fieldName := field.Names[0].Name

		typeName, err := makeTypeName(field.Type)
		if err != nil {
			return nil, fmt.Errorf("cannot make type name: %w", err)
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
		switch opt {
		case "required":
			log.Printf(
				"Deprecated: use `option:\"mandatory\"` "+
					"instead for field `%s` to force the passing "+
					"option in the constructor argument\n", fieldName)

			tagOpt.IsRequired = true

		case "mandatory":
			tagOpt.IsRequired = true

		case "not-empty":
			// NOTE: remove the tag
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
	source, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file (%q): %w", filePath, err)
	}

	fset := token.NewFileSet()

	file, err := parser.ParseFile(fset, filePath, source, 0)
	if err != nil {
		return nil, fmt.Errorf("cannot parse file (%q): %w", filePath, err)
	}

	fileImports := make([]string, 0, len(file.Imports))

	for _, importSpec := range file.Imports {
		imp := string(source[importSpec.Pos()-1 : importSpec.End()-1])
		fileImports = append(fileImports, imp)
	}

	return fileImports, nil
}

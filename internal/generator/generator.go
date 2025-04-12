package generator

import (
	"bytes"
	"embed"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"path"
	"reflect"
	"strconv"
	"strings"
	"text/template"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"golang.org/x/tools/imports"
)

//go:embed templates/options.go.tpl
var templates embed.FS

var tmpl = template.Must(template.ParseFS(templates, "templates/options.go.tpl"))

const keyValueSliceSize = 2

type OptionSpec struct {
	TypeParamsSpec string // [KeyT int | string, TT any]
	TypeParams     string // [KeyT, TT]
	Options        []OptionMeta
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
	Docstring string // contains a comment with `//`. Can be empty or contain a multi-line string.
	Field     string
	Type      string
	TagOption TagOption
}

type TagOption struct {
	IsRequired    bool
	GoValidator   string
	Default       string
	Variadic      bool
	VariadicIsSet bool
	Skip          bool
}

// RenderOptions will render file and out it's content.
func RenderOptions(
	packageName, optionsStructName string,
	fileImports []string,
	spec *OptionSpec,
	tagName, varName, funcName, prefix string,
	withIsset bool,
	constructorTypeRender string,
) ([]byte, error) {
	optionsStructType := optionsStructName
	optionsStructInstanceType := optionsStructName

	if spec.TypeParamsSpec != "" {
		optionsStructType += spec.TypeParamsSpec
		optionsStructInstanceType += spec.TypeParams
	}

	tplContext := map[string]interface{}{
		"packageName":   packageName,
		"imports":       fileImports,
		"options":       spec.Options,
		"optionsLen":    len(spec.Options),
		"hasValidation": spec.HasValidation(),

		"optionsTypeParamsSpec": spec.TypeParamsSpec,
		"optionsTypeParams":     spec.TypeParams,

		"optionsPrefix":             prefix,
		"optionsStructName":         optionsStructName,
		"optionsStructType":         optionsStructType,
		"optionsStructInstanceType": optionsStructInstanceType,
		"defaultsTagName":           tagName,
		"defaultsVarName":           varName,
		"defaultsFuncName":          funcName,

		"withIsset": withIsset,

		"constructorTypeRender": constructorTypeRender,
	}
	buf := new(bytes.Buffer)

	if err := tmpl.Execute(buf, tplContext); err != nil {
		return nil, fmt.Errorf("cannot render template: %w", err)
	}

	// reformat, remove unused and duplicate imports, sort them
	formatted, err := imports.Process("", buf.Bytes(), nil)
	if err != nil {
		_, _ = os.Stdout.Write(buf.Bytes()) // For issues debug.

		return nil, fmt.Errorf("cannot optimize imports: %w", err)
	}

	return formatted, nil
}

// GetOptionSpec read the input filename by filePath, find optionsStructName
// and scan for options.
func GetOptionSpec(filePath, optionsStructName, tagName string, allVariadic bool) (*OptionSpec, []string, error) {
	fset := token.NewFileSet()

	node, err := parser.ParseDir(fset, path.Dir(filePath), nil, parser.ParseComments)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot parse dir: %w", err)
	}

	typeParams, fields := findStructTypeParamsAndFields(node, optionsStructName)
	options := make([]OptionMeta, 0, len(fields))

	var warnings []string
	for idx := range fields {
		field := fields[idx]

		var fieldName string
		if len(field.Names) > 0 {
			fieldName = field.Names[0].Name
		} else {
			fieldName = normalizeTypeName(types.ExprString(field.Type))
		}

		tagOption, tagWarnings := parseTag(field.Tag, fieldName, tagName)
		if tagOption.Skip {
			continue
		}

		if isPublic(fieldName) {
			warnings = append(warnings,
				fmt.Sprintf(
					"Warning: consider to make `%s` is private. This is "+
						"will not allow to users to avoid constructor "+
						"method.", fieldName),
			)
		}

		warnings = append(warnings, tagWarnings...)
		optMeta := OptionMeta{
			Name:      cases.Title(language.English, cases.NoLower).String(fieldName),
			Docstring: formatComment(field.Doc.Text()),
			Field:     fieldName,
			Type:      types.ExprString(field.Type),
			TagOption: tagOption,
		}

		if optMeta.TagOption.Default != "" {
			if optMeta.TagOption.IsRequired {
				return nil, nil,
					fmt.Errorf("field `%s`: mandatory option cannot have a default value", optMeta.Field)
			}

			if err := checkDefaultValue(optMeta.Type, optMeta.TagOption.Default); err != nil {
				return nil, nil, fmt.Errorf("field `%s`: invalid `%s` tag value: %w", tagName, optMeta.Field, err)
			}
		}

		if optMeta.TagOption.Variadic || allVariadic {
			if !isSlice(optMeta.Type) {
				return nil, nil, fmt.Errorf("field `%s`: this type could not be variadic", tagName)
			}

			if !optMeta.TagOption.VariadicIsSet {
				optMeta.TagOption.Variadic = allVariadic
			}

			if optMeta.TagOption.Variadic {
				optMeta.Type = optMeta.Type[2:]
			}
		}

		options = append(options, optMeta)
	}

	tpSpec, tpString, err := typeParamsStr(typeParams)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to extract type params %w", err)
	}

	return &OptionSpec{
		TypeParamsSpec: tpSpec,
		TypeParams:     tpString,
		Options:        options,
	}, warnings, nil
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

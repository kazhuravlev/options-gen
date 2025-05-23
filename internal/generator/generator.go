package generator

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
	"go/token"
	"go/types"
	"os"
	"path"
	"regexp"
	"syscall"
	"text/template"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"golang.org/x/tools/imports"
)

//go:embed templates/options.go.tpl
var templates embed.FS

var tmpl = template.Must(template.ParseFS(templates, "templates/options.go.tpl"))

const keyValueSliceSize = 2

// Render will render file and out it's content.
func Render(opts Options) ([]byte, error) {
	if err := opts.Validate(); err != nil {
		return nil, fmt.Errorf("bad configuration: %w", err)
	}

	optionsStructType := opts.optionsStructName
	optionsStructInstanceType := opts.optionsStructName

	if opts.spec.TypeParamsSpec != "" {
		optionsStructType += opts.spec.TypeParamsSpec
		optionsStructInstanceType += opts.spec.TypeParams
	}

	tplContext := map[string]interface{}{
		"version":       opts.version,
		"packageName":   opts.packageName,
		"imports":       opts.fileImports,
		"options":       opts.spec.Options,
		"optionsLen":    len(opts.spec.Options),
		"hasValidation": opts.spec.HasValidation(),

		"optionsTypeParamsSpec": opts.spec.TypeParamsSpec,
		"optionsTypeParams":     opts.spec.TypeParams,

		"optionsPrefix":             opts.prefix,
		"optionsStructName":         opts.optionsStructName,
		"optionsStructType":         optionsStructType,
		"optionsStructInstanceType": optionsStructInstanceType,
		"optionsTypeName":           opts.optionTypeName,
		"defaultsTagName":           opts.tagName,
		"defaultsVarName":           opts.varName,
		"defaultsFuncName":          opts.funcName,

		"withIsset": opts.withIsset,

		"constructorTypeRender": opts.constructorTypeRender,
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

type GetOptionSpecRes struct {
	Spec     OptionSpec
	Warnings []string
	Imports  []string
}

// GetOptionSpec read the input filename by filePath, find optionsStructName
// and scan for options.
func GetOptionSpec(
	filePath, optStructName, tagName string,
	allVariadic bool,
	excludes []*regexp.Regexp,
) (*GetOptionSpecRes, error) {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("source file not exist: %w", syscall.ENOENT)
	}

	workDir := path.Dir(filePath)
	fset := token.NewFileSet()

	file, typeParams, fields, err := findStructTypeParamsAndFields(fset, filePath, optStructName)
	if err != nil {
		return nil, fmt.Errorf("cannot find target struct: %w", err)
	}

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
				return nil, fmt.Errorf("field `%s`: mandatory option cannot have a default value", optMeta.Field)
			}

			if err := checkDefaultValue(optMeta.Type, optMeta.TagOption.Default); err != nil {
				return nil, fmt.Errorf("field `%s`: invalid `%s` tag value: %w", tagName, optMeta.Field, err)
			}
		}

		if optMeta.TagOption.Variadic || allVariadic { //nolint:nestif
			if optMeta.TagOption.IsRequired {
				if optMeta.TagOption.Variadic {
					return nil, fmt.Errorf("field `%s`: this field is mandatory and could not be variadic", fieldName)
				}

				options = append(options, optMeta)

				continue
			}

			elementType, err := extractSliceElemType(workDir, fset, file, field.Type)
			if err != nil {
				if errors.Is(err, errIsNotSlice) && !optMeta.TagOption.Variadic {
					options = append(options, optMeta)

					continue
				}

				return nil, fmt.Errorf("field `%s`: this type could not be variadic: %w", fieldName, err)
			}

			if !optMeta.TagOption.VariadicIsSet {
				optMeta.TagOption.Variadic = allVariadic
			}

			if optMeta.TagOption.Variadic {
				optMeta.Type = elementType
			}
		}

		options = append(options, optMeta)
	}

	tpSpec, tpString, err := typeParamsStr(typeParams)
	if err != nil {
		return nil, fmt.Errorf("unable to extract type params %w", err)
	}

	// Process imports
	importSlice := make([]string, len(file.Imports))
	for i, imp := range file.Imports {
		importSlice[i] = imp.Path.Value
	}

	return &GetOptionSpecRes{
		Spec: OptionSpec{
			TypeParamsSpec: tpSpec,
			TypeParams:     tpString,
			Options:        applyExcludes(options, excludes),
		},
		Warnings: warnings,
		Imports:  importSlice,
	}, nil
}

func applyExcludes(options []OptionMeta, excludes []*regexp.Regexp) []OptionMeta {
	for _, reg := range excludes {
		var toDel []int
		for index, field := range options {
			if reg.MatchString(field.Name) {
				toDel = append(toDel, index)
			}
		}

		for i := len(toDel) - 1; i >= 0; i-- {
			idx := toDel[i]
			options = deleteByIndex(options, idx)
		}
	}

	return options
}

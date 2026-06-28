package generator

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"path"
	"regexp"
	"strconv"
	"syscall"
	"text/template"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"golang.org/x/tools/imports"
)

//go:embed templates/options.go.tpl
var templates embed.FS

var tmpl = template.Must(template.ParseFS(templates, "templates/options.go.tpl"))

const generatedFormatTabWidth = 8

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

	options := makeTemplateOptions(opts.spec.Options)
	tplContext := map[string]interface{}{
		"version":       opts.version,
		"packageName":   opts.packageName,
		"imports":       opts.fileImports,
		"options":       options,
		"optionsLen":    len(options),
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

	formatted, err := optimizeGeneratedSource(buf.Bytes())
	if err != nil {
		_, _ = os.Stdout.Write(buf.Bytes()) // For issues debug.

		return nil, fmt.Errorf("cannot optimize generated source: %w", err)
	}

	return formatted, nil
}

func optimizeGeneratedSource(src []byte) ([]byte, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", src, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("parse generated source: %w", err)
	}

	pruneUnusedImports(file)
	ast.SortImports(fset, file)

	var buf bytes.Buffer
	if err := format.Node(&buf, fset, file); err != nil {
		return nil, fmt.Errorf("format generated source: %w", err)
	}

	formatted, err := imports.Process("", buf.Bytes(), &imports.Options{
		Fragment:   false,
		AllErrors:  false,
		Comments:   true,
		TabIndent:  true,
		TabWidth:   generatedFormatTabWidth,
		FormatOnly: true,
	})
	if err != nil {
		return nil, fmt.Errorf("sort generated imports: %w", err)
	}

	return formatted, nil
}

func pruneUnusedImports(file *ast.File) {
	usedSelectors := make(map[string]struct{})
	ast.Inspect(file, func(node ast.Node) bool {
		switch node := node.(type) {
		case *ast.ImportSpec:
			return false
		case *ast.SelectorExpr:
			if ident, ok := node.X.(*ast.Ident); ok {
				usedSelectors[ident.Name] = struct{}{}
			}
		}

		return true
	})

	importDecls := file.Decls[:0]
	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.IMPORT {
			importDecls = append(importDecls, decl)

			continue
		}

		importSpecs := genDecl.Specs[:0]
		for _, spec := range genDecl.Specs {
			imp := spec.(*ast.ImportSpec)
			importName := importSpecName(imp)
			if importName == "_" || importName == "." {
				importSpecs = append(importSpecs, spec)

				continue
			}

			if _, ok := usedSelectors[importName]; ok {
				importSpecs = append(importSpecs, spec)
			}
		}

		if len(importSpecs) == 0 {
			continue
		}

		genDecl.Specs = importSpecs
		importDecls = append(importDecls, genDecl)
	}

	file.Decls = importDecls
}

func importSpecName(imp *ast.ImportSpec) string {
	if imp.Name != nil {
		return imp.Name.Name
	}

	importPath, err := strconv.Unquote(imp.Path.Value)
	if err != nil {
		return ""
	}

	return importPathBase(importPath)
}

type templateOptionMeta struct {
	OptionMeta
	TargetName  string
	TargetField string
}

func makeTemplateOptions(options []OptionMeta) []templateOptionMeta {
	res := make([]templateOptionMeta, 0, len(options))
	for _, opt := range options {
		targetName := opt.Name
		targetField := opt.Field
		if opt.TagOption.Name != "" {
			targetName = opt.TagOption.Name
			targetField = opt.TagOption.Name
		}

		res = append(res, templateOptionMeta{
			OptionMeta:  opt,
			TargetName:  targetName,
			TargetField: targetField,
		})
	}

	return res
}

type GetOptionSpecRes struct {
	Spec     OptionSpec
	Warnings []string
	Imports  []Import
}

type Import struct {
	Path  string
	Alias *string
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
	packageStore := NewPackageStore(fset, workDir)

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

			elementType, err := extractSliceElemType(file, field.Type, packageStore)
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
	importSlice := make([]Import, len(file.Imports))
	for i, imp := range file.Imports {
		var alias *string
		if imp.Name != nil {
			alias = &imp.Name.Name
		}

		importSlice[i] = Import{
			Path:  imp.Path.Value,
			Alias: alias,
		}
	}

	return &GetOptionSpecRes{
		Spec: OptionSpec{
			TypeParamsSpec: tpSpec,
			TypeParams:     tpString,
			Options:        ApplyExcludes(options, excludes),
		},
		Warnings: warnings,
		Imports:  importSlice,
	}, nil
}

func ApplyExcludes(options []OptionMeta, excludes []*regexp.Regexp) []OptionMeta {
	if len(options) == 0 || len(excludes) == 0 {
		return options
	}

	filtered := make([]OptionMeta, 0, len(options))
	for _, field := range options {
		var excluded bool
		for _, reg := range excludes {
			if reg.MatchString(field.Name) {
				excluded = true

				break
			}
		}

		if !excluded {
			filtered = append(filtered, field)
		}
	}

	return filtered
}

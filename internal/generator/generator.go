package generator

import (
	"bytes"
	"cmp"
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
	"slices"
	"strconv"
	"strings"
	"syscall"
	"text/template"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

//go:embed templates/options.go.tpl
var templates embed.FS

var tmpl = template.Must(template.ParseFS(templates, "templates/options.go.tpl"))

// Rough size of the rendered source: a fixed part plus the constructor line,
// the With* function and the validator of every option. It only pre-sizes the
// render buffer, so a miss costs nothing but a reallocation.
const (
	renderedBaseSize   = 1024
	renderedOptionSize = 512
)

// Render will render file and out it's content.
func Render(opts Options) ([]byte, error) {
	src, err := renderTemplate(opts)
	if err != nil {
		return nil, err
	}

	formatted, err := formatGeneratedSource(src)
	if err != nil {
		_, _ = os.Stdout.Write(src) // For issues debug.

		return nil, fmt.Errorf("cannot optimize generated source: %w", err)
	}

	return formatted, nil
}

// renderTemplate executes the options template and returns the raw, not yet
// formatted source.
func renderTemplate(opts Options) ([]byte, error) {
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
	buf := bytes.NewBuffer(make([]byte, 0, renderedBaseSize+renderedOptionSize*len(options)))

	if err := tmpl.Execute(buf, tplContext); err != nil {
		return nil, fmt.Errorf("cannot render template: %w", err)
	}

	return buf.Bytes(), nil
}

// formatGeneratedSource turns the rendered template output into the final
// file: it drops the imports that the code does not reference, lays the rest
// out the way goimports does (standard library first, then the other packages,
// sorted by path within a group, duplicates removed) and gofmt-formats the
// source once.
//
// It produces the same bytes as the previous pipeline, optimizeGeneratedSource,
// which parsed and printed the source four times (go/format and
// golang.org/x/tools/imports each do it twice). That pipeline lives in
// legacy_format_test.go as the reference implementation for the parity tests.
// The single pass relies on the template rendering doc comments in column 1:
// go/printer reformats a doc comment only when it is unindented and abuts its
// declaration, and the multi-pass pipeline got that on its second pass.
func formatGeneratedSource(src []byte) ([]byte, error) {
	// Neither comments nor object resolution matter for finding import usages.
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", src, parser.SkipObjectResolution)
	if err != nil {
		return nil, fmt.Errorf("parse generated source: %w", err)
	}

	// Imports always precede other declarations, so the import declarations
	// form one contiguous span of the source that is replaced as a whole.
	// The template renders a single import declaration.
	importDecls := make([]*ast.GenDecl, 0, 1)
	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.IMPORT {
			break
		}

		importDecls = append(importDecls, genDecl)
	}

	if len(importDecls) > 0 {
		used := usedSelectorBases(file)

		// The used imports are collected in runs: an unused import ends the
		// current run. That mirrors the multi-pass pipeline, where a pruned
		// spec left a blank line behind and gofmt/goimports sort and group
		// imports separated by a blank line independently.
		var runs [][]*ast.ImportSpec
		var run []*ast.ImportSpec
		for _, decl := range importDecls {
			for _, spec := range decl.Specs {
				imp := spec.(*ast.ImportSpec)
				if isImportUsed(imp, used) {
					run = append(run, imp)
				} else if len(run) > 0 {
					runs = append(runs, run)
					run = nil
				}
			}
		}

		if len(run) > 0 {
			runs = append(runs, run)
		}

		tokFile := fset.File(file.Pos())
		start := tokFile.Offset(importDecls[0].Pos())
		end := tokFile.Offset(importDecls[len(importDecls)-1].End())
		block := renderImportBlock(runs)

		spliced := make([]byte, 0, len(src)-(end-start)+len(block))
		spliced = append(spliced, src[:start]...)
		spliced = append(spliced, block...)
		spliced = append(spliced, src[end:]...)
		src = spliced
	}

	formatted, err := format.Source(src)
	if err != nil {
		return nil, fmt.Errorf("format generated source: %w", err)
	}

	return formatted, nil
}

// Import groups as goimports numbers them when no local prefix is configured.
// The appengine group is a goimports legacy that is mirrored for byte parity.
const (
	importGroupStd = iota
	importGroupOther
	importGroupAppengine
)

func importGroup(importPath string) int {
	if strings.HasPrefix(importPath, "appengine") {
		return importGroupAppengine
	}

	firstElem, _, _ := strings.Cut(importPath, "/")
	if strings.Contains(firstElem, ".") {
		return importGroupOther
	}

	return importGroupStd
}

type importLine struct {
	group int
	path  string // Unquoted import path, "" if the literal is malformed.
	name  string // Explicit import name, "" for an implicit one.
	text  string // The spec as written inside the import block.
}

// renderImportBlock renders a parenthesized import declaration from runs of
// import specs. Every run is laid out on its own in goimports style: specs
// sorted by group, path and name, exact duplicates dropped and a blank line
// between groups; a blank line separates the runs. It returns "" when there
// is nothing to import.
func renderImportBlock(runs [][]*ast.ImportSpec) string {
	var buf strings.Builder
	for _, run := range runs {
		if buf.Len() == 0 {
			buf.WriteString("import (\n")
		} else {
			buf.WriteByte('\n')
		}

		lines := sortedImportLines(run)
		for i, line := range lines {
			if i > 0 && lines[i-1].group != line.group {
				buf.WriteByte('\n')
			}

			buf.WriteByte('\t')
			buf.WriteString(line.text)
			buf.WriteByte('\n')
		}
	}

	if buf.Len() == 0 {
		return ""
	}

	buf.WriteByte(')')

	return buf.String()
}

// sortedImportLines returns specs sorted by group, path and name, without
// exact duplicates.
func sortedImportLines(specs []*ast.ImportSpec) []importLine {
	lines := make([]importLine, 0, len(specs))
	for _, spec := range specs {
		importPath, _ := strconv.Unquote(spec.Path.Value)
		line := importLine{group: importGroup(importPath), path: importPath, name: "", text: spec.Path.Value}
		if spec.Name != nil {
			line.name = spec.Name.Name
			line.text = spec.Name.Name + " " + spec.Path.Value
		}

		lines = append(lines, line)
	}

	slices.SortFunc(lines, func(a, b importLine) int {
		return cmp.Or(
			cmp.Compare(a.group, b.group),
			strings.Compare(a.path, b.path),
			strings.Compare(a.name, b.name),
		)
	})

	return slices.CompactFunc(lines, func(a, b importLine) bool {
		return a.path == b.path && a.name == b.name
	})
}

// usedSelectorBases returns the identifiers used as the base of a selector
// expression (the `pkg` of `pkg.Name`) anywhere in file outside import specs.
func usedSelectorBases(file *ast.File) map[string]struct{} {
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

	return usedSelectors
}

// isImportUsed reports whether imp has to stay: dot and blank imports always
// do, the others only when their name is used as a selector base.
func isImportUsed(imp *ast.ImportSpec, usedSelectors map[string]struct{}) bool {
	importName := importSpecName(imp)
	if importName == "_" || importName == "." {
		return true
	}

	_, ok := usedSelectors[importName]

	return ok
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
	res := make([]templateOptionMeta, len(options))
	for i := range options {
		tplOpt := &res[i]
		tplOpt.OptionMeta = options[i]

		if name := tplOpt.TagOption.Name; name != "" {
			tplOpt.TargetName = name
			tplOpt.TargetField = name
		} else {
			tplOpt.TargetName = tplOpt.Name
			tplOpt.TargetField = tplOpt.Field
		}
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
	stat, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("source file not exist: %w", syscall.ENOENT)
	}

	if err == nil && stat.IsDir() {
		return nil, fmt.Errorf("source file `%s` is a directory", filePath)
	}

	workDir := path.Dir(filePath)
	fset := token.NewFileSet()

	file, typeParams, fields, err := findStructTypeParamsAndFields(fset, filePath, optStructName)
	if err != nil {
		return nil, fmt.Errorf("cannot find target struct: %w", err)
	}

	options := make([]OptionMeta, 0, len(fields))
	packageStore := NewPackageStore(fset, workDir)
	// A cases.Caser is not safe for concurrent use, but building one per
	// field is needless: this one is used sequentially within this call.
	titleCaser := cases.Title(language.English, cases.NoLower)

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
			Name:      titleCaser.String(fieldName),
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
				return nil, fmt.Errorf("field `%s`: invalid `%s` tag value: %w", optMeta.Field, tagName, err)
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

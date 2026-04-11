package optionsgen

import (
	"fmt"
	"os"
	"regexp"

	"github.com/kazhuravlev/options-gen/internal/ctype"
	"github.com/kazhuravlev/options-gen/internal/generator"
)

type DefaultsFrom string

const (
	DefaultsFromTag  DefaultsFrom = "tag"
	DefaultsFromNone DefaultsFrom = "none"
	DefaultsFromVar  DefaultsFrom = "var"
	DefaultsFromFunc DefaultsFrom = "func"
)

type Defaults struct {
	From DefaultsFrom `json:"from"`
	// Param is function name/variable name for func and var accordingly
	Param string `json:"param"`
}

type ConstructorTypeRender string

const (
	ConstructorPublicRender  ConstructorTypeRender = "public"
	ConstructorPrivateRender ConstructorTypeRender = "private"
	ConstructorNoRender      ConstructorTypeRender = "no"
)

var outOptionTypeNamePattern = regexp.MustCompile(`^[a-zA-Z]+$`)

const defaultTagName = "default"

func Run(opts Options) error {
	if err := opts.Validate(); err != nil {
		return fmt.Errorf("bad configuration: %w", err)
	}

	tagName, varName, funcName := resolveDefaults(opts.defaults, opts.structName)

	spec, err := generator.GetOptionSpec(
		opts.inFilename,
		opts.structName,
		tagName,
		opts.allVariadic,
		opts.exclude,
	)
	if err != nil {
		return fmt.Errorf("cannot get options spec: %w", err)
	}

	outOptionTypeName, err := resolveOutOptionTypeName(opts.structName, opts.outOptionTypeName)
	if err != nil {
		return err
	}

	res, err := generator.Render(generator.NewOptions(
		generator.WithVersion(opts.version),
		generator.WithPackageName(opts.packageName),
		generator.WithOptionsStructName(opts.structName),
		generator.WithFileImports(spec.Imports),
		generator.WithSpec(&spec.Spec),
		generator.WithTagName(tagName),
		generator.WithVarName(varName),
		generator.WithFuncName(funcName),
		generator.WithPrefix(opts.outPrefix),
		generator.WithWithIsset(opts.withIsset),
		generator.WithConstructorTypeRender(string(opts.constructorTypeRender)),
		generator.WithOptionTypeName(outOptionTypeName),
	))
	if err != nil {
		return fmt.Errorf("cannot renderOptions template: %w", err)
	}

	if err := os.WriteFile(opts.outFilename, res, ctype.DefaultPermission); err != nil {
		return fmt.Errorf("cannot write result: %w", err)
	}

	if opts.showWarnings {
		for _, warning := range spec.Warnings {
			opts.warningsHandler(warning)
		}
	}

	return nil
}

func resolveDefaults(defaults Defaults, structName string) (tagName, varName, funcName string) {
	switch defaults.From {
	case DefaultsFromTag:
		tagName = defaults.Param
		if tagName == "" {
			tagName = defaultTagName
		}
	case DefaultsFromVar:
		varName = defaults.Param
		if varName == "" {
			varName = fmt.Sprintf("default%s", structName)
		}
	case DefaultsFromFunc:
		funcName = defaults.Param
		if funcName == "" {
			funcName = fmt.Sprintf("getDefault%s", structName)
		}
	}

	return tagName, varName, funcName
}

func resolveOutOptionTypeName(structName, outOptionTypeName string) (string, error) {
	if outOptionTypeName == "" {
		return "Opt" + structName + "Setter", nil
	}

	if !outOptionTypeNamePattern.MatchString(outOptionTypeName) {
		return "", fmt.Errorf("outOptionTypeName must be a valid type name, contains only letters a-z or A-Z")
	}

	return outOptionTypeName, nil
}

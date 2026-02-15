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

func Run(opts Options) error {
	if err := opts.Validate(); err != nil {
		return fmt.Errorf("bad configuration: %w", err)
	}

	var tagName, varName, funcName string
	switch opts.defaults.From {
	case DefaultsFromNone:
	case DefaultsFromTag:
		tagName = opts.defaults.Param
		if tagName == "" {
			tagName = "default"
		}
	case DefaultsFromVar:
		varName = opts.defaults.Param
		if varName == "" {
			varName = fmt.Sprintf("default%s", opts.structName)
		}
	case DefaultsFromFunc:
		funcName = opts.defaults.Param
		if funcName == "" {
			funcName = fmt.Sprintf("getDefault%s", opts.structName)
		}
	}

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

	outOptionTypeName := opts.outOptionTypeName
	if outOptionTypeName == "" {
		outOptionTypeName = "Opt" + opts.structName + "Setter"
	} else if !outOptionTypeNamePattern.MatchString(outOptionTypeName) {
		return fmt.Errorf("outOptionTypeName must be a valid type name, contains only letters a-z or A-Z")
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

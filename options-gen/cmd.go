package optionsgen

import (
	"fmt"
	"log"
	"os"
	"regexp"

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

func (c ConstructorTypeRender) Valid() bool {
	switch c {
	case ConstructorPublicRender,
		ConstructorPrivateRender,
		ConstructorNoRender:
	default:
		return false
	}

	return true
}

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
	)
	if err != nil {
		return fmt.Errorf("cannot get options spec: %w", err)
	}

	if err := applyExcludes(spec, opts.exclude); err != nil {
		return fmt.Errorf("apply excludes: %w", err)
	}

	outOptionTypeName := opts.outOptionTypeName
	if outOptionTypeName == "" {
		outOptionTypeName = "Opt" + opts.structName + "Setter"
	} else {
		onlyLetters := regexp.MustCompile(`^[a-zA-Z]+$`)
		if !onlyLetters.MatchString(outOptionTypeName) {
			return fmt.Errorf("outOptionTypeName must be a valid type name, contains only letters a-z or A-Z")
		}
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

	const perm = 0o644
	if err := os.WriteFile(opts.outFilename, res, perm); err != nil {
		return fmt.Errorf("cannot write result: %w", err)
	}

	if opts.showWarnings {
		for _, warning := range spec.Warnings {
			log.Println(warning)
		}
	}

	return nil
}

func applyExcludes(specs *generator.GetOptionSpecRes, excludes []string) error {
	for _, pattern := range excludes {
		reg, err := regexp.Compile(pattern)
		if err != nil {
			return fmt.Errorf("compile pattern '%s': %w", pattern, err)
		}

		var toDel []int
		for index, field := range specs.Spec.Options {
			if reg.MatchString(field.Name) {
				toDel = append(toDel, index)
			}
		}

		for i := len(toDel) - 1; i >= 0; i-- {
			idx := toDel[i]
			specs.Spec.Options = generator.DeleteByIndex(specs.Spec.Options, idx)
		}
	}

	return nil
}

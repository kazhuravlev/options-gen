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

	optionSpec, warnings, imports, err := generator.GetOptionSpec(
		opts.inFilename,
		opts.structName,
		tagName,
		opts.allVariadic,
	)
	if err != nil {
		return fmt.Errorf("cannot get options spec: %w", err)
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

	res, err := generator.RenderOptions(
		opts.version,
		opts.packageName, opts.structName, imports,
		optionSpec,
		tagName, varName, funcName,
		opts.outPrefix,
		opts.withIsset,
		string(opts.constructorTypeRender),
		outOptionTypeName,
	)
	if err != nil {
		return fmt.Errorf("cannot renderOptions template: %w", err)
	}

	const perm = 0o644
	if err := os.WriteFile(opts.outFilename, res, perm); err != nil {
		return fmt.Errorf("cannot write result: %w", err)
	}

	if opts.showWarnings {
		for _, warning := range warnings {
			log.Println(warning)
		}
	}

	return nil
}

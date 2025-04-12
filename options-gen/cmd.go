package optionsgen

import (
	"fmt"
	"log"
	"os"
	"strings"

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

func Run(
	inFilename, outFilename, structName, packageName, outPrefix string,
	defaults Defaults,
	showWarnings bool,
	withIsset bool,
	allVariadic bool,
	constructorTypeRender ConstructorTypeRender,
) error {
	outPrefix = strings.TrimSpace(outPrefix)

	var tagName, varName, funcName string
	switch defaults.From {
	case DefaultsFromNone:
	case DefaultsFromTag:
		tagName = defaults.Param
		if tagName == "" {
			tagName = "default"
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

	optionSpec, warnings, err := generator.GetOptionSpec(inFilename, structName, tagName, allVariadic)
	if err != nil {
		return fmt.Errorf("cannot get options spec: %w", err)
	}

	imports, err := generator.GetFileImports(inFilename)
	if err != nil {
		return fmt.Errorf("cannot get imports: %w", err)
	}

	res, err := generator.RenderOptions(
		packageName, structName, imports,
		optionSpec,
		tagName, varName, funcName,
		outPrefix,
		withIsset,
		string(constructorTypeRender),
	)
	if err != nil {
		return fmt.Errorf("cannot renderOptions template: %w", err)
	}

	const perm = 0o644
	if err := os.WriteFile(outFilename, res, perm); err != nil {
		return fmt.Errorf("cannot write result: %w", err)
	}

	if showWarnings {
		for _, warning := range warnings {
			log.Println(warning)
		}
	}

	return nil
}

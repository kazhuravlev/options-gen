package optionsgen

import (
	"fmt"
	"log"
	"os"

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

func Run(inFilename, outFilename, structName, packageName string, defaults Defaults, showWarnings bool) error {
	// парсим исходный файл так, что бы получить не только структуру, но и токены, связанные с defaults.
	// то есть defaults это модификатор парсинга, который заставит парсер вытащить доп инфу

	var tagName, varName, funcName string
	switch defaults.From {
	case DefaultsFromTag:
		tagName = defaults.Param
	case DefaultsFromVar:
		varName = defaults.Param
	case DefaultsFromFunc:
		funcName = defaults.Param
	}

	optionSpec, warnings, err := generator.GetOptionSpec(inFilename, structName, tagName)
	if err != nil {
		return fmt.Errorf("cannot get options spec: %w", err)
	}

	imports, err := generator.GetFileImports(inFilename)
	if err != nil {
		return fmt.Errorf("cannot get imports: %w", err)
	}

	res, err := generator.RenderOptions(packageName, structName, imports, optionSpec, tagName, varName, funcName)
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

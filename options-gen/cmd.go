package optionsgen

import (
	"fmt"
	"log"
	"os"

	"github.com/kazhuravlev/options-gen/internal/generator"
)

func Run(inFilename, outFilename, structName, packageName string, showWarnings bool) error {
	data, warnings, err := generator.GetOptionSpec(inFilename, structName)
	if err != nil {
		return fmt.Errorf("cannot get options spec: %w", err)
	}

	imports, err := generator.GetFileImports(inFilename)
	if err != nil {
		return fmt.Errorf("cannot get imports: %w", err)
	}

	res, err := generator.RenderOptions(packageName, structName, imports, data)
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

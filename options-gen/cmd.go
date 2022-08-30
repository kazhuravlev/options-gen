package optionsgen

import (
	"fmt"
	"os"

	"github.com/kazhuravlev/options-gen/internal/generator"
)

func Run(inFilename, outFilename, structName, packageName string) error {
	data, err := generator.GetOptionSpec(inFilename, structName)
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

	return nil
}

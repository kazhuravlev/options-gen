package optionsgen

import (
	"io/ioutil"

	"github.com/kazhuravlev/options-gen/generator"
	"github.com/pkg/errors"
)

func Run(inFilename, outFilename, structName, packageName string) error {
	data, err := generator.GetOptionSpec(inFilename, structName)
	if err != nil {
		return errors.Wrap(err, "cannot get options spec")
	}

	imports, err := generator.GetFileImports(inFilename)
	if err != nil {
		return errors.Wrap(err, "cannot get imports")
	}

	res, err := generator.RenderOptions(packageName, structName, imports, data)
	if err != nil {
		return errors.Wrap(err, "cannot renderOptions template")
	}

	const perm = 0o644
	if err := ioutil.WriteFile(outFilename, res, perm); err != nil {
		return errors.Wrap(err, "cannot write result")
	}

	return nil
}

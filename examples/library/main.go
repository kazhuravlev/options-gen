package main

import (
	optionsgen "github.com/kazhuravlev/options-gen/options-gen"
)

func main() {
	for _, params := range []struct {
		outFname   string
		structName string
	}{
		{
			outFname:   "./example_out_options.go",
			structName: "Options",
		},
		{
			outFname:   "./example_out_config.go",
			structName: "Config",
		},
		{
			outFname:   "./example_out_params.go",
			structName: "Params",
		},
	} {
		if err := optionsgen.Run(
			optionsgen.NewOptions(
				optionsgen.WithVersion("qa-version"),
				optionsgen.WithInFilename("./example_in.go"),
				optionsgen.WithOutFilename(params.outFname),
				optionsgen.WithStructName(params.structName),
				optionsgen.WithPackageName("main"),
				optionsgen.WithOutPrefix("Some"),
				optionsgen.WithDefaults(optionsgen.Defaults{From: optionsgen.DefaultsFromTag, Param: ""}),
				optionsgen.WithShowWarnings(true),
				optionsgen.WithWithIsset(false),
				optionsgen.WithAllVariadic(false),
				optionsgen.WithConstructorTypeRender(optionsgen.ConstructorPublicRender),
				optionsgen.WithOutOptionTypeName(""),
			),
		); err != nil {
			panic(err)
		}
	}
}

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
			"./example_in.go",
			params.outFname,
			params.structName,
			"main",
			"Some",
			optionsgen.Defaults{From: optionsgen.DefaultsFromTag, Param: ""},
			true,
			false,
			true,
			true,
		); err != nil {
			panic(err)
		}
	}
}

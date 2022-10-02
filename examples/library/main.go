package main

import (
	"github.com/kazhuravlev/options-gen/options-gen"
)

func main() {
	if err := optionsgen.Run(
		"./example_in.go",
		"./example_out.go",
		"Options",
		"main",
		true,
	); err != nil {
		panic(err)
	}

	if err := optionsgen.Run(
		"./example_in.go",
		"./example_out_config.go",
		"Config",
		"main",
		true,
	); err != nil {
		panic(err)
	}
}

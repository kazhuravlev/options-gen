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
	); err != nil {
		panic(err)
	}

	if err := optionsgen.Run(
		"./example_in.go",
		"./example_out_config.go",
		"Config",
		"main",
	); err != nil {
		panic(err)
	}
}

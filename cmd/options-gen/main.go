package main

import (
	"flag"
	"fmt"

	optionsgen "github.com/kazhuravlev/options-gen/options-gen"
)

func main() {
	var inFilename string
	var outFilename string
	var optionsStructName string

	var outPackageName string

	flag.StringVar(&inFilename, "filename", "", "input filename")
	flag.StringVar(&outFilename, "out-filename", "", "output filename")
	flag.StringVar(&outPackageName, "pkg", "", "output package name")
	flag.StringVar(&optionsStructName, "from-struct", "", "struct that contains options")
	flag.Parse()

	if inFilename == "" || outFilename == "" || outPackageName == "" || optionsStructName == "" {
		flag.Usage()
		//nolint:forbidigo
		fmt.Println("all options are required")
		return
	}

	if err := optionsgen.Run(inFilename, outFilename, optionsStructName, outPackageName); err != nil {
		//nolint:forbidigo
		fmt.Println("cannot run options gen", err.Error())
		return
	}
}

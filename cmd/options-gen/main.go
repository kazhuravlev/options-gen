package main

import (
	"flag"
	"fmt"
	"os"

	optionsgen "github.com/kazhuravlev/options-gen/options-gen"
)

func main() {
	var (
		inFilename        string
		outFilename       string
		optionsStructName string
		outPackageName    string
	)

	envGoFile := os.Getenv("GOFILE")
	envGoPackage := os.Getenv("GOPACKAGE")

	flag.StringVar(&inFilename, "filename", envGoFile, "input filename")
	flag.StringVar(&outPackageName, "pkg", envGoPackage, "output package name")
	flag.StringVar(&outFilename, "out-filename", "", "output filename")
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

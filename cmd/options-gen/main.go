package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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

	defaultOutFilename := strings.Replace(filepath.Base(envGoFile), ".go", "_generated.go", 1)

	flag.StringVar(&inFilename, "filename", envGoFile, "input filename")
	flag.StringVar(&outPackageName, "pkg", envGoPackage, "output package name")
	flag.StringVar(&outFilename, "out-filename", defaultOutFilename, "output filename")
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

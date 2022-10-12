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
		defaultsFrom      string
		muteWarnings      bool
	)

	envGoFile := os.Getenv("GOFILE")
	envGoPackage := os.Getenv("GOPACKAGE")

	defaultOutFilename := strings.Replace(filepath.Base(envGoFile), ".go", "_generated.go", 1)

	flag.StringVar(&inFilename, "filename", envGoFile, "input filename")
	flag.StringVar(&outPackageName, "pkg", envGoPackage, "output package name")
	flag.StringVar(&outFilename, "out-filename", defaultOutFilename, "output filename")
	flag.StringVar(&optionsStructName, "from-struct", "", "struct that contains options")
	flag.StringVar(&defaultsFrom, "defaults-from", "tag", "where to get defaults for options. none, tag=TagName, func=FuncName, var=VarName")
	flag.BoolVar(&muteWarnings, "mute-warnings", false, "mute all warnings")
	flag.Parse()

	if isEmpty(inFilename, outFilename, outPackageName, optionsStructName, defaultsFrom) {
		flag.Usage()
		//nolint:forbidigo
		fmt.Println("missed required options")

		return
	}

	err := optionsgen.Run(
		inFilename,
		outFilename,
		optionsStructName,
		outPackageName,
		!muteWarnings,
	)
	if err != nil {
		//nolint:forbidigo
		fmt.Println("cannot run options gen", err.Error())

		return
	}
}

func isEmpty(values ...string) bool {
	for i := range values {
		if values[i] == "" {
			return false
		}
	}

	return true
}

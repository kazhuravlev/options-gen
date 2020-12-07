package main

import (
	"flag"
	"fmt"
	"github.com/kazhuravlev/options-gen/generator"
	"io/ioutil"
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
		fmt.Println("all options are required")
		return
	}

	data, err := generator.GetOptionSpec(inFilename, optionsStructName)
	if err != nil {
		fmt.Println("cannot get options spec:", err.Error())
		return
	}

	imports, err := generator.GetImports(inFilename)
	if err != nil {
		fmt.Println("cannot get imports:", err.Error())
		return
	}

	res, err := generator.RenderOptions(outPackageName, imports, data)
	if err != nil {
		fmt.Println("cannot renderOptions template:", err.Error())
		return
	}

	if err := ioutil.WriteFile(outFilename, []byte(res), 0644); err != nil {
		fmt.Println("cannot write result:", err.Error())
		return
	}
}

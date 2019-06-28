package main

import (
	"flag"
	"fmt"
	"github.com/kazhuravlev/options-gen"
	"io/ioutil"
)

func main() {
	var inFilename string
	var outFilename string

	var outPackageName string

	flag.StringVar(&inFilename, "filename", "", "input filename")
	flag.StringVar(&outFilename, "out-filename", "", "output filename")
	flag.StringVar(&outPackageName, "pkg", "", "output package name")
	flag.Parse()

	data, err := optionsgen.GetOptionSpec(inFilename)
	if err != nil {
		fmt.Println("cannot get options spec:", err.Error())
		return
	}

	res, err := optionsgen.RenderOptions(outPackageName, data)
	if err != nil {
		fmt.Println("cannot renderOptions template:", err.Error())
		return
	}

	if err := ioutil.WriteFile(outFilename, []byte(res), 0644); err != nil {
		fmt.Println("cannot write result:", err.Error())
		return
	}
}

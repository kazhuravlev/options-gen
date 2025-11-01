package optionsgen

import (
	"fmt"
	"regexp"
)

//go:generate toolset run options-gen -from-struct=Options -all-variadic=true -defaults-from=var
type Options struct {
	version               string `validate:"required"`
	inFilename            string `validate:"required"`
	outFilename           string `validate:"required"`
	structName            string `validate:"required"`
	packageName           string `validate:"required"`
	outPrefix             string
	defaults              Defaults `validate:"required"`
	showWarnings          bool
	withIsset             bool
	allVariadic           bool
	constructorTypeRender ConstructorTypeRender `validate:"required"`
	outOptionTypeName     string
	exclude               []*regexp.Regexp
	warningsHandler       func(string) `validate:"required"`
}

var defaultOptions = Options{
	version:     "",
	inFilename:  "",
	outFilename: "",
	structName:  "",
	packageName: "",
	outPrefix:   "",
	defaults: Defaults{
		From:  DefaultsFromNone,
		Param: "",
	},
	showWarnings:          false,
	withIsset:             false,
	allVariadic:           false,
	constructorTypeRender: "",
	outOptionTypeName:     "",
	exclude:               nil,
	warningsHandler: func(msg string) {
		fmt.Println(msg)
	},
}

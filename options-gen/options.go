package optionsgen

//go:generate toolset run options-gen -from-struct=Options -all-variadic=true
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
	exclude               []string
}

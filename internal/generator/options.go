package generator

//go:generate toolset run options-gen -from-struct=Options
type Options struct {
	version               string `validate:"required"`
	packageName           string `validate:"required"`
	optionsStructName     string `validate:"required"`
	fileImports           []string
	spec                  *OptionSpec `validate:"required"`
	tagName               string
	varName               string
	funcName              string
	prefix                string
	withIsset             bool
	constructorTypeRender string `validate:"required"`
	optionTypeName        string `validate:"required"`
}

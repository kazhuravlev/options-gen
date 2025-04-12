//nolint:mnd
package gogenerate

//go:generate options-gen -from-struct=Options1 -out-prefix=KKK -out-filename=options1_generated.go -defaults-from=var=defaultOptions1 -with-isset
type Options1 struct {
	// Options1.field0
	field0 int `validate:"min:3"`
	// Options1.field1
	field1 int `validate:"min:3"`
	// Options1.field2
	field2 int `validate:"min:3"`
	// Options1.field3
	field3 int `validate:"min:3"`
}

var defaultOptions1 = Options1{
	field0: 0,
	field1: 1,
	field2: 2,
	field3: 3,
}

//go:generate options-gen -from-struct=Options2 -out-prefix=NNN -out-filename=options2_generated.go -defaults-from=var=defaultOptions2 -with-isset
type Options2 struct {
	// Options2.field1
	field1 int `validate:"min:3"`
	// Options2.field2
	field2 int `validate:"min:3"`
	// Options2.field3
	field3 int `validate:"min:3"`
	// Options2.field4
	field4 int `validate:"min:3"`
}

var defaultOptions2 = Options2{
	field1: 1,
	field2: 2,
	field3: 3,
	field4: 4,
}

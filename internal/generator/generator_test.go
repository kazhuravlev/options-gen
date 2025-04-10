package generator_test

import (
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/kazhuravlev/options-gen/internal/generator"
	"github.com/kazhuravlev/options-gen/internal/generator/testdata"

	// test named imports.
	req "github.com/stretchr/testify/require"
)

const gofile = "generator_test.go"

func TestGetImports(t *testing.T) {
	t.Parallel()

	imports, err := generator.GetFileImports(gofile)
	req.NoError(t, err)

	requiredImports := []string{
		`"fmt"`,
		`"sort"`,
		`"testing"`,
		`"time"`,
		`"github.com/kazhuravlev/options-gen/internal/generator"`,
		`req "github.com/stretchr/testify/require"`,
	}
	sort.Strings(requiredImports)
	sort.Strings(imports)
	req.EqualValues(t, requiredImports, imports)
}

func TestGetOptionSpec(t *testing.T) { //nolint:funlen
	t.Parallel()

	spec, warnings, err := generator.GetOptionSpec(gofile, "TestOptions", "default")
	req.NoError(t, err)
	req.Equal(t, []string{
		"Deprecated: use `option:\"mandatory\"` instead for field `oldStyleOpt1` to force the passing option in the constructor argument\n",                                                                                                                                                                                                                                                                                                                                                                                                                                                                                              //nolint:lll
		"Deprecated: use github.com/go-playground/validator `validate` tag to check the field `oldStyleOpt1` content\n",                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                  //nolint:lll
		"Deprecated: use `option:\"mandatory\"` instead for field `oldStyleOpt2` to force the passing option in the constructor argument\n", "Deprecated: use github.com/go-playground/validator `validate` tag to check the field `oldStyleOpt2` content\n", "Deprecated: use `option:\"mandatory\"` instead for field `oldStyleOpt3` to force the passing option in the constructor argument\n", "Deprecated: use github.com/go-playground/validator `validate` tag to check the field `oldStyleOpt3` content\n", "Warning: consider to make `PublicOption1` is private. This is will not allow to users to avoid constructor method.", //nolint:lll
		"Warning: consider to make `PublicOption2` is private. This is will not allow to users to avoid constructor method.", //nolint:lll
	}, warnings)
	req.Equal(t, &generator.OptionSpec{
		TypeParamsSpec: "",
		TypeParams:     "",
		Options: []generator.OptionMeta{
			{
				Name:      "Stringer",
				Docstring: "// stringer bla-bla",
				Field:     "stringer",
				Type:      "fmt.Stringer",
				TagOption: generator.TagOption{IsRequired: true, GoValidator: "required", Default: ""},
			},
			{
				Name:      "Str",
				Docstring: "// comment-without-field-name-mention",
				Field:     "str",
				Type:      "string",
				TagOption: generator.TagOption{IsRequired: false, GoValidator: "required", Default: ""},
			},
			{
				Name:      "SomeMap",
				Docstring: "",
				Field:     "someMap",
				Type:      "map[string]string",
				TagOption: generator.TagOption{IsRequired: true, GoValidator: "required", Default: ""},
			},
			{
				Name:      "BoolTrue",
				Docstring: "",
				Field:     "boolTrue",
				Type:      "bool",
				TagOption: generator.TagOption{IsRequired: false, GoValidator: "", Default: "true"},
			},
			{
				Name:      "BoolFalse",
				Docstring: "",
				Field:     "boolFalse",
				Type:      "bool",
				TagOption: generator.TagOption{IsRequired: false, GoValidator: "", Default: "false"},
			},
			{
				Name:      "NoValidation",
				Docstring: "// multi\n// line\n// \n// comment",
				Field:     "noValidation",
				Type:      "string",
				TagOption: generator.TagOption{IsRequired: false, GoValidator: "", Default: ""},
			},
			{
				Name:      "StarOpt",
				Docstring: "",
				Field:     "starOpt",
				Type:      "*int",
				TagOption: generator.TagOption{IsRequired: true, GoValidator: "", Default: ""},
			},
			{
				Name:      "SliceOpt",
				Docstring: "",
				Field:     "sliceOpt",
				Type:      "[]int",
				TagOption: generator.TagOption{IsRequired: true, GoValidator: "", Default: ""},
			},
			{
				Name:      "OldStyleOpt1",
				Docstring: "",
				Field:     "oldStyleOpt1",
				Type:      "string",
				TagOption: generator.TagOption{IsRequired: true, GoValidator: "required", Default: ""},
			},
			{
				Name:      "OldStyleOpt2",
				Docstring: "",
				Field:     "oldStyleOpt2",
				Type:      "string",
				TagOption: generator.TagOption{IsRequired: true, GoValidator: "required", Default: ""},
			},
			{
				Name:      "OldStyleOpt3",
				Docstring: "",
				Field:     "oldStyleOpt3",
				Type:      "string",
				TagOption: generator.TagOption{IsRequired: true, GoValidator: "min=10,required", Default: ""},
			},
			{
				Name:      "PublicOption1",
				Docstring: "",
				Field:     "PublicOption1",
				Type:      "int",
				TagOption: generator.TagOption{IsRequired: true, GoValidator: "", Default: ""},
			},
			{
				Name:      "PublicOption2",
				Docstring: "",
				Field:     "PublicOption2",
				Type:      "int",
				TagOption: generator.TagOption{IsRequired: false, GoValidator: "", Default: ""},
			},
			{
				Name:      "WithDefaultValue",
				Docstring: "",
				Field:     "withDefaultValue",
				Type:      "time.Duration",
				TagOption: generator.TagOption{IsRequired: false, GoValidator: "", Default: "1m"},
			},
		},
	}, spec)
}

func TestGetOptionSpec_Generics(t *testing.T) {
	t.Parallel()

	spec, warnings, err := generator.GetOptionSpec(gofile, "TestOptionsGen", "default")
	req.NoError(t, err)
	req.Empty(t, warnings)
	req.Equal(t, &generator.OptionSpec{
		TypeParamsSpec: "[T1 int | string, T2, T3 any]",
		TypeParams:     "[T1, T2, T3]",
		Options: []generator.OptionMeta{
			{
				Name:      "Opt1",
				Docstring: "",
				Field:     "opt1",
				Type:      "T1",
				TagOption: generator.TagOption{IsRequired: true, GoValidator: "", Default: ""},
			},
			{
				Name:      "Opt2",
				Docstring: "",
				Field:     "opt2",
				Type:      "T2",
				TagOption: generator.TagOption{IsRequired: true, GoValidator: "required", Default: ""},
			},
			{
				Name:      "Opt3",
				Docstring: "",
				Field:     "opt3",
				Type:      "int",
				TagOption: generator.TagOption{IsRequired: false, GoValidator: "min=10", Default: ""},
			},
			{
				Name:      "Opt4",
				Docstring: "",
				Field:     "opt4",
				Type:      "T3",
				TagOption: generator.TagOption{IsRequired: false, GoValidator: "", Default: ""},
			},
		},
	}, spec)
}

// NOTE: this structs is used by testcases in current file

type TestOptions struct {
	// stringer bla-bla
	stringer fmt.Stringer `option:"mandatory" validate:"required"` //nolint:unused
	// comment-without-field-name-mention
	str string `validate:"required"` //nolint:unused

	someMap   map[string]string `option:"mandatory" validate:"required"` //nolint:unused
	boolTrue  bool              `default:"true"`                         //nolint:unused
	boolFalse bool              `default:"false"`                        //nolint:unused
	// multi
	// line
	//
	// comment
	noValidation string //nolint:unused
	starOpt      *int   `option:"mandatory"` //nolint:unused
	sliceOpt     []int  `option:"mandatory"` //nolint:unused

	oldStyleOpt1 string `option:"required,not-empty"`                     //nolint:unused
	oldStyleOpt2 string `option:"required,not-empty" validate:"required"` //nolint:unused
	oldStyleOpt3 string `option:"required,not-empty" validate:"min=10"`   //nolint:unused

	PublicOption1 int `option:"mandatory"`
	PublicOption2 int

	withDefaultValue time.Duration `default:"1m"` //nolint:unused
}

type TestOptionsGen[T1 int | string, T2, T3 any] struct {
	opt1 T1  `option:"mandatory"`                     //nolint:unused
	opt2 T2  `option:"mandatory" validate:"required"` //nolint:unused
	opt3 int `validate:"min=10"`                      //nolint:unused
	opt4 T3  //nolint:unused
}

type TestOptionsInline struct {
	InlineStruct struct {
		Field1 string
	}
}

type TestOptionsInlinePtr struct {
	InlineStruct *struct {
		Field1 string
	}
}

type EmbedStruct struct {
	String string
}

type TestOptionsEmbed struct {
	EmbedStruct
}

type TestOptionsEmbedPtr struct {
	*EmbedStruct
}

type TestOptionsEmbedAnotherPkg struct {
	testdata.StructForEmbed
}

type TestOptionsEmbedAnotherPkgPtr struct {
	*testdata.StructForEmbed
}

func TestGetOptionSpecInline(t *testing.T) { //nolint:funlen
	t.Parallel()

	spec, warnings, err := generator.GetOptionSpec(gofile, "TestOptionsInline", "default")
	req.NoError(t, err)
	req.Equal(t, []string{
		"Warning: consider to make `InlineStruct` is private. This is will not allow to users to avoid constructor method.",
	}, warnings)
	req.Equal(t, &generator.OptionSpec{
		TypeParamsSpec: "",
		TypeParams:     "",
		Options: []generator.OptionMeta{
			{
				Name:  "InlineStruct",
				Field: "InlineStruct",
				Type:  "struct{Field1 string}",
			},
		},
	}, spec)
}

func TestGetOptionSpecInlinePtr(t *testing.T) { //nolint:funlen
	t.Parallel()

	spec, warnings, err := generator.GetOptionSpec(gofile, "TestOptionsInlinePtr", "default")
	req.NoError(t, err)
	req.Equal(t, []string{
		"Warning: consider to make `InlineStruct` is private. This is will not allow to users to avoid constructor method.",
	}, warnings)
	req.Equal(t, &generator.OptionSpec{
		TypeParamsSpec: "",
		TypeParams:     "",
		Options: []generator.OptionMeta{
			{
				Name:  "InlineStruct",
				Field: "InlineStruct",
				Type:  "*struct{Field1 string}",
			},
		},
	}, spec)
}

func TestGetOptionSpecEmbed(t *testing.T) { //nolint:funlen
	t.Parallel()

	spec, warnings, err := generator.GetOptionSpec(gofile, "TestOptionsEmbed", "default")
	req.NoError(t, err)
	req.Equal(t, []string{
		"Warning: consider to make `EmbedStruct` is private. This is will not allow to users to avoid constructor method.",
	}, warnings)
	req.Equal(t, &generator.OptionSpec{
		TypeParamsSpec: "",
		TypeParams:     "",
		Options: []generator.OptionMeta{
			{
				Name:  "EmbedStruct",
				Field: "EmbedStruct",
				Type:  "EmbedStruct",
			},
		},
	}, spec)
}

func TestGetOptionSpecEmbedPtr(t *testing.T) { //nolint:funlen
	t.Parallel()

	spec, warnings, err := generator.GetOptionSpec(gofile, "TestOptionsEmbedPtr", "default")
	req.NoError(t, err)
	req.Equal(t, []string{
		"Warning: consider to make `EmbedStruct` is private. This is will not allow to users to avoid constructor method.",
	}, warnings)
	req.Equal(t, &generator.OptionSpec{
		TypeParamsSpec: "",
		TypeParams:     "",
		Options: []generator.OptionMeta{
			{
				Name:  "EmbedStruct",
				Field: "EmbedStruct",
				Type:  "*EmbedStruct",
			},
		},
	}, spec)
}

func TestGetOptionSpecEmbedAnotherPkg(t *testing.T) { //nolint:funlen
	t.Parallel()

	spec, warnings, err := generator.GetOptionSpec(gofile, "TestOptionsEmbedAnotherPkg", "default")
	req.NoError(t, err)
	req.Equal(t, []string{
		"Warning: consider to make `StructForEmbed` is private. This is will not allow to users to avoid constructor method.",
	}, warnings)
	req.Equal(t, &generator.OptionSpec{
		TypeParamsSpec: "",
		TypeParams:     "",
		Options: []generator.OptionMeta{
			{
				Name:  "StructForEmbed",
				Field: "StructForEmbed",
				Type:  "testdata.StructForEmbed",
			},
		},
	}, spec)
}

func TestGetOptionSpecEmbedAnotherPkgPtr(t *testing.T) { //nolint:funlen
	t.Parallel()

	spec, warnings, err := generator.GetOptionSpec(gofile, "TestOptionsEmbedAnotherPkgPtr", "default")
	req.NoError(t, err)
	req.Equal(t, []string{
		"Warning: consider to make `StructForEmbed` is private. This is will not allow to users to avoid constructor method.",
	}, warnings)
	req.Equal(t, &generator.OptionSpec{
		TypeParamsSpec: "",
		TypeParams:     "",
		Options: []generator.OptionMeta{
			{
				Name:  "StructForEmbed",
				Field: "StructForEmbed",
				Type:  "*testdata.StructForEmbed",
			},
		},
	}, spec)
}

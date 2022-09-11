package generator_test

import (
	"fmt"
	"sort"
	"testing"

	"github.com/kazhuravlev/options-gen/internal/generator"
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
		`"github.com/kazhuravlev/options-gen/internal/generator"`,
		`req "github.com/stretchr/testify/require"`,
	}
	sort.Strings(requiredImports)
	sort.Strings(imports)
	req.EqualValues(t, requiredImports, imports)
}

func TestGetOptionSpec(t *testing.T) { //nolint:funlen
	t.Parallel()

	spec, err := generator.GetOptionSpec(gofile, "TestOptions")
	req.NoError(t, err)
	req.Equal(t, &generator.OptionSpec{
		TypeParamsSpec: "",
		TypeParams:     "",
		Options: []generator.OptionMeta{
			{
				Name:      "Stringer",
				Field:     "stringer",
				Type:      "fmt.Stringer",
				TagOption: generator.TagOption{IsRequired: true, GoValidator: "required"},
			},
			{
				Name:      "Str",
				Field:     "str",
				Type:      "string",
				TagOption: generator.TagOption{IsRequired: false, GoValidator: "required"},
			},
			{
				Name:      "SomeMap",
				Field:     "someMap",
				Type:      "map[string]string",
				TagOption: generator.TagOption{IsRequired: true, GoValidator: "required"},
			},
			{
				Name:      "NoValidation",
				Field:     "noValidation",
				Type:      "string",
				TagOption: generator.TagOption{IsRequired: false, GoValidator: ""},
			},
			{
				Name:      "StarOpt",
				Field:     "starOpt",
				Type:      "*int",
				TagOption: generator.TagOption{IsRequired: true, GoValidator: ""},
			},
			{
				Name:      "SliceOpt",
				Field:     "sliceOpt",
				Type:      "[]int",
				TagOption: generator.TagOption{IsRequired: true, GoValidator: ""},
			},
			{
				Name:      "OldStyleOpt1",
				Field:     "oldStyleOpt1",
				Type:      "string",
				TagOption: generator.TagOption{IsRequired: true, GoValidator: "required"},
			},
			{
				Name:      "OldStyleOpt2",
				Field:     "oldStyleOpt2",
				Type:      "string",
				TagOption: generator.TagOption{IsRequired: true, GoValidator: "required"},
			},
			{
				Name:      "OldStyleOpt3",
				Field:     "oldStyleOpt3",
				Type:      "string",
				TagOption: generator.TagOption{IsRequired: true, GoValidator: "min=10,required"},
			},
		},
	}, spec)
}

func TestGetOptionSpec_Generics(t *testing.T) {
	t.Parallel()

	spec, err := generator.GetOptionSpec(gofile, "TestOptionsGen")
	req.NoError(t, err)
	req.Equal(t, &generator.OptionSpec{
		TypeParamsSpec: "[T1 int | string,T2 any]",
		TypeParams:     "[T1,T2]",
		Options: []generator.OptionMeta{
			{
				Name:      "Opt1",
				Field:     "opt1",
				Type:      "T1",
				TagOption: generator.TagOption{IsRequired: true, GoValidator: ""},
			},
			{
				Name:      "Opt2",
				Field:     "opt2",
				Type:      "T2",
				TagOption: generator.TagOption{IsRequired: true, GoValidator: "required"},
			},
			{
				Name:      "Opt3",
				Field:     "opt3",
				Type:      "int",
				TagOption: generator.TagOption{IsRequired: false, GoValidator: "min=10"},
			},
		},
	}, spec)
}

// NOTE: this structs is used by testcases in current file

type TestOptions struct {
	stringer     fmt.Stringer      `option:"mandatory" validate:"required"` //nolint:unused
	str          string            `validate:"required"`                    //nolint:unused
	someMap      map[string]string `option:"mandatory" validate:"required"` //nolint:unused
	noValidation string            //nolint:unused
	starOpt      *int              `option:"mandatory"` //nolint:unused
	sliceOpt     []int             `option:"mandatory"` //nolint:unused

	oldStyleOpt1 string `option:"required,not-empty"`                     //nolint:unused
	oldStyleOpt2 string `option:"required,not-empty" validate:"required"` //nolint:unused
	oldStyleOpt3 string `option:"required,not-empty" validate:"min=10"`   //nolint:unused
}

type TestOptionsGen[T1 int | string, T2 any] struct {
	opt1 T1  `option:"mandatory"`                     //nolint:unused
	opt2 T2  `option:"mandatory" validate:"required"` //nolint:unused
	opt3 int `validate:"min=10"`                      //nolint:unused
}

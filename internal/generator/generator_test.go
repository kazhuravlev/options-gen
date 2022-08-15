package generator_test

import (
	"fmt"
	"github.com/kazhuravlev/options-gen/internal/generator"
	"sort"
	"testing"

	// test named imports.
	req "github.com/stretchr/testify/require"
)

const (
	gofile        = "generator_test.go"
	optionsStruct = "TestOptions"
)

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

	data, err := generator.GetOptionSpec(gofile, optionsStruct)
	req.NoError(t, err)
	req.Equal(t, []generator.OptionMeta{
		{
			Name:  "Stringer",
			Field: "stringer",
			Type:  "fmt.Stringer",
			TagOption: generator.TagOption{
				IsRequired:  true,
				IsNotEmpty:  false,
				GoValidator: "required",
			},
		},
		{
			Name:  "Str",
			Field: "str",
			Type:  "string",
			TagOption: generator.TagOption{
				IsRequired:  false,
				IsNotEmpty:  false,
				GoValidator: "required",
			},
		},
		{
			Name:  "SomeMap",
			Field: "someMap",
			Type:  "map[string]string",
			TagOption: generator.TagOption{
				IsRequired:  true,
				IsNotEmpty:  false,
				GoValidator: "required",
			},
		},
		{
			Name:  "NoValidation",
			Field: "noValidation",
			Type:  "string",
			TagOption: generator.TagOption{
				IsRequired:  false,
				IsNotEmpty:  false,
				GoValidator: "",
			},
		},
		{
			Name:      "StarOpt",
			Field:     "starOpt",
			Type:      "*int",
			TagOption: generator.TagOption{IsRequired: true, IsNotEmpty: false, GoValidator: ""},
		},
		{
			Name:      "SliceOpt",
			Field:     "sliceOpt",
			Type:      "[]int",
			TagOption: generator.TagOption{IsRequired: true, IsNotEmpty: false, GoValidator: ""},
		},
		{
			Name:      "OldStyleOpt",
			Field:     "oldStyleOpt",
			Type:      "string",
			TagOption: generator.TagOption{IsRequired: true, IsNotEmpty: true, GoValidator: ""},
		},
	}, data)
}

// NOTE: this struct is used by testcases in current file

type TestOptions struct {
	stringer     fmt.Stringer      `option:"mandatory" validate:"required"` //nolint:unused
	str          string            `validate:"required"`                    //nolint:unused
	someMap      map[string]string `option:"mandatory" validate:"required"` //nolint:unused
	noValidation string            //nolint:unused
	starOpt      *int              `option:"mandatory"`          //nolint:unused
	sliceOpt     []int             `option:"mandatory"`          //nolint:unused
	oldStyleOpt  string            `option:"required,not-empty"` //nolint:unused
}

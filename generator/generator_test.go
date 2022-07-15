package generator

import (
	"fmt"
	"testing"

	// test named imports.
	req "github.com/stretchr/testify/require"
)

const gofile = "generator_test.go"
const optionsStruct = "TestOptions"

func TestGetImports(t *testing.T) {
	imports, err := GetFileImports(gofile)
	req.NoError(t, err)
	requiredImports := []string{
		`"fmt"`,
		`"testing"`,
		`req "github.com/stretchr/testify/require"`,
	}
	req.Equal(t, requiredImports, imports)
}

func TestGetOptionSpec(t *testing.T) {
	data, err := GetOptionSpec(gofile, optionsStruct)
	req.NoError(t, err)
	req.Equal(t, []OptionMeta{
		{
			Name:  "Stringer",
			Field: "stringer",
			Type:  "fmt.Stringer",
			TagOption: TagOption{
				IsRequired:  true,
				IsNotEmpty:  false,
				GoValidator: "required",
			},
		},
		{
			Name:  "Str",
			Field: "str",
			Type:  "string",
			TagOption: TagOption{
				IsRequired:  false,
				IsNotEmpty:  false,
				GoValidator: "required",
			},
		},
		{
			Name:  "SomeMap",
			Field: "someMap",
			Type:  "map[string]string",
			TagOption: TagOption{
				IsRequired:  true,
				IsNotEmpty:  false,
				GoValidator: "required",
			},
		},
		{
			Name:  "NoValidation",
			Field: "noValidation",
			Type:  "string",
			TagOption: TagOption{
				IsRequired:  false,
				IsNotEmpty:  false,
				GoValidator: "",
			},
		},
	}, data)
}

func TestToStopCIFromComplaining(t *testing.T) {
	s := TestOptions{
		stringer:     nil,
		str:          "123",
		noValidation: "qwe",
	}
	req.Equal(t, TestOptions{
		stringer:     nil,
		str:          "123",
		noValidation: "qwe",
	}, s)
}

type TestOptions struct {
	stringer     fmt.Stringer      `option:"mandatory" validate:"required"`
	str          string            `validate:"required"`
	someMap      map[string]string `option:"mandatory" validate:"required"`
	noValidation string
}

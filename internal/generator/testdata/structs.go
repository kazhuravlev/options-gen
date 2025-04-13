package testdata

import (
	"github.com/go-playground/validator/v10"
	"github.com/kazhuravlev/options-gen/internal/generator/testdata/subpkg"
)

type StructForEmbed struct {
	String string
}

type Int32s []int32

type RefType subpkg.Slice

type RefExtSliceType validator.ValidationErrors
type RefExtSliceType2 []validator.FieldError

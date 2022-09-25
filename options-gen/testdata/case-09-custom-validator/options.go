package testcase

import "github.com/go-playground/validator/v10"

var v = validator.New()

func init() {
	if err := v.RegisterValidation("adult", func(fl validator.FieldLevel) bool {
		return fl.Field().Int() >= 18
	}); err != nil {
		panic(err)
	}
}

type Options struct {
	amount int `option:"mandatory"`
	age    int `option:"mandatory" validate:"adult"`
}

func (Options) Validator() *validator.Validate {
	return v
}

package testcase

import goplvalidator "github.com/go-playground/validator/v10"

var v = goplvalidator.New()

func init() {
	if err := v.RegisterValidation("adult", func(fl goplvalidator.FieldLevel) bool {
		return fl.Field().Int() >= 18
	}); err != nil {
		panic(err)
	}
}

type Options struct {
	amount int `option:"mandatory"`
	age    int `option:"mandatory" validate:"adult"`
}

func (Options) Validator() *goplvalidator.Validate {
	return v
}

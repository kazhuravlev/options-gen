package validator

import goplvalidator "github.com/go-playground/validator/v10"

var Validator = goplvalidator.New()

type validatorProvider interface {
	Validator() *goplvalidator.Validate
}

func GetProvidedValidatorOrDefault(opts any) *goplvalidator.Validate {
	if v, ok := opts.(validatorProvider); ok {
		return v.Validator()
	}
	return Validator
}

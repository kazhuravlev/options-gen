package validator

import goplvalidator "github.com/go-playground/validator/v10"

var validator = goplvalidator.New()

// Set sets global validator used in options-gen generated code.
// Panics if provided validator is nil.
func Set(v *goplvalidator.Validate) {
	if v == nil {
		panic("incorrect use of validator.Set: unexpected nil")
	}
	validator = v
}

type validatorProvider interface {
	Validator() *goplvalidator.Validate
}

// GetValidatorFor returns validator provided by opts
// or default global validator else.
func GetValidatorFor(opts any) *goplvalidator.Validate {
	if v, ok := opts.(validatorProvider); ok {
		return v.Validator()
	}

	return validator
}

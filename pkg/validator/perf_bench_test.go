package validator_test

import (
	"testing"

	goplvalidator "github.com/go-playground/validator/v10"
	"github.com/kazhuravlev/options-gen/pkg/validator"
)

var benchValidatorSink *goplvalidator.Validate

// BenchmarkGetValidatorForNil benchmarks getting default validator with nil input
func BenchmarkGetValidatorForNil(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		benchValidatorSink = validator.GetValidatorFor(nil)
	}
}

// BenchmarkGetValidatorForCustom benchmarks getting custom validator from options
func BenchmarkGetValidatorForCustom(b *testing.B) {
	customOpts := &validatorProvider{}

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		benchValidatorSink = validator.GetValidatorFor(customOpts)
	}
}

// BenchmarkGetValidatorForWithoutProvider benchmarks getting default validator from non-provider struct
func BenchmarkGetValidatorForWithoutProvider(b *testing.B) {
	opts := &struct {
		field string
	}{}

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		benchValidatorSink = validator.GetValidatorFor(opts)
	}
}

// BenchmarkValidatorSet benchmarks setting a new global validator
func BenchmarkValidatorSet(b *testing.B) {
	v := goplvalidator.New()
	originalValidator := validator.GetValidatorFor(nil)

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		validator.Set(v)
	}

	// Restore original validator after test
	validator.Set(originalValidator)
}

type validatorProvider struct{}

func (vp *validatorProvider) Validator() *goplvalidator.Validate {
	return customValidator
}

var customValidator = goplvalidator.New()

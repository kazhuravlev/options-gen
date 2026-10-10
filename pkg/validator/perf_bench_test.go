package validator_test

import (
	"testing"

	goplvalidator "github.com/go-playground/validator/v10"
	"github.com/kazhuravlev/options-gen/pkg/validator"
)

var benchValidatorSink *goplvalidator.Validate

// GetValidatorFor runs inside every generated _validate_* function, so these
// benchmarks guard the fast path (a type assertion and a global read, no allocs).

// BenchmarkGetValidatorForNil benchmarks getting the default validator for nil input.
func BenchmarkGetValidatorForNil(b *testing.B) {
	b.ReportAllocs()

	for b.Loop() {
		benchValidatorSink = validator.GetValidatorFor(nil)
	}
}

// BenchmarkGetValidatorForCustom benchmarks getting a custom validator from an
// options struct that implements the provider interface.
func BenchmarkGetValidatorForCustom(b *testing.B) {
	customOpts := &validatorProvider{}

	b.ReportAllocs()

	for b.Loop() {
		benchValidatorSink = validator.GetValidatorFor(customOpts)
	}
}

// BenchmarkGetValidatorForWithoutProvider benchmarks getting the default validator
// for an options struct without a provider method.
func BenchmarkGetValidatorForWithoutProvider(b *testing.B) {
	opts := &struct {
		field string
	}{field: ""}

	b.ReportAllocs()

	for b.Loop() {
		benchValidatorSink = validator.GetValidatorFor(opts)
	}
}

type validatorProvider struct{}

func (vp *validatorProvider) Validator() *goplvalidator.Validate {
	return customValidator
}

var customValidator = goplvalidator.New()

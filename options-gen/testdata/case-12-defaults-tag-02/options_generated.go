// Code generated by options-gen. DO NOT EDIT.
package testcase

import (
	fmt461e464ebed9 "fmt"
	"time"

	errors461e464ebed9 "github.com/kazhuravlev/options-gen/pkg/errors"
	validator461e464ebed9 "github.com/kazhuravlev/options-gen/pkg/validator"
)

type OptOptionsSetter func(o *Options)

func NewOptions(
	options ...OptOptionsSetter,
) Options {
	o := Options{}

	// Setting defaults from field tag (if present)
	o.pingPeriod, _ = time.ParseDuration("3s")
	o.name = "unknown"
	o.maxAttempts = 10
	o.eps = 0.0001

	for _, opt := range options {
		opt(&o)
	}
	return o
}

func WithSomePingPeriod(opt time.Duration) OptOptionsSetter {
	return func(o *Options) {
		o.pingPeriod = opt
	}
}

func WithSomeName(opt string) OptOptionsSetter {
	return func(o *Options) {
		o.name = opt
	}
}

func WithSomeMaxAttempts(opt int) OptOptionsSetter {
	return func(o *Options) {
		o.maxAttempts = opt
	}
}

func WithSomeEps(opt float32) OptOptionsSetter {
	return func(o *Options) {
		o.eps = opt
	}
}

func (o *Options) Validate() error {
	errs := new(errors461e464ebed9.ValidationErrors)
	errs.Add(errors461e464ebed9.NewValidationError("pingPeriod", _validate_Options_pingPeriod(o)))
	errs.Add(errors461e464ebed9.NewValidationError("name", _validate_Options_name(o)))
	errs.Add(errors461e464ebed9.NewValidationError("maxAttempts", _validate_Options_maxAttempts(o)))
	errs.Add(errors461e464ebed9.NewValidationError("eps", _validate_Options_eps(o)))
	return errs.AsError()
}

func _validate_Options_pingPeriod(o *Options) error {
	if err := validator461e464ebed9.GetValidatorFor(o).Var(o.pingPeriod, "min=100ms,max=30s"); err != nil {
		return fmt461e464ebed9.Errorf("field `pingPeriod` did not pass the test: %w", err)
	}
	return nil
}

func _validate_Options_name(o *Options) error {
	if err := validator461e464ebed9.GetValidatorFor(o).Var(o.name, "required"); err != nil {
		return fmt461e464ebed9.Errorf("field `name` did not pass the test: %w", err)
	}
	return nil
}

func _validate_Options_maxAttempts(o *Options) error {
	if err := validator461e464ebed9.GetValidatorFor(o).Var(o.maxAttempts, "min=1,max=10"); err != nil {
		return fmt461e464ebed9.Errorf("field `maxAttempts` did not pass the test: %w", err)
	}
	return nil
}

func _validate_Options_eps(o *Options) error {
	if err := validator461e464ebed9.GetValidatorFor(o).Var(o.eps, "gt=0"); err != nil {
		return fmt461e464ebed9.Errorf("field `eps` did not pass the test: %w", err)
	}
	return nil
}

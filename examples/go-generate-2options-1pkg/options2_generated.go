// Code generated by options-gen. DO NOT EDIT.
package gogenerate

import (
	fmt461e464ebed9 "fmt"

	errors461e464ebed9 "github.com/kazhuravlev/options-gen/pkg/errors"
	validator461e464ebed9 "github.com/kazhuravlev/options-gen/pkg/validator"
)

type OptOptions2Setter func(o *Options2)

func NewOptions2(
	options ...OptOptions2Setter,
) Options2 {
	o := Options2{}

	// Setting defaults from variable
	o.field1 = defaultOptions2.field1
	o.field2 = defaultOptions2.field2
	o.field3 = defaultOptions2.field3
	o.field4 = defaultOptions2.field4

	for _, opt := range options {
		opt(&o)
	}
	return o
}

// Options2.field1
func WithNNNField1(opt int) OptOptions2Setter {
	return func(o *Options2) {
		o.field1 = opt
	}
}

// Options2.field2
func WithNNNField2(opt int) OptOptions2Setter {
	return func(o *Options2) {
		o.field2 = opt
	}
}

// Options2.field3
func WithNNNField3(opt int) OptOptions2Setter {
	return func(o *Options2) {
		o.field3 = opt
	}
}

// Options2.field4
func WithNNNField4(opt int) OptOptions2Setter {
	return func(o *Options2) {
		o.field4 = opt
	}
}

func (o *Options2) Validate() error {
	errs := new(errors461e464ebed9.ValidationErrors)
	errs.Add(errors461e464ebed9.NewValidationError("field1", _validate_Options2_field1(o)))
	errs.Add(errors461e464ebed9.NewValidationError("field2", _validate_Options2_field2(o)))
	errs.Add(errors461e464ebed9.NewValidationError("field3", _validate_Options2_field3(o)))
	errs.Add(errors461e464ebed9.NewValidationError("field4", _validate_Options2_field4(o)))
	return errs.AsError()
}

func _validate_Options2_field1(o *Options2) error {
	if err := validator461e464ebed9.GetValidatorFor(o).Var(o.field1, "min:3"); err != nil {
		return fmt461e464ebed9.Errorf("field `field1` did not pass the test: %w", err)
	}
	return nil
}

func _validate_Options2_field2(o *Options2) error {
	if err := validator461e464ebed9.GetValidatorFor(o).Var(o.field2, "min:3"); err != nil {
		return fmt461e464ebed9.Errorf("field `field2` did not pass the test: %w", err)
	}
	return nil
}

func _validate_Options2_field3(o *Options2) error {
	if err := validator461e464ebed9.GetValidatorFor(o).Var(o.field3, "min:3"); err != nil {
		return fmt461e464ebed9.Errorf("field `field3` did not pass the test: %w", err)
	}
	return nil
}

func _validate_Options2_field4(o *Options2) error {
	if err := validator461e464ebed9.GetValidatorFor(o).Var(o.field4, "min:3"); err != nil {
		return fmt461e464ebed9.Errorf("field `field4` did not pass the test: %w", err)
	}
	return nil
}
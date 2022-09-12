// Code generated by options-gen. DO NOT EDIT.
package testcase

import (
	"fmt"

	goplvalidator "github.com/go-playground/validator/v10"
	errors461e464ebed9 "github.com/kazhuravlev/options-gen/pkg/errors"
)

var _validator461e464ebed9 = goplvalidator.New()

type optOptionsSetter[T string] func(o *Options[T])

func NewOptions[T string](
	RequiredKey T,
	Key T,
	options ...optOptionsSetter[T],
) Options[T] {
	o := Options[T]{}
	o.RequiredKey = RequiredKey
	o.Key = Key

	for _, opt := range options {
		opt(&o)
	}
	return o
}

func WithOptKey[T string](opt T) optOptionsSetter[T] {
	return func(o *Options[T]) {
		o.OptKey = opt
	}
}

func (o *Options[T]) Validate() error {
	errs := new(errors461e464ebed9.ValidationErrors)
	errs.Add(errors461e464ebed9.NewValidationError("RequiredKey", _validate_Options_RequiredKey[T](o)))
	return errs.AsError()
}

func _validate_Options_RequiredKey[T string](o *Options[T]) error {
	if err := _validator461e464ebed9.Var(o.RequiredKey, "required"); err != nil {
		return fmt.Errorf("field `RequiredKey` did not pass the test: %w", err)
	}
	return nil
}

// Code generated by options-gen. DO NOT EDIT.
package testcase

import (
	"fmt"

	goplvalidator "github.com/go-playground/validator/v10"
	errors461e464ebed9 "github.com/kazhuravlev/options-gen/pkg/errors"
)

var _validator461e464ebed9 = goplvalidator.New()

type optOptionsMeta[T string] struct {
	setter    func(o *Options[T])
	validator func(o *Options[T]) error
}

func NewOptions[T string](
	RequiredKey T,
	Key T,

	options ...optOptionsMeta[T],
) Options[T] {
	o := Options[T]{}
	o.RequiredKey = RequiredKey
	o.Key = Key

	for i := range options {
		options[i].setter(&o)
	}

	return o
}

func WithOptKey[T string](opt T) optOptionsMeta[T] {
	return optOptionsMeta[T]{
		setter:    func(o *Options[T]) { o.OptKey = opt },
		validator: _Options_OptKeyValidator[T],
	}
}

func (o *Options[T]) Validate() error {
	errs := new(errors461e464ebed9.ValidationErrors)

	errs.Add(errors461e464ebed9.NewValidationError("RequiredKey", _Options_RequiredKeyValidator(o)))
	errs.Add(errors461e464ebed9.NewValidationError("Key", _Options_KeyValidator(o)))
	errs.Add(errors461e464ebed9.NewValidationError("OptKey", _Options_OptKeyValidator(o)))
	return errs.AsError()
}

func _Options_RequiredKeyValidator[T string](o *Options[T]) error {
	if err := _validator461e464ebed9.Var(o.RequiredKey, "required"); err != nil {
		return fmt.Errorf("field `RequiredKey` did not pass the test: %w", err)
	}
	return nil
}

func _Options_KeyValidator[T string](o *Options[T]) error {
	return nil
}

func _Options_OptKeyValidator[T string](o *Options[T]) error {
	return nil
}

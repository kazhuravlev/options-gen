// Code generated by options-gen. DO NOT EDIT.
package main

import (
	"fmt"

	subpackage "github.com/kazhuravlev/options-gen/examples/library/sub-package"
	"github.com/kazhuravlev/options-gen/validator"
)

type optOptionsMeta struct {
	setter    func(o *Options)
	validator func(o *Options) error
}

func _Options_service1Validator(o *Options) error {
	if validator.IsNil(o.service1) {
		return fmt.Errorf("%w: service1 must be set (type *subpackage.Service1)", ErrInvalidOption)
	}
	return nil
}

func NewOptions(
	service1 *subpackage.Service1,

	options ...optOptionsMeta,
) Options {
	o := Options{}
	o.service1 = service1

	for i := range options {
		options[i].setter(&o)
	}

	return o
}

func (o *Options) Validate() error {
	if err := _Options_service1Validator(o); err != nil {
		return fmt.Errorf("%w: invalid value for option WithService1", err)
	}

	return nil
}
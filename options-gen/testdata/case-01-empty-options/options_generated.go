// Code generated by options-gen. DO NOT EDIT.
package testcase

import (
	"golang.org/x/sync/errgroup"
)

type optOptionsMeta struct {
	setter    func(o *Options)
	validator func(o *Options) error
}

func NewOptions(

	options ...optOptionsMeta,
) Options {
	o := Options{}

	for i := range options {
		options[i].setter(&o)
	}

	return o
}

func (o *Options) Validate() error {
	g := new(errgroup.Group)

	return g.Wait()
}
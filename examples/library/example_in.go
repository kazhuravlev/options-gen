package main

import (
	"github.com/kazhuravlev/options-gen/examples/library/sub-package"
	"github.com/pkg/errors"
)

var ErrInvalidOption = errors.New("invalid option")

type Options struct {
	service1   *subpackage.Service1 `option:"mandatory" validate:"required"`
	s3Endpoint string               `option:"mandatory" validate:"required,url"`
	port       int                  `validate:"required,min=10"`
}

type Config struct {
	name string
}

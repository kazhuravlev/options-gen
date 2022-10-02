package main

import (
	"github.com/kazhuravlev/options-gen/examples/library/sub-package"
)

type Options struct {
	service1   *subpackage.Service1 `option:"mandatory" validate:"required"`
	s3Endpoint string               `option:"mandatory" validate:"required,url"`
	port       int                  `validate:"required,min=10"`
}

type Config struct {
	name string
}

type Params struct {
	hash string `option:"mandatory" validate:"hexadecimal"`
}

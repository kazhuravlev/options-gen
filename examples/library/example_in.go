package main

import (
	"github.com/kazhuravlev/options-gen/examples/library/sub-package"
	"github.com/pkg/errors"
)

var ErrInvalidOption = errors.New("invalid option")

type Options struct {
	service1 *subpackage.Service1 `option:"required,not-empty"`
}

type Config struct {
	name string
}

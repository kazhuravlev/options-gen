package testcase

import (
	"fmt"
	"io"
)

type localInterface interface {
	Hello()
}

type Options struct {
	Any      any                `option:"mandatory"`
	Stringer fmt.Stringer       `option:"mandatory"`
	RWCloser io.ReadWriteCloser `option:"mandatory"`
	Local    localInterface     `option:"mandatory"`

	OptAny      any
	OptStringer fmt.Stringer
	OptRWCloser io.ReadWriteCloser
	OptLocal    localInterface
}

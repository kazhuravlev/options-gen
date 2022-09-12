package testcase

import (
	"fmt"
	"io"
)

type localInterface interface {
	Hello()
}

type Options struct {
	any      any                `option:"mandatory"`
	stringer fmt.Stringer       `option:"mandatory"`
	rWCloser io.ReadWriteCloser `option:"mandatory"`
	local    localInterface     `option:"mandatory"`

	optAny      any
	optStringer fmt.Stringer
	optRWCloser io.ReadWriteCloser
	optLocal    localInterface
}

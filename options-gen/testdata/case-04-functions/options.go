package testcase

import (
	"net/http"
)

type FnType func() (int, error)
type localFnType func(a, b, c string) error

type Options struct {
	FnTypeParam FnType                          `option:"mandatory"`
	FnParam     func(server *http.Server) error `option:"mandatory"`
	HandlerFunc http.HandlerFunc                `option:"mandatory"`
	Local       localFnType                     `option:"mandatory"`

	OptFnTypeParam FnType
	OptFnParam     func(server *http.Server) error
	OptHandlerFunc http.HandlerFunc
	OptLocal       localFnType
}

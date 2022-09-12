package testcase

import (
	"net/http"
)

type FnType func() (int, error)
type localFnType func(a, b, c string) error

type Options struct {
	fnTypeParam FnType                                       `option:"mandatory"`
	fnParam     func(server *http.Server) error              `option:"mandatory"`
	handlerFunc http.HandlerFunc                             `option:"mandatory"`
	middleware  func(next http.HandlerFunc) http.HandlerFunc `option:"mandatory"`
	local       localFnType                                  `option:"mandatory"`

	optFnTypeParam FnType
	optFnParam     func(server *http.Server) error
	optHandlerFunc http.HandlerFunc
	optMiddleware  func(next http.HandlerFunc) http.HandlerFunc
	optLocal       localFnType
}

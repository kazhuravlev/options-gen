package testcase

import "net/http"

type Options[KeyT int | string, TT any] struct {
	requiredHandler http.Handler `option:"mandatory" validate:"required"`
	requiredKey     KeyT         `option:"mandatory" validate:"required"`

	handler http.Handler `option:"mandatory"`
	key     KeyT         `option:"mandatory"`

	optHandler http.Handler
	optKey     KeyT

	anyOpt TT
}

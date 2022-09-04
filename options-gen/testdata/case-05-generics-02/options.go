package testcase

import "net/http"

type Options[KeyT int | string, TT any] struct {
	RequiredHandler http.Handler `option:"mandatory" validate:"required"`
	RequiredKey     KeyT         `option:"mandatory" validate:"required"`

	Handler http.Handler `option:"mandatory"`
	Key     KeyT         `option:"mandatory"`

	OptHandler http.Handler
	OptKey     KeyT

	AnyOpt TT
}

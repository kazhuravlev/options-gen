// Code generated by options-gen qa-version. DO NOT EDIT.

package testcase

import (
	"net/http"
)

type OptOptionsSetter func(o *Options)

func NewOptions(
	fnTypeParam FnType,
	fnParam func(server *http.Server) error,
	handlerFunc http.HandlerFunc,
	middleware func(next http.HandlerFunc) http.HandlerFunc,
	local localFnType,
	options ...OptOptionsSetter,
) Options {
	var o Options

	// Setting defaults from field tag (if present)

	o.fnTypeParam = fnTypeParam
	o.fnParam = fnParam
	o.handlerFunc = handlerFunc
	o.middleware = middleware
	o.local = local

	for _, opt := range options {
		opt(&o)
	}
	return o
}

func WithOptFnTypeParam(opt FnType) OptOptionsSetter {
	return func(o *Options) { o.optFnTypeParam = opt }
}

func WithOptFnParam(opt func(server *http.Server) error) OptOptionsSetter {
	return func(o *Options) { o.optFnParam = opt }
}

func WithOptHandlerFunc(opt http.HandlerFunc) OptOptionsSetter {
	return func(o *Options) { o.optHandlerFunc = opt }
}

func WithOptMiddleware(opt func(next http.HandlerFunc) http.HandlerFunc) OptOptionsSetter {
	return func(o *Options) { o.optMiddleware = opt }
}

func WithOptLocal(opt localFnType) OptOptionsSetter {
	return func(o *Options) { o.optLocal = opt }
}

func (o *Options) Validate() error {
	return nil
}

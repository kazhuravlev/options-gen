package gogenerate

import (
	"net/http"
)

//go:generate options-gen -from-struct=Options
type Options struct {
	httpClient *http.Client `option:"mandatory" validate:"required"`
	token      string       `option:"mandatory"`
	// Address that will be used for each request to the remote server.
	//
	// By default, it will be set to 127.0.0.1:8000
	addr string `default:"127.0.0.1:8000"`
}

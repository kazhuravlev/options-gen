package gogenerate

import (
	"net/http"
)

//go:generate options-gen -from-struct=Options
type Options struct {
	httpClient *http.Client `option:"mandatory" validate:"required"`
	token      string       `option:"mandatory"`
	addr       string       `default:"127.0.0.1:8000"`
}

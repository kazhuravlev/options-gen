package cli

import (
	"github.com/pkg/errors"
	"net/http"
)

var ErrInvalidOption = errors.New("invalid option")

//go:generate options-gen -out-filename=options_generated.go -from-struct=Options
type Options struct {
	httpClient *http.Client `option:"mandatory" validate:"required"`
	token      string       `option:"mandatory"`
}

package cli

import (
	"github.com/pkg/errors"
	"net/http"
)

var ErrInvalidOption = errors.New("invalid option")

//go:generate options-gen -filename=$GOFILE -out-filename=options_generated.go -pkg=cli -from-struct=Options
type Options struct {
	httpClient *http.Client `option:"required,not-empty"`
	token      string       `option:"required"`
}

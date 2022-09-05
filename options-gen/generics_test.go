package optionsgen_test

import (
	"net"
	"net/http"
	"testing"

	testcase "github.com/kazhuravlev/options-gen/options-gen/testdata/case-05-generics-02"
	"github.com/stretchr/testify/assert"
)

func TestGenericsOptions(t *testing.T) {
	t.Run("validation failed", func(t *testing.T) {
		opts := testcase.NewOptions[string, interface{ Timeout() bool }](
			nil,
			"string key",
			nil,
			"",
			testcase.WithAnyOpt[string, interface{ Timeout() bool }](net.Error(nil)),
		)
		assert.Error(t, opts.Validate())
	})

	t.Run("valid options", func(t *testing.T) {
		opts := testcase.NewOptions[int, float32](
			new(handlerMock),
			42,
			new(handlerMock),
			24,
			testcase.WithAnyOpt[int, float32](24.24),
			testcase.WithOptHandler[int, float32](new(handlerMock)),
		)
		assert.NoError(t, opts.Validate())
	})
}

var _ http.Handler = (*handlerMock)(nil)

type handlerMock struct{}

func (h *handlerMock) ServeHTTP(_ http.ResponseWriter, _ *http.Request) {
}

package optionsgen_test

import (
	"testing"

	testcase "github.com/kazhuravlev/options-gen/options-gen/testdata/case-09-custom-validator"
	"github.com/stretchr/testify/assert"
)

func TestOptionsWithCustomValidator(t *testing.T) {
	t.Run("valid options", func(t *testing.T) {
		opts := testcase.NewOptions(100, 19)
		assert.NoError(t, opts.Validate())
	})

	t.Run("invalid options", func(t *testing.T) {
		opts := testcase.NewOptions(100, 17)
		assert.Error(t, opts.Validate())
	})
}

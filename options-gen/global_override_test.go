package optionsgen_test

import (
	"testing"

	goplvalidator "github.com/go-playground/validator/v10"
	testcase "github.com/kazhuravlev/options-gen/options-gen/testdata/case-10-global-override"
	"github.com/kazhuravlev/options-gen/pkg/validator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOptionsWithOverridenValidator(t *testing.T) {
	v := goplvalidator.New()
	require.NoError(t, v.RegisterValidation("child", func(fl goplvalidator.FieldLevel) bool {
		return fl.Field().Int() < 14
	}))

	old := validator.GetValidatorFor(nil)
	validator.Set(v)
	t.Cleanup(func() { validator.Set(old) })

	t.Run("valid options", func(t *testing.T) {
		opts := testcase.NewOptions(100, 13)
		assert.NoError(t, opts.Validate())
	})

	t.Run("invalid options", func(t *testing.T) {
		opts := testcase.NewOptions(100, 14)
		assert.Error(t, opts.Validate())
	})
}

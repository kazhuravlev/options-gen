package validator_test

import (
	"testing"

	goplvalidator "github.com/go-playground/validator/v10"
	"github.com/kazhuravlev/options-gen/pkg/validator"
	"github.com/stretchr/testify/assert"
)

func TestGetValidatorFor(t *testing.T) {
	t.Run("set nil", func(t *testing.T) {
		assert.Panics(t, func() {
			validator.Set(nil)
		})
	})

	t.Run("default validator is not nil", func(t *testing.T) {
		v := validator.GetValidatorFor(nil)
		assert.NotNil(t, v)
	})

	t.Run("override validator on options level", func(t *testing.T) {
		v := validator.GetValidatorFor(new(options))
		assert.Equal(t, localValidator, v)
	})

	t.Run("override validator on global level", func(t *testing.T) {
		v := validator.GetValidatorFor(nil)
		defer validator.Set(v)

		validator.Set(localValidator)
		v1 := validator.GetValidatorFor(nil)
		assert.NotEqual(t, v1, v)
		assert.Equal(t, localValidator, v1)
	})
}

var localValidator = goplvalidator.New()

type options struct{}                                //
func (o options) Validator() *goplvalidator.Validate { return localValidator }

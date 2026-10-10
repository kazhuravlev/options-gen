package validator_test

import (
	"sync"
	"testing"

	goplvalidator "github.com/go-playground/validator/v10"
	"github.com/kazhuravlev/options-gen/pkg/validator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// nilProvider implements the provider interface but returns no validator.
type nilProvider struct{}

func (nilProvider) Validator() *goplvalidator.Validate { return nil }

// ptrProvider implements the provider interface only on the pointer receiver.
type ptrProvider struct{}

func (*ptrProvider) Validator() *goplvalidator.Validate { return ptrValidator }

var ptrValidator = goplvalidator.New()

const (
	concurrentReaders = 16
	readsPerGoroutine = 200
)

// NOTE: tests in this file mutate the package-level validator, so they are intentionally not parallel.
func TestGetValidatorFor_CornerCases(t *testing.T) {
	t.Run("provider returning nil wins over the global validator", func(t *testing.T) {
		// Documented behavior: a value that implements Validator() is trusted, even when it returns nil.
		// Generated code would then call methods on a nil validator, so providers must not return nil.
		assert.Nil(t, validator.GetValidatorFor(nilProvider{}))
		assert.NotNil(t, validator.GetValidatorFor(nil))
	})

	t.Run("pointer receiver provider is detected only through a pointer", func(t *testing.T) {
		assert.Same(t, ptrValidator, validator.GetValidatorFor(&ptrProvider{}))

		// A plain value does not implement the interface, so the global validator is used.
		global := validator.GetValidatorFor(nil)
		assert.Same(t, global, validator.GetValidatorFor(ptrProvider{}))
		assert.NotSame(t, ptrValidator, validator.GetValidatorFor(ptrProvider{}))
	})

	t.Run("non-provider values fall back to the global validator", func(t *testing.T) {
		global := validator.GetValidatorFor(nil)

		assert.Same(t, global, validator.GetValidatorFor(struct{}{}))
		assert.Same(t, global, validator.GetValidatorFor("string"))
		assert.Same(t, global, validator.GetValidatorFor(new(int)))
	})

	t.Run("Set replaces and restores the global validator", func(t *testing.T) {
		original := validator.GetValidatorFor(nil)
		t.Cleanup(func() { validator.Set(original) })

		custom := goplvalidator.New()
		validator.Set(custom)
		assert.Same(t, custom, validator.GetValidatorFor(nil))

		// A provider is unaffected by the global override.
		assert.Same(t, ptrValidator, validator.GetValidatorFor(&ptrProvider{}))

		validator.Set(original)
		assert.Same(t, original, validator.GetValidatorFor(nil))
	})

	t.Run("Set nil keeps the previous validator", func(t *testing.T) {
		before := validator.GetValidatorFor(nil)
		require.Panics(t, func() { validator.Set(nil) })
		assert.Same(t, before, validator.GetValidatorFor(nil))
	})

	t.Run("concurrent reads are race-free", func(t *testing.T) {
		expected := validator.GetValidatorFor(nil)

		var wg sync.WaitGroup
		mismatches := make(chan int, concurrentReaders)

		for range concurrentReaders {
			wg.Add(1)

			go func() {
				defer wg.Done()

				var bad int
				for range readsPerGoroutine {
					if validator.GetValidatorFor(nil) != expected {
						bad++
					}

					if validator.GetValidatorFor(&ptrProvider{}) != ptrValidator {
						bad++
					}
				}

				mismatches <- bad
			}()
		}

		wg.Wait()
		close(mismatches)

		for bad := range mismatches {
			assert.Zero(t, bad)
		}
	})
}

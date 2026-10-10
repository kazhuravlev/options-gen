package errors_test

import (
	stderrors "errors"
	"fmt"
	"io"
	"syscall"
	"testing"

	"github.com/kazhuravlev/options-gen/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// codeError is a typed error used to check errors.As through the aggregate.
type codeError struct {
	code int
}

func (e *codeError) Error() string {
	return fmt.Sprintf("code %d", e.code)
}

const (
	testCode      = 42
	testOtherCode = 7
)

func TestValidationErrors_Unwrap(t *testing.T) {
	t.Parallel()

	t.Run("errors.Is reaches every element through the aggregate", func(t *testing.T) {
		t.Parallel()

		errs := new(errors.ValidationErrors)
		errs.Add(errors.NewValidationError("field1", syscall.ENOENT))
		errs.Add(errors.NewValidationError("field2", fmt.Errorf("wrapped: %w", io.EOF)))

		aggregate := errs.AsError()
		require.ErrorIs(t, aggregate, syscall.ENOENT)
		require.ErrorIs(t, aggregate, io.EOF)
		require.NotErrorIs(t, aggregate, io.ErrUnexpectedEOF)

		// One more wrapping layer on top of the aggregate, like Run does with "bad configuration: %w".
		require.ErrorIs(t, fmt.Errorf("bad configuration: %w", aggregate), io.EOF)
	})

	t.Run("errors.As reaches the element type through the aggregate", func(t *testing.T) {
		t.Parallel()

		errs := new(errors.ValidationErrors)
		errs.Add(errors.NewValidationError("field1", io.EOF))
		errs.Add(errors.NewValidationError("field2", &codeError{code: testCode}))

		var target *codeError
		require.ErrorAs(t, errs.AsError(), &target)
		assert.Equal(t, testCode, target.code)

		// The aggregate itself is still matched by errors.As, also through a wrapper.
		var aggregate errors.ValidationErrors
		require.ErrorAs(t, fmt.Errorf("wrapped: %w", errs.AsError()), &aggregate)
		assert.Len(t, aggregate, 2)
	})

	t.Run("Unwrap returns one error per element in order", func(t *testing.T) {
		t.Parallel()

		errs := new(errors.ValidationErrors)
		errs.Add(errors.NewValidationError("a", io.EOF))
		errs.Add(errors.NewValidationError("b", syscall.ENOENT))

		unwrapped := errs.Unwrap()
		require.Len(t, unwrapped, 2)
		assert.EqualError(t, unwrapped[0], "(a): EOF")
		assert.EqualError(t, unwrapped[1], "(b): no such file or directory")
		assert.ErrorIs(t, unwrapped[0], io.EOF)
		assert.ErrorIs(t, unwrapped[1], syscall.ENOENT)
	})

	t.Run("empty aggregate unwraps to nil", func(t *testing.T) {
		t.Parallel()

		assert.Nil(t, errors.ValidationErrors{}.Unwrap())
		assert.Nil(t, errors.ValidationErrors(nil).Unwrap())
	})
}

func TestValidationError_Unwrap(t *testing.T) {
	t.Parallel()

	inner := &codeError{code: testOtherCode}
	err := errors.NewValidationError("field", inner)

	var target *codeError
	require.ErrorAs(t, err, &target)
	assert.Same(t, inner, target)
	assert.Same(t, inner, stderrors.Unwrap(err))
	assert.EqualError(t, err, "(field): code 7")
}

func TestValidationErrors_CornerCases(t *testing.T) {
	t.Parallel()

	t.Run("single element has no separator", func(t *testing.T) {
		t.Parallel()

		errs := new(errors.ValidationErrors)
		errs.Add(errors.NewValidationError("field1", io.EOF))

		assert.Equal(t, "ValidationErrors: (field1): EOF", errs.Error())
		assert.EqualError(t, errs.AsError(), "ValidationErrors: (field1): EOF")
	})

	t.Run("Add nil is a no-op", func(t *testing.T) {
		t.Parallel()

		errs := new(errors.ValidationErrors)
		errs.Add(nil)
		errs.Add(errors.NewValidationError("field1", nil))

		assert.Empty(t, *errs)
		assert.NoError(t, errs.AsError())
		assert.Empty(t, errs.Errors())
	})

	t.Run("Errors returns a copy", func(t *testing.T) {
		t.Parallel()

		errs := new(errors.ValidationErrors)
		errs.Add(errors.NewValidationError("field1", io.EOF))
		errs.Add(errors.NewValidationError("field2", syscall.ENOENT))

		before := errs.Error()

		copied := errs.Errors()
		require.Len(t, copied, 2)
		copied[0], copied[1] = copied[1], copied[0]
		copied = copied[:1]

		assert.Len(t, *errs, 2)
		assert.Equal(t, before, errs.Error())
		assert.Len(t, copied, 1)
	})

	t.Run("AsError on nil pointer receiver panics", func(t *testing.T) {
		t.Parallel()

		// AsError has a value receiver, so a nil *ValidationErrors cannot be dereferenced.
		// Generated code always allocates the collection with new(ValidationErrors), which is safe.
		var errs *errors.ValidationErrors
		assert.Panics(t, func() { _ = errs.AsError() })
		assert.NoError(t, new(errors.ValidationErrors).AsError())
	})

	t.Run("AsError on empty value is nil, not a typed nil", func(t *testing.T) {
		t.Parallel()

		var errs errors.ValidationErrors
		assert.NoError(t, errs.AsError())
		assert.NoError(t, errors.ValidationErrors{}.AsError())
	})
}

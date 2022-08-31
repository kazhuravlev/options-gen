package errors_test

import (
	"fmt"
	"io"
	"syscall"
	"testing"

	"github.com/kazhuravlev/options-gen/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestValidationErrors(t *testing.T) {
	t.Parallel()

	errs := new(errors.ValidationErrors)
	assert.NoError(t, errs.AsError())

	assert.Equal(t, "", errs.Error())

	errs.Add(errors.NewValidationError("field1", syscall.ENOENT))
	errs.Add(errors.NewValidationError("field2", nil))
	errs.Add(errors.NewValidationError("field3-dupl", io.EOF))
	errs.Add(errors.NewValidationError("field3-dupl", io.EOF))

	assert.Error(t, errs.AsError())
	expErrStr := `ValidationErrors: (field1): no such file or directory; (field3-dupl): EOF; (field3-dupl): EOF`
	assert.Equal(t, expErrStr, errs.Error())
	assert.Len(t, errs.Errors(), 3)

	var err errors.ValidationErrors
	assert.ErrorAs(t, errs.AsError(), &err)
	assert.Len(t, err.Errors(), 3)
}

func TestValidationError(t *testing.T) {
	t.Parallel()

	err := errors.NewValidationError("field1", nil)
	// NOTE(kazhuravlev): We do not using NoError, because err has the
	//  type *validationError. see the next assertion.
	assert.Nil(t, err)
	assert.NotEqual(t, (*errors.ValidationErrors)(nil), error(nil))

	err2 := errors.NewValidationError("field1", io.EOF)
	assert.ErrorIs(t, err2, io.EOF)

	err3 := errors.NewValidationError("field1", fmt.Errorf("some error is occurs: %w", io.EOF))
	assert.ErrorIs(t, err3, io.EOF)
}

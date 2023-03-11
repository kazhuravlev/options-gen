package errors

import (
	"bytes"
	"errors"
	"fmt"
)

type validationError struct {
	fieldName string
	err       error
}

func NewValidationError(fieldName string, err error) *validationError { //nolint:revive
	if err == nil {
		return nil
	}

	return &validationError{
		fieldName: fieldName,
		err:       err,
	}
}

func (e *validationError) Error() string {
	return fmt.Sprintf("(%s): %s", e.fieldName, e.err.Error())
}

func (e *validationError) Is(err error) bool {
	return errors.Is(e.err, err)
}

type ValidationErrors []validationError

func (e ValidationErrors) Error() string {
	if len(e) == 0 {
		return ""
	}

	buf := bytes.NewBufferString("ValidationErrors: ")
	buf.Grow(len(e) * 16) //nolint:gomnd // just because
	for i := range e {
		buf.WriteString(e[i].Error())
		if i != len(e)-1 {
			buf.WriteString("; ")
		}
	}

	return buf.String()
}

func (e ValidationErrors) Errors() []validationError { //nolint:revive
	errs := make([]validationError, len(e))
	copy(errs, e)

	return errs
}

func (e *ValidationErrors) Add(err *validationError) {
	if err != nil {
		*e = append(*e, *err)
	}
}

func (e ValidationErrors) AsError() error {
	if len(e) == 0 {
		return nil
	}

	return e
}

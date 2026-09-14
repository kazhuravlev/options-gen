package errors_test

import (
	"fmt"
	"io"
	"testing"

	"github.com/kazhuravlev/options-gen/pkg/errors"
)

var (
	errBenchValidationSink  error
	errBenchValidationsSink error
)

// BenchmarkNewValidationError benchmarks creating a single validation error.
func BenchmarkNewValidationError(b *testing.B) {
	err := io.EOF

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		errBenchValidationSink = errors.NewValidationError("fieldName", err)
	}
}

// BenchmarkNewValidationErrorNil benchmarks creating a validation error with nil error.
func BenchmarkNewValidationErrorNil(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		errBenchValidationSink = errors.NewValidationError("fieldName", nil)
	}
}

// BenchmarkValidationErrorError benchmarks error string generation.
func BenchmarkValidationErrorError(b *testing.B) {
	err := errors.NewValidationError("fieldName", io.EOF)

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		_ = err.Error()
	}
}

// BenchmarkValidationErrorIs benchmarks error checking.
func BenchmarkValidationErrorIs(b *testing.B) {
	err := errors.NewValidationError("fieldName", io.EOF)

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		_ = err.Is(io.EOF)
	}
}

// BenchmarkValidationErrorsAdd benchmarks adding errors to collection.
func BenchmarkValidationErrorsAdd(b *testing.B) {
	b.Run("small_collection", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()

		for b.Loop() {
			errs := new(errors.ValidationErrors)
			for i := range 5 {
				errs.Add(errors.NewValidationError(
					fmt.Sprintf("field%d", i),
					io.EOF,
				))
			}
		}
	})

	b.Run("medium_collection", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()

		for b.Loop() {
			errs := new(errors.ValidationErrors)
			for i := range 50 {
				errs.Add(errors.NewValidationError(
					fmt.Sprintf("field%d", i),
					io.EOF,
				))
			}
		}
	})

	b.Run("large_collection", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()

		for b.Loop() {
			errs := new(errors.ValidationErrors)
			for i := range 500 {
				errs.Add(errors.NewValidationError(
					fmt.Sprintf("field%d", i),
					io.EOF,
				))
			}
		}
	})
}

// BenchmarkValidationErrorsError benchmarks generating error string from collection.
func BenchmarkValidationErrorsError(b *testing.B) {
	errs := new(errors.ValidationErrors)
	for i := range 10 {
		errs.Add(errors.NewValidationError(
			fmt.Sprintf("field%d", i),
			io.EOF,
		))
	}

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		_ = errs.Error()
	}
}

// BenchmarkValidationErrorsAsError benchmarks converting to error interface.
func BenchmarkValidationErrorsAsError(b *testing.B) {
	errs := new(errors.ValidationErrors)
	for i := range 10 {
		errs.Add(errors.NewValidationError(
			fmt.Sprintf("field%d", i),
			io.EOF,
		))
	}

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		_ = errs.AsError()
	}
}

// BenchmarkValidationErrorsErrors benchmarks getting error list copy.
func BenchmarkValidationErrorsErrors(b *testing.B) {
	errs := new(errors.ValidationErrors)
	for i := range 20 {
		errs.Add(errors.NewValidationError(
			fmt.Sprintf("field%d", i),
			io.EOF,
		))
	}

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		errBenchValidationsSink = errs.AsError()
	}
}

// BenchmarkValidationErrorsEmptyError benchmarks empty collection error string.
func BenchmarkValidationErrorsEmptyError(b *testing.B) {
	errs := new(errors.ValidationErrors)

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		_ = errs.Error()
	}
}

// BenchmarkValidationErrorsEmptyAsError benchmarks empty collection as error.
func BenchmarkValidationErrorsEmptyAsError(b *testing.B) {
	errs := new(errors.ValidationErrors)

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		errBenchValidationsSink = errs.AsError()
	}
}

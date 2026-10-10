package errors_test

import (
	"io"
	"strconv"
	"testing"

	"github.com/kazhuravlev/options-gen/pkg/errors"
)

var (
	errBenchValidationSink  error
	errBenchValidationsSink error
	benchErrorStringSink    string
	benchIsSink             bool
	benchErrorsCopySink     int
	benchCollectionLenSink  int
)

// benchFieldNames returns n distinct field names, built once outside the timed loop.
func benchFieldNames(n int) []string {
	names := make([]string, n)
	for i := range names {
		names[i] = "field" + strconv.Itoa(i)
	}

	return names
}

// benchCollection returns a collection with n validation errors.
func benchCollection(n int) *errors.ValidationErrors {
	errs := new(errors.ValidationErrors)
	for _, name := range benchFieldNames(n) {
		errs.Add(errors.NewValidationError(name, io.EOF))
	}

	return errs
}

// BenchmarkNewValidationError benchmarks creating a single validation error.
func BenchmarkNewValidationError(b *testing.B) {
	err := io.EOF

	b.ReportAllocs()

	for b.Loop() {
		errBenchValidationSink = errors.NewValidationError("fieldName", err)
	}
}

// BenchmarkNewValidationErrorNil benchmarks the nil fast path of NewValidationError.
func BenchmarkNewValidationErrorNil(b *testing.B) {
	b.ReportAllocs()

	for b.Loop() {
		errBenchValidationSink = errors.NewValidationError("fieldName", nil)
	}
}

// BenchmarkValidationErrorError benchmarks error string generation.
func BenchmarkValidationErrorError(b *testing.B) {
	err := errors.NewValidationError("fieldName", io.EOF)

	b.ReportAllocs()

	for b.Loop() {
		benchErrorStringSink = err.Error()
	}
}

// BenchmarkValidationErrorIs benchmarks error matching through errors.Is.
func BenchmarkValidationErrorIs(b *testing.B) {
	err := errors.NewValidationError("fieldName", io.EOF)

	b.ReportAllocs()

	for b.Loop() {
		benchIsSink = err.Is(io.EOF)
	}
}

// BenchmarkValidationErrorsAdd benchmarks building a collection of 5/50/500 errors
// from scratch (the field names are prepared outside the loop so only
// NewValidationError and Add are measured).
func BenchmarkValidationErrorsAdd(b *testing.B) {
	for _, tt := range []struct {
		name string
		size int
	}{
		{name: "small_collection", size: 5},
		{name: "medium_collection", size: 50},
		{name: "large_collection", size: 500},
	} {
		b.Run(tt.name, func(b *testing.B) {
			names := benchFieldNames(tt.size)

			b.ReportAllocs()

			for b.Loop() {
				errs := new(errors.ValidationErrors)
				for _, name := range names {
					errs.Add(errors.NewValidationError(name, io.EOF))
				}

				benchCollectionLenSink = len(*errs)
			}
		})
	}
}

// BenchmarkValidationErrorsError benchmarks generating the error string of a
// collection of ten errors.
func BenchmarkValidationErrorsError(b *testing.B) {
	errs := benchCollection(10)

	b.ReportAllocs()

	for b.Loop() {
		benchErrorStringSink = errs.Error()
	}
}

// BenchmarkValidationErrorsAsError benchmarks converting a non-empty collection to
// the error interface.
func BenchmarkValidationErrorsAsError(b *testing.B) {
	errs := benchCollection(10)

	b.ReportAllocs()

	for b.Loop() {
		errBenchValidationsSink = errs.AsError()
	}
}

// BenchmarkValidationErrorsErrors benchmarks copying the error list of a collection
// of twenty errors through Errors().
func BenchmarkValidationErrorsErrors(b *testing.B) {
	errs := benchCollection(20)

	b.ReportAllocs()

	for b.Loop() {
		benchErrorsCopySink = len(errs.Errors())
	}
}

// BenchmarkValidationErrorsEmptyError benchmarks the empty collection error string.
func BenchmarkValidationErrorsEmptyError(b *testing.B) {
	errs := new(errors.ValidationErrors)

	b.ReportAllocs()

	for b.Loop() {
		benchErrorStringSink = errs.Error()
	}
}

// BenchmarkValidationErrorsEmptyAsError benchmarks the empty collection as error.
func BenchmarkValidationErrorsEmptyAsError(b *testing.B) {
	errs := new(errors.ValidationErrors)

	b.ReportAllocs()

	for b.Loop() {
		errBenchValidationsSink = errs.AsError()
	}
}

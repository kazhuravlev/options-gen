package optionsgen_test

import (
	"testing"
	"time"

	testcase "github.com/kazhuravlev/options-gen/options-gen/testdata/case-12-defaults-tag-02"
	"github.com/stretchr/testify/assert"
)

func TestDefaultValues(t *testing.T) {
	cases := []struct {
		opts      testcase.Options
		wantError bool
	}{
		{
			opts:      testcase.NewOptions(),
			wantError: false,
		},
		{
			opts:      testcase.NewOptions(testcase.WithSomePingPeriod(0)),
			wantError: true,
		},
		{
			opts:      testcase.NewOptions(testcase.WithSomePingPeriod(time.Hour)),
			wantError: true,
		},
		{
			opts:      testcase.NewOptions(testcase.WithSomeName("")),
			wantError: true,
		},
		{
			opts:      testcase.NewOptions(testcase.WithSomeMaxAttempts(0)),
			wantError: true,
		},
		{
			opts:      testcase.NewOptions(testcase.WithSomeMaxAttempts(-1)),
			wantError: true,
		},
		{
			opts:      testcase.NewOptions(testcase.WithSomeMaxAttempts(11)),
			wantError: true,
		},
		{
			opts:      testcase.NewOptions(testcase.WithSomeEps(0.)),
			wantError: true,
		},
	}

	for _, tt := range cases {
		t.Run("", func(t *testing.T) {
			err := tt.opts.Validate()
			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

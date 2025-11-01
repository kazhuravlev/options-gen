package main

import (
	"testing"

	optionsgen "github.com/kazhuravlev/options-gen/options-gen"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_parseDefaults(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    *optionsgen.Defaults
		wantErr bool
	}{
		{
			name:  "none",
			input: "none",
			want: &optionsgen.Defaults{
				From:  optionsgen.DefaultsFromNone,
				Param: "",
			},
			wantErr: false,
		},
		{
			name:  "tag with parameter",
			input: "tag=default",
			want: &optionsgen.Defaults{
				From:  optionsgen.DefaultsFromTag,
				Param: "default",
			},
			wantErr: false,
		},
		{
			name:  "tag without parameter",
			input: "tag",
			want: &optionsgen.Defaults{
				From:  optionsgen.DefaultsFromTag,
				Param: "",
			},
			wantErr: false,
		},
		{
			name:  "tag with custom name",
			input: "tag=custom",
			want: &optionsgen.Defaults{
				From:  optionsgen.DefaultsFromTag,
				Param: "custom",
			},
			wantErr: false,
		},
		{
			name:  "var with parameter",
			input: "var=defaultOptions",
			want: &optionsgen.Defaults{
				From:  optionsgen.DefaultsFromVar,
				Param: "defaultOptions",
			},
			wantErr: false,
		},
		{
			name:  "var without parameter",
			input: "var",
			want: &optionsgen.Defaults{
				From:  optionsgen.DefaultsFromVar,
				Param: "",
			},
			wantErr: false,
		},
		{
			name:  "func with parameter",
			input: "func=getDefaults",
			want: &optionsgen.Defaults{
				From:  optionsgen.DefaultsFromFunc,
				Param: "getDefaults",
			},
			wantErr: false,
		},
		{
			name:  "func without parameter",
			input: "func",
			want: &optionsgen.Defaults{
				From:  optionsgen.DefaultsFromFunc,
				Param: "",
			},
			wantErr: false,
		},
		{
			name:    "invalid source",
			input:   "invalid",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "empty string",
			input:   "",
			want:    nil,
			wantErr: true,
		},
		{
			name:  "tag with equals in value",
			input: "tag=some=value",
			want: &optionsgen.Defaults{
				From:  optionsgen.DefaultsFromTag,
				Param: "",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := parseDefaults(tt.input)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want.From, got.From)
				assert.Equal(t, tt.want.Param, got.Param)
			}
		})
	}
}

func Test_get1(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input []string
		want  string
	}{
		{
			name:  "two elements",
			input: []string{"first", "second"},
			want:  "second",
		},
		{
			name:  "one element",
			input: []string{"first"},
			want:  "",
		},
		{
			name:  "empty slice",
			input: []string{},
			want:  "",
		},
		{
			name:  "three elements",
			input: []string{"first", "second", "third"},
			want:  "",
		},
		{
			name:  "two elements with empty second",
			input: []string{"first", ""},
			want:  "",
		},
		{
			name:  "two elements with empty first",
			input: []string{"", "second"},
			want:  "second",
		},
		{
			name:  "nil slice",
			input: nil,
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := get1(tt.input)
			require.Equal(t, tt.want, got)
		})
	}
}

func Test_isEmpty(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		values []string
		want   bool
	}{
		{
			name:   "all non-empty",
			values: []string{"first", "second", "third"},
			want:   false,
		},
		{
			name:   "one empty at start",
			values: []string{"", "second", "third"},
			want:   true,
		},
		{
			name:   "one empty in middle",
			values: []string{"first", "", "third"},
			want:   true,
		},
		{
			name:   "one empty at end",
			values: []string{"first", "second", ""},
			want:   true,
		},
		{
			name:   "all empty",
			values: []string{"", "", ""},
			want:   true,
		},
		{
			name:   "single non-empty",
			values: []string{"value"},
			want:   false,
		},
		{
			name:   "single empty",
			values: []string{""},
			want:   true,
		},
		{
			name:   "empty slice",
			values: []string{},
			want:   false,
		},
		{
			name:   "nil slice",
			values: nil,
			want:   false,
		},
		{
			name:   "whitespace is not empty",
			values: []string{"  ", "value"},
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := isEmpty(tt.values...)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_splitExcludes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    []string // pattern strings for comparison
		wantErr bool
	}{
		{
			name:    "empty string",
			input:   "",
			want:    nil,
			wantErr: false,
		},
		{
			name:    "single pattern",
			input:   "^test.*",
			want:    []string{"^test.*"},
			wantErr: false,
		},
		{
			name:    "multiple patterns",
			input:   "^test.*;^debug.*;^internal.*",
			want:    []string{"^test.*", "^debug.*", "^internal.*"},
			wantErr: false,
		},
		{
			name:    "pattern with special chars",
			input:   `^\w+_test$`,
			want:    []string{`^\w+_test$`},
			wantErr: false,
		},
		{
			name:    "invalid regex pattern",
			input:   "[invalid",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "mixed valid and invalid",
			input:   "^valid.*;[invalid",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "simple word pattern",
			input:   "test",
			want:    []string{"test"},
			wantErr: false,
		},
		{
			name:    "multiple simple patterns",
			input:   "foo;bar;baz",
			want:    []string{"foo", "bar", "baz"},
			wantErr: false,
		},
		{
			name:    "pattern with dots and stars",
			input:   ".*_internal.*;.*_private.*",
			want:    []string{".*_internal.*", ".*_private.*"},
			wantErr: false,
		},
		{
			name:    "pattern matching any",
			input:   ".*",
			want:    []string{".*"},
			wantErr: false,
		},
		{
			name:    "pattern with alternation",
			input:   "^(foo|bar)$",
			want:    []string{"^(foo|bar)$"},
			wantErr: false,
		},
		{
			name:    "empty pattern between semicolons",
			input:   "foo;;bar",
			want:    []string{"foo", "", "bar"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := splitExcludes(tt.input)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Len(t, got, len(tt.want))

				for i, pattern := range tt.want {
					assert.Equal(t, pattern, got[i].String())
				}
			}
		})
	}
}

func Test_splitExcludes_Matching(t *testing.T) {
	t.Parallel()

	t.Run("patterns match correctly", func(t *testing.T) {
		t.Parallel()

		patterns, err := splitExcludes("^test.*;.*_internal$")
		require.NoError(t, err)

		testCases := []struct {
			field       string
			shouldMatch bool
		}{
			{"testField", true},
			{"test_something", true},
			{"field_internal", true},
			{"normalField", false},
			{"internal_field", false},
		}

		for _, testCase := range testCases {
			matched := false
			for _, pattern := range patterns {
				if pattern.MatchString(testCase.field) {
					matched = true

					break
				}
			}

			assert.Equal(t, testCase.shouldMatch, matched, "field %q matching", testCase.field)
		}
	})

	t.Run("compiled regex is usable", func(t *testing.T) {
		t.Parallel()

		patterns, err := splitExcludes(`^\d+$`)
		require.NoError(t, err)
		require.Len(t, patterns, 1)

		assert.True(t, patterns[0].MatchString("123"))
		assert.False(t, patterns[0].MatchString("abc"))
	})
}

func Test_splitExcludes_ErrorMessages(t *testing.T) {
	t.Parallel()

	t.Run("error contains pattern info", func(t *testing.T) {
		t.Parallel()

		_, err := splitExcludes("[invalid")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "[invalid")
	})

	t.Run("error contains compile info", func(t *testing.T) {
		t.Parallel()

		_, err := splitExcludes("(unclosed")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "compile")
	})
}

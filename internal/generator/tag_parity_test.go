//nolint:exhaustruct
package generator //nolint:testpackage

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

// tagParitySeeds are struct tags (and the configured defaults tag name) that
// exercise the corner cases of the tag grammar: malformed keys, unterminated
// values, escapes, duplicates and keys that collide with the configured name.
var tagParitySeeds = []struct {
	tag     string
	tagName string
}{
	{tag: `validate:"required" default:"42" option:"mandatory"`, tagName: "default"},
	{tag: `option:"variadic=true,name=Custom" json:"x,omitempty"`, tagName: "default"},
	{tag: `my-default-tag:"3s" validate:"min=100ms"`, tagName: "my-default-tag"},
	{tag: `validate:"required"`, tagName: "validate"},
	{tag: `option:"mandatory"`, tagName: "option"},
	{tag: `default:"42"`, tagName: ""},
	{tag: `default:"first" default:"second"`, tagName: "default"},
	{tag: `validate:"\q" validate:"required"`, tagName: "default"},
	{tag: `option:"a\"b\\c"`, tagName: "default"},
	{tag: `validate:"unterminated`, tagName: "default"},
	{tag: `novalue option:"mandatory"`, tagName: "default"},
	{tag: `validate "required"`, tagName: "default"},
	{tag: `validate:required`, tagName: "default"},
	{tag: `   validate:"required"   `, tagName: "default"},
	{tag: "option:\"mandatory\"\tvalidate:\"required\"", tagName: "default"},
	{tag: "opt\x7fion:\"mandatory\"", tagName: "default"},
	{tag: "option:\"\n\"", tagName: "default"},
	{tag: `:"x" option:"mandatory"`, tagName: "default"},
	{tag: `option:""`, tagName: "default"},
	{tag: ``, tagName: "default"},
}

// FuzzLookupTagValues checks that the hand-written struct tag scanner agrees
// with reflect.StructTag.Lookup for every key it extracts, including
// malformed tags.
func FuzzLookupTagValues(f *testing.F) {
	for _, seed := range tagParitySeeds {
		f.Add(seed.tag, seed.tagName)
	}

	f.Fuzz(func(t *testing.T, tag, tagName string) {
		validate, defaultValue, option := lookupTagValues(tag, tagName)

		structTag := reflect.StructTag(tag)
		wantValidate, _ := structTag.Lookup("validate")
		wantDefault, _ := structTag.Lookup(tagName)
		wantOption, _ := structTag.Lookup("option")

		require.Equal(t, wantValidate, validate, "validate key, tag %q", tag)
		require.Equal(t, wantDefault, defaultValue, "%q key, tag %q", tagName, tag)
		require.Equal(t, wantOption, option, "option key, tag %q", tag)
	})
}

func Test_lookupTagValues_CornerCases(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name         string
		tag          string
		tagName      string
		wantValidate string
		wantDefault  string
		wantOption   string
	}{
		{
			name:       "unterminated_value_stops_scanning",
			tag:        `option:"mandatory" validate:"required`,
			tagName:    "default",
			wantOption: "mandatory",
		},
		{
			name:    "key_without_colon_stops_scanning",
			tag:     `json validate:"required"`,
			tagName: "default",
		},
		{
			name:    "key_without_quote_stops_scanning",
			tag:     `validate:required option:"mandatory"`,
			tagName: "default",
		},
		{
			name:        "escaped_quote_inside_value",
			tag:         `default:"a\"b" option:"mandatory"`,
			tagName:     "default",
			wantDefault: `a"b`,
			wantOption:  "mandatory",
		},
		{
			name:        "escaped_backslash",
			tag:         `default:"a\\b"`,
			tagName:     "default",
			wantDefault: `a\b`,
		},
		{
			name:        "unicode_escape",
			tag:         `default:"café"`,
			tagName:     "default",
			wantDefault: "café",
		},
		{
			name:       "invalid_escape_resolves_to_empty_value",
			tag:        `default:"\q" option:"mandatory"`,
			tagName:    "default",
			wantOption: "mandatory",
		},
		{
			name:    "control_character_in_key",
			tag:     "op\x01tion:\"mandatory\"",
			tagName: "default",
		},
		{
			name:        "duplicate_key_first_wins",
			tag:         `default:"1" default:"2"`,
			tagName:     "default",
			wantDefault: "1",
		},
		{
			name:    "duplicate_key_malformed_first_wins",
			tag:     `default:"\q" default:"2"`,
			tagName: "default",
		},
		{
			name:         "tag_name_equal_to_validate",
			tag:          `validate:"required"`,
			tagName:      "validate",
			wantValidate: "required",
			wantDefault:  "required",
		},
		{
			name:        "tag_name_equal_to_option",
			tag:         `option:"mandatory"`,
			tagName:     "option",
			wantDefault: "mandatory",
			wantOption:  "mandatory",
		},
		{
			name:    "empty_tag_name_never_matches",
			tag:     `default:"42"`,
			tagName: "",
		},
		{
			name:       "unknown_keys_are_skipped",
			tag:        `json:"x" yaml:"y" option:"mandatory"`,
			tagName:    "default",
			wantOption: "mandatory",
		},
		{
			name:       "trailing_spaces",
			tag:        `option:"mandatory"   `,
			tagName:    "default",
			wantOption: "mandatory",
		},
		{
			name:       "tab_separator_is_a_syntax_error",
			tag:        "option:\"mandatory\"\tvalidate:\"required\"",
			tagName:    "default",
			wantOption: "mandatory",
		},
		{
			name:        "colon_inside_value",
			tag:         `default:"127.0.0.1:8000"`,
			tagName:     "default",
			wantDefault: "127.0.0.1:8000",
		},
		{
			name:    "empty_value",
			tag:     `option:""`,
			tagName: "default",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			validate, defaultValue, option := lookupTagValues(testCase.tag, testCase.tagName)
			require.Equal(t, testCase.wantValidate, validate)
			require.Equal(t, testCase.wantDefault, defaultValue)
			require.Equal(t, testCase.wantOption, option)

			// The expectations above must match the reference implementation.
			structTag := reflect.StructTag(testCase.tag)
			refValidate, _ := structTag.Lookup("validate")
			refDefault, _ := structTag.Lookup(testCase.tagName)
			refOption, _ := structTag.Lookup("option")
			require.Equal(t, refValidate, validate)
			require.Equal(t, refDefault, defaultValue)
			require.Equal(t, refOption, option)
		})
	}
}

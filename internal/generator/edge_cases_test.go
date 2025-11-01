//nolint:testpackage
package generator

import (
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestRender_EdgeCases tests template rendering with edge cases that might expose bugs.
func TestRender_EdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		opts    Options
		wantErr bool
		errMsg  string
	}{
		{
			name: "empty package name",
			opts: NewOptions(
				WithVersion("test"),
				WithPackageName(""),
				WithSpec(&OptionSpec{
					TypeParamsSpec: "",
					TypeParams:     "",
					Options:        []OptionMeta{},
				}),
			),
			wantErr: true, // Empty package name fails validation
			errMsg:  "",
		},
		{
			name: "very long option name",
			opts: NewOptions(
				WithVersion("test"),
				WithPackageName("test"),
				WithOptionsStructName("Options"),
				WithOptionTypeName("Option"),
				WithTagName("default"),
				WithConstructorTypeRender("public"),
				WithSpec(&OptionSpec{
					TypeParamsSpec: "",
					TypeParams:     "",
					Options: []OptionMeta{
						{
							Name:      strings.Repeat("A", 1000),
							Docstring: "",
							Field:     strings.Repeat("a", 1000),
							Type:      "string",
							TagOption: TagOption{
								IsRequired:    false,
								GoValidator:   "",
								Default:       "",
								Variadic:      false,
								VariadicIsSet: false,
								Skip:          false,
							},
						},
					},
				}),
			),
			wantErr: false,
			errMsg:  "",
		},
		{
			name: "special characters in type",
			opts: NewOptions(
				WithVersion("test"),
				WithPackageName("test"),
				WithOptionsStructName("Options"),
				WithOptionTypeName("Option"),
				WithTagName("default"),
				WithConstructorTypeRender("public"),
				WithSpec(&OptionSpec{
					TypeParamsSpec: "",
					TypeParams:     "",
					Options: []OptionMeta{
						{
							Name:      "Field",
							Docstring: "",
							Field:     "field",
							Type:      "map[string][]interface{}",
							TagOption: TagOption{
								IsRequired:    false,
								GoValidator:   "",
								Default:       "",
								Variadic:      false,
								VariadicIsSet: false,
								Skip:          false,
							},
						},
					},
				}),
			),
			wantErr: false,
			errMsg:  "",
		},
		{
			name: "unicode in field names",
			opts: NewOptions(
				WithVersion("test"),
				WithPackageName("test"),
				WithOptionsStructName("Options"),
				WithOptionTypeName("Option"),
				WithTagName("default"),
				WithConstructorTypeRender("public"),
				WithSpec(&OptionSpec{
					TypeParamsSpec: "",
					TypeParams:     "",
					Options: []OptionMeta{
						{
							Name:      "Field世界",
							Docstring: "",
							Field:     "field世界",
							Type:      "string",
							TagOption: TagOption{
								IsRequired:    false,
								GoValidator:   "",
								Default:       "",
								Variadic:      false,
								VariadicIsSet: false,
								Skip:          false,
							},
						},
					},
				}),
			),
			wantErr: false,
			errMsg:  "",
		},
		{
			name: "deeply nested generic types",
			opts: NewOptions(
				WithVersion("test"),
				WithPackageName("test"),
				WithOptionsStructName("Options"),
				WithOptionTypeName("Option"),
				WithTagName("default"),
				WithConstructorTypeRender("public"),
				WithSpec(&OptionSpec{
					TypeParamsSpec: "[T any, U comparable, V interface{ Method() string }]",
					TypeParams:     "[T, U, V]",
					Options: []OptionMeta{
						{
							Name:      "NestedGeneric",
							Docstring: "",
							Field:     "nestedGeneric",
							Type:      "map[T][]map[U]V",
							TagOption: TagOption{
								IsRequired:    false,
								GoValidator:   "",
								Default:       "",
								Variadic:      false,
								VariadicIsSet: false,
								Skip:          false,
							},
						},
					},
				}),
			),
			wantErr: false,
			errMsg:  "",
		},
		{
			name: "all options mandatory",
			opts: NewOptions(
				WithVersion("test"),
				WithPackageName("test"),
				WithOptionsStructName("Options"),
				WithOptionTypeName("Option"),
				WithTagName("default"),
				WithConstructorTypeRender("public"),
				WithSpec(&OptionSpec{
					TypeParamsSpec: "",
					TypeParams:     "",
					Options: []OptionMeta{
						{
							Name:      "Field1",
							Docstring: "",
							Field:     "field1",
							Type:      "string",
							TagOption: TagOption{
								IsRequired:    true,
								GoValidator:   "",
								Default:       "",
								Variadic:      false,
								VariadicIsSet: false,
								Skip:          false,
							},
						},
						{
							Name:      "Field2",
							Docstring: "",
							Field:     "field2",
							Type:      "int",
							TagOption: TagOption{
								IsRequired:    true,
								GoValidator:   "",
								Default:       "",
								Variadic:      false,
								VariadicIsSet: false,
								Skip:          false,
							},
						},
					},
				}),
			),
			wantErr: false,
			errMsg:  "",
		},
		{
			name: "option with validation and default",
			opts: NewOptions(
				WithVersion("test"),
				WithPackageName("test"),
				WithOptionsStructName("Options"),
				WithOptionTypeName("Option"),
				WithTagName("default"),
				WithConstructorTypeRender("public"),
				WithSpec(&OptionSpec{
					TypeParamsSpec: "",
					TypeParams:     "",
					Options: []OptionMeta{
						{
							Name:      "Email",
							Docstring: "",
							Field:     "email",
							Type:      "string",
							TagOption: TagOption{
								IsRequired:    false,
								GoValidator:   "email,required",
								Default:       "test@example.com",
								Variadic:      false,
								VariadicIsSet: false,
								Skip:          false,
							},
						},
					},
				}),
			),
			wantErr: false,
			errMsg:  "",
		},
		{
			name: "variadic slice option",
			opts: NewOptions(
				WithVersion("test"),
				WithPackageName("test"),
				WithOptionsStructName("Options"),
				WithOptionTypeName("Option"),
				WithTagName("default"),
				WithConstructorTypeRender("public"),
				WithSpec(&OptionSpec{
					TypeParamsSpec: "",
					TypeParams:     "",
					Options: []OptionMeta{
						{
							Name:      "Items",
							Docstring: "",
							Field:     "items",
							Type:      "string",
							TagOption: TagOption{
								IsRequired:    false,
								GoValidator:   "",
								Default:       "",
								Variadic:      true,
								VariadicIsSet: false,
								Skip:          false,
							},
						},
					},
				}),
			),
			wantErr: false,
			errMsg:  "",
		},
		{
			name: "zero options",
			opts: NewOptions(
				WithVersion("test"),
				WithPackageName("test"),
				WithOptionsStructName("Options"),
				WithOptionTypeName("Option"),
				WithTagName("default"),
				WithConstructorTypeRender("public"),
				WithSpec(&OptionSpec{
					TypeParamsSpec: "",
					TypeParams:     "",
					Options:        []OptionMeta{},
				}),
			),
			wantErr: false,
			errMsg:  "",
		},
		{
			name: "with isset enabled",
			opts: NewOptions(
				WithVersion("test"),
				WithPackageName("test"),
				WithOptionsStructName("Options"),
				WithOptionTypeName("Option"),
				WithTagName("default"),
				WithConstructorTypeRender("public"),
				WithWithIsset(true),
				WithSpec(&OptionSpec{
					TypeParamsSpec: "",
					TypeParams:     "",
					Options: []OptionMeta{
						{
							Name:      "Field",
							Docstring: "",
							Field:     "field",
							Type:      "string",
							TagOption: TagOption{
								IsRequired:    false,
								GoValidator:   "",
								Default:       "",
								Variadic:      false,
								VariadicIsSet: false,
								Skip:          false,
							},
						},
					},
				}),
			),
			wantErr: false,
			errMsg:  "",
		},
		{
			name: "private constructor",
			opts: NewOptions(
				WithVersion("test"),
				WithPackageName("test"),
				WithOptionsStructName("Options"),
				WithOptionTypeName("Option"),
				WithTagName("default"),
				WithConstructorTypeRender("private"),
				WithSpec(&OptionSpec{
					TypeParamsSpec: "",
					TypeParams:     "",
					Options: []OptionMeta{
						{
							Name:      "Field",
							Docstring: "",
							Field:     "field",
							Type:      "string",
							TagOption: TagOption{
								IsRequired:    false,
								GoValidator:   "",
								Default:       "",
								Variadic:      false,
								VariadicIsSet: false,
								Skip:          false,
							},
						},
					},
				}),
			),
			wantErr: false,
			errMsg:  "",
		},
		{
			name: "no constructor",
			opts: NewOptions(
				WithVersion("test"),
				WithPackageName("test"),
				WithOptionsStructName("Options"),
				WithSpec(&OptionSpec{
					TypeParamsSpec: "",
					TypeParams:     "",
					Options: []OptionMeta{
						{
							Name:      "Field",
							Docstring: "",
							Field:     "field",
							Type:      "string",
							TagOption: TagOption{
								IsRequired:    false,
								GoValidator:   "",
								Default:       "",
								Variadic:      false,
								VariadicIsSet: false,
								Skip:          false,
							},
						},
					},
				}),
				WithOptionTypeName("Option"),
				WithConstructorTypeRender("no"),
			),
			wantErr: false,
			errMsg:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Render(tt.opts)
			if tt.wantErr {
				require.Error(t, err)
				require.ErrorContains(t, err, tt.errMsg)
			} else {
				require.NoError(t, err)
				require.NotEmpty(t, result)
			}
		})
	}
}

// TestApplyExcludes_EdgeCases tests field exclusion logic with edge cases.
func TestApplyExcludes_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		options  []OptionMeta
		excludes []*regexp.Regexp
		want     int // expected number of options after exclusion
	}{
		{
			name:     "nil excludes",
			options:  []OptionMeta{{Name: "Field1"}, {Name: "Field2"}}, //nolint:exhaustruct
			excludes: nil,
			want:     2,
		},
		{
			name:     "empty excludes",
			options:  []OptionMeta{{Name: "Field1"}, {Name: "Field2"}}, //nolint:exhaustruct
			excludes: []*regexp.Regexp{},
			want:     2,
		},
		{
			name:    "exclude all",
			options: []OptionMeta{{Name: "Field1"}, {Name: "Field2"}}, //nolint:exhaustruct
			excludes: []*regexp.Regexp{
				regexp.MustCompile(".*"),
			},
			want: 0,
		},
		{
			name:    "exclude none",
			options: []OptionMeta{{Name: "Field1"}, {Name: "Field2"}}, //nolint:exhaustruct
			excludes: []*regexp.Regexp{
				regexp.MustCompile("NonExistent"),
			},
			want: 2,
		},
		{
			name:    "multiple patterns",
			options: []OptionMeta{{Name: "FieldA"}, {Name: "FieldB"}, {Name: "OtherC"}}, //nolint:exhaustruct
			excludes: []*regexp.Regexp{
				regexp.MustCompile("^Field"),
				regexp.MustCompile("C$"),
			},
			want: 0,
		},
		{
			name:    "case sensitive exclusion",
			options: []OptionMeta{{Name: "field"}, {Name: "Field"}}, //nolint:exhaustruct
			excludes: []*regexp.Regexp{
				regexp.MustCompile("^field$"),
			},
			want: 1,
		},
		{
			name:     "empty options",
			options:  []OptionMeta{},
			excludes: []*regexp.Regexp{regexp.MustCompile(".*")},
			want:     0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ApplyExcludes(tt.options, tt.excludes)
			require.Equal(t, tt.want, len(result))
		})
	}
}

// TestCheckDefaultValue_AllTypes tests all supported default value types.
func TestCheckDefaultValue_AllTypes(t *testing.T) {
	tests := []struct {
		name      string
		fieldType string
		value     string
		wantErr   bool
	}{
		// Integer types - valid
		{"int valid", "int", "42", false},
		{"int negative", "int", "-42", false},
		{"int zero", "int", "0", false},
		{"int8 valid", "int8", "127", false},
		{"int16 valid", "int16", "32767", false},
		{"int32 valid", "int32", "2147483647", false},
		{"int64 valid", "int64", "9223372036854775807", false},

		// Unsigned integer types - valid
		{"uint valid", "uint", "42", false},
		{"uint zero", "uint", "0", false},
		{"uint8 valid", "uint8", "255", false},
		{"uint16 valid", "uint16", "65535", false},
		{"uint32 valid", "uint32", "4294967295", false},
		{"uint64 valid", "uint64", "18446744073709551615", false},

		// Float types - valid
		{"float32 valid", "float32", "3.14", false},
		{"float32 negative", "float32", "-3.14", false},
		{"float64 valid", "float64", "3.141592653589793", false},
		{"float64 scientific", "float64", "1.23e-4", false},

		// Duration - valid
		{"duration seconds", "time.Duration", "5s", false},
		{"duration minutes", "time.Duration", "10m", false},
		{"duration hours", "time.Duration", "2h", false},
		{"duration mixed", "time.Duration", "1h30m", false},
		{"duration nanoseconds", "time.Duration", "500ns", false},

		// Bool - valid
		{"bool true", "bool", "true", false},
		{"bool false", "bool", "false", false},

		// String - always valid
		{"string empty", "string", "", false},
		{"string normal", "string", "hello", false},
		{"string special", "string", "!@#$%", false},

		// Integer types - invalid
		{"int invalid", "int", "not a number", true},
		{"int float", "int", "3.14", true},
		{"int overflow", "int64", "99999999999999999999999", true},
		{"uint negative", "uint", "-1", true},

		// Float types - invalid
		{"float invalid", "float32", "not a float", true},

		// Duration - invalid
		{"duration invalid", "time.Duration", "5 seconds", true},
		{"duration bad unit", "time.Duration", "5x", true},

		// Bool - invalid
		{"bool yes", "bool", "yes", true},
		{"bool 1", "bool", "1", true},
		{"bool empty", "bool", "", true},

		// Unsupported type
		{"unsupported type", "CustomType", "value", true},
		{"unsupported complex", "complex64", "1+2i", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := checkDefaultValue(tt.fieldType, tt.value)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestIsPublic_EdgeCases tests isPublic with various unicode and edge cases.
func TestIsPublic_EdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		fieldName string
		want      bool
	}{
		{"lowercase ascii", "field", false},
		{"uppercase ascii", "Field", true},
		{"starts with number", "1Field", false},
		{"starts with underscore", "_Field", false},
		{"empty string", "", false},
		{"single lowercase", "a", false},
		{"single uppercase", "A", true},
		{"unicode lowercase", "фield", false},
		{"unicode uppercase", "Фield", true},
		{"chinese character", "字段", false},
		{"greek uppercase", "Δelta", true},
		{"mixed case", "fIELD", false},
		{"all caps", "FIELD", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := isPublic(tt.fieldName)
			require.Equal(t, tt.want, res)
		})
	}
}

// TestNormalizeTypeName_EdgeCases tests type name normalization edge cases.
func TestNormalizeTypeName_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		typeName string
		want     string
	}{
		{"simple type", "string", "string"},
		{"pointer", "*string", "string"},
		{"slice", "[]string", "string"},
		{"pointer slice", "*[]string", "[]string"}, // Only removes one prefix level
		{"slice pointer", "[]*string", "string"},
		{"package type", "pkg.Type", "Type"},
		{"pointer package type", "*pkg.Type", "Type"},
		{"slice package type", "[]pkg.Type", "Type"},
		{"nested package", "github.com/user/pkg.Type", "Type"},
		{"multiple dots", "a.b.c.Type", "Type"},
		{"empty", "", ""},
		{"just pointer", "*", ""},
		{"just slice", "[]", ""},
		{"double pointer", "**Type", "*Type"},    // Only removes one level
		{"double slice", "[][]Type", "[]Type"},   // Only removes one level
		{"triple prefix", "*[]*Type", "[]*Type"}, // Only removes one level
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeTypeName(tt.typeName)
			require.Equal(t, tt.want, got)
		})
	}
}

// TestFormatComment_EdgeCases tests comment formatting edge cases.
func TestFormatComment_EdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		comment string
		want    string
	}{
		{
			name:    "empty comment",
			comment: "",
			want:    "",
		},
		{
			name:    "single line",
			comment: "This is a comment",
			want:    "// This is a comment",
		},
		{
			name:    "multiple lines",
			comment: "Line 1\nLine 2",
			want:    "// Line 1\n// Line 2",
		},
		{
			name:    "trailing newline",
			comment: "Comment\n",
			want:    "// Comment",
		},
		{
			name:    "multiple trailing newlines",
			comment: "Comment\n\n\n",
			want:    "// Comment\n// \n// ",
		},
		{
			name:    "only newlines",
			comment: "\n\n",
			want:    "// \n// ",
		},
		{
			name:    "special characters",
			comment: "Special: !@#$%^&*()",
			want:    "// Special: !@#$%^&*()",
		},
		{
			name:    "unicode",
			comment: "Unicode: 你好 🚀",
			want:    "// Unicode: 你好 🚀",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatComment(tt.comment)
			require.Equal(t, tt.want, got)
		})
	}
}

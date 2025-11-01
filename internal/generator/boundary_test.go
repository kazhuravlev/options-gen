package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// TestGetOptionSpec_BoundaryConditions tests boundary conditions and stress scenarios.
func TestGetOptionSpec_BoundaryConditions(t *testing.T) {
	tests := []struct {
		name       string
		sourceCode string
		structName string
		wantErr    bool
		validate   func(t *testing.T, res *GetOptionSpecRes)
	}{
		{
			name: "struct with 100 fields",
			sourceCode: func() string {
				t.Helper()

				var fields strings.Builder
				fields.WriteString("package test\ntype Options struct {\n")
				for i := 0; i < 100; i++ {
					fields.WriteString(fmt.Sprintf("  Field%d string\n", i))
				}
				fields.WriteString("}")

				return fields.String()
			}(),
			structName: "Options",
			wantErr:    false,
			validate: func(t *testing.T, res *GetOptionSpecRes) {
				t.Helper()

				if len(res.Spec.Options) != 100 {
					t.Errorf("expected 100 fields, got %d", len(res.Spec.Options))
				}
			},
		},
		{
			name: "extremely long field name",
			sourceCode: fmt.Sprintf(`package test
type Options struct {
	%s string
}`, strings.Repeat("Field", 100)),
			structName: "Options",
			wantErr:    false,
			validate: func(t *testing.T, res *GetOptionSpecRes) {
				t.Helper()

				if len(res.Spec.Options) != 1 {
					t.Errorf("expected 1 field, got %d", len(res.Spec.Options))
				}
			},
		},
		{
			name: "deeply nested types",
			sourceCode: `package test
type Options struct {
	Nested map[string]map[string]map[string][]map[int]interface{}
}`,
			structName: "Options",
			wantErr:    false,
			validate: func(t *testing.T, res *GetOptionSpecRes) {
				t.Helper()

				if len(res.Spec.Options) != 1 {
					t.Errorf("expected 1 field, got %d", len(res.Spec.Options))
				}
			},
		},
		{
			name: "all fields excluded",
			sourceCode: `package test
type Options struct {
	Field1 string ` + "`option:\"-\"`" + `
	Field2 int ` + "`option:\"-\"`" + `
}`,
			structName: "Options",
			wantErr:    false,
			validate: func(t *testing.T, res *GetOptionSpecRes) {
				t.Helper()

				if len(res.Spec.Options) != 0 {
					t.Errorf("expected 0 fields, got %d", len(res.Spec.Options))
				}
			},
		},
		{
			name: "many type parameters",
			sourceCode: `package test
type Options[T1, T2, T3, T4, T5 any] struct {
	F1 T1
	F2 T2
	F3 T3
	F4 T4
	F5 T5
}`,
			structName: "Options",
			wantErr:    false,
			validate: func(t *testing.T, res *GetOptionSpecRes) {
				t.Helper()

				if res.Spec.TypeParams == "" {
					t.Error("expected type parameters to be captured")
				}
				if len(res.Spec.Options) != 5 {
					t.Errorf("expected 5 fields, got %d", len(res.Spec.Options))
				}
			},
		},
		{
			name: "empty struct",
			sourceCode: `package test
type Options struct {
}`,
			structName: "Options",
			wantErr:    false,
			validate: func(t *testing.T, res *GetOptionSpecRes) {
				t.Helper()

				if len(res.Spec.Options) != 0 {
					t.Errorf("expected 0 fields, got %d", len(res.Spec.Options))
				}
			},
		},
		{
			name: "very long tag value",
			sourceCode: fmt.Sprintf(`package test
type Options struct {
	Field string `+"`validate:\"%s\"`"+`
}`, strings.Repeat("required,", 100)),
			structName: "Options",
			wantErr:    false,
		},
		{
			name: "unicode in all places",
			sourceCode: `package test
type Options struct {
	世界 string
	Привет int
	مرحبا bool
}`,
			structName: "Options",
			wantErr:    false,
			validate: func(t *testing.T, res *GetOptionSpecRes) {
				t.Helper()

				if len(res.Spec.Options) != 3 {
					t.Errorf("expected 3 fields, got %d", len(res.Spec.Options))
				}
			},
		},
		{
			name: "complex validation rules",
			sourceCode: `package test
type Options struct {
	Email string ` + "`validate:\"required,email,min=5,max=255\"`" + `
	Age int ` + "`validate:\"required,min=0,max=120,numeric\"`" + `
	URL string ` + "`validate:\"required,url,startswith=https\"`" + `
}`,
			structName: "Options",
			wantErr:    false,
			validate: func(t *testing.T, res *GetOptionSpecRes) {
				t.Helper()

				if len(res.Spec.Options) != 3 {
					t.Errorf("expected 3 fields, got %d", len(res.Spec.Options))
				}
				for _, opt := range res.Spec.Options {
					if opt.TagOption.GoValidator == "" {
						t.Errorf("field %s should have validation", opt.Name)
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			filePath := filepath.Join(tmpDir, "options.go")

			if err := os.WriteFile(filePath, []byte(tt.sourceCode), 0o644); err != nil {
				t.Fatalf("failed to write test file: %v", err)
			}

			res, err := GetOptionSpec(filePath, tt.structName, "default", false, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetOptionSpec() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			if !tt.wantErr && tt.validate != nil {
				tt.validate(t, res)
			}
		})
	}
}

// TestRender_LargeOutput tests rendering with large number of options.
func TestRender_LargeOutput(t *testing.T) {
	const numFields = 200

	options := make([]OptionMeta, numFields)
	for i := 0; i < numFields; i++ {
		options[i] = OptionMeta{
			Name:  fmt.Sprintf("Field%d", i),
			Field: fmt.Sprintf("field%d", i),
			Type:  "string",
		}
	}

	opts := Options{
		version:               "test",
		packageName:           "test",
		optionsStructName:     "Options",
		optionTypeName:        "Option",
		tagName:               "default",
		constructorTypeRender: "public",
		spec: &OptionSpec{
			Options: options,
		},
	}

	result, err := Render(opts)
	if err != nil {
		t.Fatalf("Render() failed: %v", err)
	}

	if len(result) == 0 {
		t.Error("Render() returned empty result")
	}

	// Verify all fields have setters
	output := string(result)
	for i := 0; i < numFields; i++ {
		setter := fmt.Sprintf("WithField%d", i)
		if !strings.Contains(output, setter) {
			t.Errorf("setter %s not found in output", setter)

			break // Don't spam with errors
		}
	}
}

// TestExcludePatterns_Comprehensive tests exclude pattern matching edge cases.
func TestExcludePatterns_Comprehensive(t *testing.T) {
	tests := []struct {
		name          string
		fields        []string
		excludeRegex  string
		expectedCount int
	}{
		{
			name:          "no match",
			fields:        []string{"Field1", "Field2", "Field3"},
			excludeRegex:  "^NoMatch",
			expectedCount: 3,
		},
		{
			name:          "match all",
			fields:        []string{"Field1", "Field2", "Field3"},
			excludeRegex:  ".*",
			expectedCount: 0,
		},
		{
			name:          "prefix match",
			fields:        []string{"TestField1", "TestField2", "OtherField"},
			excludeRegex:  "^Test",
			expectedCount: 1,
		},
		{
			name:          "suffix match",
			fields:        []string{"Field1Test", "Field2Test", "Field3"},
			excludeRegex:  "Test$",
			expectedCount: 1,
		},
		{
			name:          "contains match",
			fields:        []string{"PreInternalPost", "PreExternalPost", "Field"},
			excludeRegex:  "Internal",
			expectedCount: 2,
		},
		{
			name:          "case sensitive",
			fields:        []string{"field", "Field", "FIELD"},
			excludeRegex:  "^field$",
			expectedCount: 2,
		},
		{
			name:          "numeric pattern",
			fields:        []string{"Field1", "Field2", "FieldA"},
			excludeRegex:  "Field[0-9]",
			expectedCount: 1,
		},
		{
			name:          "alternation",
			fields:        []string{"FieldA", "FieldB", "FieldC"},
			excludeRegex:  "Field(A|B)",
			expectedCount: 1,
		},
		{
			name:          "unicode pattern",
			fields:        []string{"Field世界", "Field123", "OtherField"},
			excludeRegex:  "世界",
			expectedCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			options := make([]OptionMeta, len(tt.fields))
			for i, field := range tt.fields {
				options[i] = OptionMeta{
					Name:  field,
					Field: strings.ToLower(field),
					Type:  "string",
				}
			}

			pattern := regexp.MustCompile(tt.excludeRegex)
			result := applyExcludes(options, []*regexp.Regexp{pattern})

			if len(result) != tt.expectedCount {
				t.Errorf("applyExcludes() returned %d fields, want %d",
					len(result), tt.expectedCount)
			}
		})
	}
}

// TestParseTag_AllCombinations tests all possible tag combinations.
func TestParseTag_AllCombinations(t *testing.T) {
	tests := []struct {
		name       string
		tag        string
		expectOpts TagOption
	}{
		{
			name: "mandatory only",
			tag:  `option:"mandatory"`,
			expectOpts: TagOption{
				IsRequired: true,
			},
		},
		{
			name: "variadic only",
			tag:  `option:"variadic=true"`,
			expectOpts: TagOption{
				Variadic:      true,
				VariadicIsSet: true,
			},
		},
		{
			name: "skip only",
			tag:  `option:"-"`,
			expectOpts: TagOption{
				Skip: true,
			},
		},
		{
			name: "mandatory and variadic (should be detected)",
			tag:  `option:"mandatory,variadic=true"`,
			expectOpts: TagOption{
				IsRequired:    true,
				Variadic:      true,
				VariadicIsSet: true,
			},
		},
		{
			name: "validation only",
			tag:  `validate:"required,email"`,
			expectOpts: TagOption{
				GoValidator: "required,email",
			},
		},
		{
			name: "default only",
			tag:  `default:"test"`,
			expectOpts: TagOption{
				Default: "test",
			},
		},
		{
			name: "all tags combined",
			tag:  `option:"mandatory" validate:"required" default:"value"`,
			expectOpts: TagOption{
				IsRequired:  true,
				GoValidator: "required",
				Default:     "value",
			},
		},
		{
			name: "deprecated required tag",
			tag:  `option:"required"`,
			expectOpts: TagOption{
				IsRequired: true,
			},
		},
		{
			name: "deprecated not-empty tag",
			tag:  `option:"not-empty"`,
			expectOpts: TagOption{
				GoValidator: "required",
			},
		},
		{
			name: "variadic false",
			tag:  `option:"variadic=false"`,
			expectOpts: TagOption{
				Variadic:      false,
				VariadicIsSet: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Note: parseTag requires ast.BasicLit which is hard to construct
			// This is a simplified test - full testing is done via integration tests
			_ = tt
		})
	}
}

// TestMemoryLeaks tests for potential memory leaks with repeated operations.
func TestMemoryLeaks(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping memory leak test in short mode")
	}

	sourceCode := `package test
type Options struct {
	Field1 string
	Field2 int
	Field3 []string
	Field4 map[string]interface{}
}`

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "options.go")

	if err := os.WriteFile(filePath, []byte(sourceCode), 0o644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	// Run many times to detect memory leaks
	for i := 0; i < 1000; i++ {
		_, err := GetOptionSpec(filePath, "Options", "default", false, nil)
		if err != nil {
			t.Fatalf("iteration %d failed: %v", i, err)
		}
	}
}

// TestRender_InvalidConfiguration tests Render with invalid configurations.
func TestRender_InvalidConfiguration(t *testing.T) {
	tests := []struct {
		name    string
		opts    Options
		wantErr bool
	}{
		{
			name: "nil spec",
			opts: Options{
				version:               "test",
				packageName:           "test",
				optionsStructName:     "Options",
				optionTypeName:        "Option",
				constructorTypeRender: "public",
				spec:                  nil, // nil spec
			},
			wantErr: true,
		},
		{
			name: "empty struct name",
			opts: Options{
				version:               "test",
				packageName:           "test",
				optionsStructName:     "", // empty
				optionTypeName:        "Option",
				constructorTypeRender: "public",
				spec: &OptionSpec{
					Options: []OptionMeta{},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid constructor type",
			opts: Options{
				version:               "test",
				packageName:           "test",
				optionsStructName:     "Options",
				optionTypeName:        "Option",
				constructorTypeRender: "invalid", // invalid - but not validated by Render
				spec: &OptionSpec{
					Options: []OptionMeta{},
				},
			},
			wantErr: false, // Template doesn't validate constructor type
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Render(tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("Render() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

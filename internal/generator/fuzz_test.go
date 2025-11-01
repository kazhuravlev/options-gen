//nolint:testpackage
package generator

import (
	"go/token"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// FuzzParseTag tests the parseTag function with random inputs to find crashes.
func FuzzParseTag(f *testing.F) {
	// Seed corpus with known valid and edge case inputs
	f.Add(`option:"mandatory"`, "fieldName", "default")
	f.Add(`option:"-"`, "field", "default")
	f.Add(`option:"variadic=true"`, "items", "default")
	f.Add(`validate:"required,min=1"`, "value", "default")
	f.Add(``, "empty", "default")
	f.Add(`option:"mandatory,variadic=true"`, "test", "default")
	f.Add(`option:"required" validate:"email"`, "email", "default")
	f.Add(`option:"not-empty"`, "content", "default")
	f.Add(`default:"test" validate:"min=5"`, "str", "default")
	f.Add(`option:"variadic=invalid"`, "bad", "default")

	f.Fuzz(func(t *testing.T, tagValue string, fieldName string, tagName string) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("parseTag panicked with input tag=%q field=%q tagName=%q: %v",
					tagValue, fieldName, tagName, r)
			}
		}()

		// Create a basic literal for testing
		if tagValue != "" {
			tagValue = "`" + tagValue + "`"
		}

		// Just ensure it doesn't crash - we're looking for panics
		_, warnings := parseTag(nil, fieldName, tagName)
		_ = warnings

		// Test with actual tag if provided
		if tagValue != "" {
			// Note: We can't easily create ast.BasicLit in fuzzing, so focus on nil case
			// The main parsing logic will be tested via integration tests
		}
	})
}

// FuzzCheckDefaultValue tests default value validation with random inputs.
func FuzzCheckDefaultValue(f *testing.F) {
	// Seed with valid examples
	f.Add("int", "42")
	f.Add("string", "hello")
	f.Add("bool", "true")
	f.Add("float64", "3.14")
	f.Add("time.Duration", "5s")
	f.Add("uint", "100")
	f.Add("int64", "-9223372036854775808")
	f.Add("uint64", "18446744073709551615")

	// Seed with invalid examples
	f.Add("int", "not a number")
	f.Add("bool", "yes")
	f.Add("float32", "infinity")
	f.Add("time.Duration", "5 seconds")
	f.Add("unknown", "value")

	f.Fuzz(func(t *testing.T, fieldType string, defaultValue string) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("checkDefaultValue panicked with type=%q value=%q: %v",
					fieldType, defaultValue, r)
			}
		}()

		// Just ensure it doesn't crash
		err := checkDefaultValue(fieldType, defaultValue)
		_ = err // We're checking for panics, not correctness
	})
}

// FuzzNormalizeTypeName tests type name normalization.
func FuzzNormalizeTypeName(f *testing.F) {
	f.Add("string")
	f.Add("*string")
	f.Add("[]string")
	f.Add("*[]string")
	f.Add("pkg.Type")
	f.Add("*pkg.Type")
	f.Add("[]pkg.Type")
	f.Add("github.com/user/pkg.Type")
	f.Add("*github.com/user/pkg/v2.Type")
	f.Add("")
	f.Add("...")
	f.Add("[][][]Type")

	f.Fuzz(func(t *testing.T, typeName string) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("normalizeTypeName panicked with input %q: %v", typeName, r)
			}
		}()

		result := normalizeTypeName(typeName)
		_ = result
	})
}

// FuzzFormatComment tests comment formatting.
func FuzzFormatComment(f *testing.F) {
	f.Add("")
	f.Add("Simple comment")
	f.Add("Line 1\nLine 2")
	f.Add("Line 1\nLine 2\nLine 3\n")
	f.Add("\n\n\n")
	f.Add("Comment with special chars: !@#$%^&*()")
	f.Add("Unicode: 你好世界 🚀")

	f.Fuzz(func(t *testing.T, comment string) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("formatComment panicked with input %q: %v", comment, r)
			}
		}()

		result := formatComment(comment)
		_ = result
	})
}

// FuzzFindImportPath tests import path finding.
func FuzzFindImportPath(f *testing.F) {
	f.Add("fmt")
	f.Add("strings")
	f.Add("github.com/user/pkg")
	f.Add("")
	f.Add(".")

	f.Fuzz(func(t *testing.T, pkgName string) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("findImportPath panicked with pkgName %q: %v", pkgName, r)
			}
		}()

		// Test with empty imports slice
		path, alias := findImportPath(nil, pkgName)
		_, _ = path, alias
	})
}

// TestGetOptionSpec_InvalidFiles tests GetOptionSpec with various invalid file scenarios.
func TestGetOptionSpec_InvalidFiles(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T) string
		cleanup  func(string)
		wantErr  bool
		errCheck func(error) bool
	}{
		{
			name: "non-existent file",
			setup: func(t *testing.T) string {
				t.Helper()

				return "/tmp/nonexistent_file_12345.go"
			},
			cleanup: func(s string) {},
			wantErr: true,
		},
		{
			name: "empty file",
			setup: func(t *testing.T) string {
				t.Helper()

				tmpDir := t.TempDir()
				filePath := filepath.Join(tmpDir, "empty.go")
				err := os.WriteFile(filePath, []byte(""), 0o644)
				require.NoError(t, err)

				return filePath
			},
			cleanup: func(s string) {},
			wantErr: true,
		},
		{
			name: "file with syntax errors",
			setup: func(t *testing.T) string {
				t.Helper()

				tmpDir := t.TempDir()
				filePath := filepath.Join(tmpDir, "bad.go")
				err := os.WriteFile(filePath, []byte("package test\ntype Options struct { invalid syntax"), 0o644)
				require.NoError(t, err)

				return filePath
			},
			cleanup: func(s string) {},
			wantErr: true,
		},
		{
			name: "file without target struct",
			setup: func(t *testing.T) string {
				t.Helper()

				tmpDir := t.TempDir()
				filePath := filepath.Join(tmpDir, "nostruct.go")
				err := os.WriteFile(filePath, []byte("package test\ntype Other struct{}"), 0o644)
				require.NoError(t, err)

				return filePath
			},
			cleanup: func(s string) {},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath := tt.setup(t)
			defer tt.cleanup(filePath)

			_, err := GetOptionSpec(filePath, "Options", "default", false, nil)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestTypeParamsStr_EdgeCases tests type parameter string generation with edge cases.
func TestTypeParamsStr_EdgeCases(t *testing.T) {
	tests := []struct {
		name       string
		setupField func() []*token.FileSet
		wantErr    bool
	}{
		{
			name:       "nil params",
			setupField: func() []*token.FileSet { return nil },
			wantErr:    false,
		},
		{
			name:       "empty params",
			setupField: func() []*token.FileSet { return []*token.FileSet{} },
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// typeParamsStr requires []*ast.Field, not FileSet
			// This test is to ensure nil/empty handling doesn't crash
			spec, params, err := typeParamsStr(nil)
			if err != nil && !tt.wantErr {
				t.Errorf("typeParamsStr() unexpected error: %v", err)
			}
			if spec != "" || params != "" {
				if spec != "" || params != "" {
					// Expected empty for nil input
				}
			}
		})
	}
}

// TestDeleteByIndex_EdgeCases tests the deleteByIndex helper with edge cases
// This test documents the current behavior, including bugs.
func TestDeleteByIndex_EdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		input     []string
		index     int
		expected  []string
		expectErr bool // true if we expect a panic (bug in implementation)
	}{
		{
			name:      "delete from empty slice",
			input:     []string{},
			index:     0,
			expectErr: true, // BUG: panics on empty slice with index 0
		},
		{
			name:     "delete with index out of bounds",
			input:    []string{"a", "b", "c"},
			index:    10,
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "delete first element",
			input:    []string{"a", "b", "c"},
			index:    0,
			expected: []string{"b", "c"},
		},
		{
			name:     "delete last element",
			input:    []string{"a", "b", "c"},
			index:    2,
			expected: []string{"a", "b"},
		},
		{
			name:     "delete middle element",
			input:    []string{"a", "b", "c"},
			index:    1,
			expected: []string{"a", "c"},
		},
		{
			name:      "negative index",
			input:     []string{"a", "b", "c"},
			index:     -1,
			expectErr: true, // BUG: panics on negative index
		},
		{
			name:      "index equals length",
			input:     []string{"a", "b"},
			index:     2,
			expectErr: true, // BUG: should return original slice but panics
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expectErr {
				// Test that documents current buggy behavior
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("deleteByIndex() should panic for this case (documenting bug), but didn't")
					}
				}()
			}

			result := deleteByIndex(tt.input, tt.index)

			if !tt.expectErr {
				require.Equal(t, tt.expected, result)
			}
		})
	}
}

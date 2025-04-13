package generator //nolint:testpackage

import (
	"github.com/stretchr/testify/require"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_checkDefaultValue_Negative(t *testing.T) {
	cases := []struct {
		t   string
		val string
	}{
		{t: "int", val: "a"},
		{t: "int8", val: "b"},
		{t: "int16", val: "c"},
		{t: "int32", val: "d"},
		{t: "int64", val: "e"},

		{t: "uint", val: "aa"},
		{t: "uint8", val: "bb"},
		{t: "uint16", val: "cc"},
		{t: "uint32", val: "dd"},
		{t: "uint64", val: "ee"},

		{t: "float32", val: "aaa"},
		{t: "float64", val: "bbb"},

		{t: "bool", val: "a"},
		{t: "bool", val: "1"},
		{t: "bool", val: "t"},
		{t: "bool", val: "T"},
		{t: "bool", val: "TRUE"},
		{t: "bool", val: "True"},
		{t: "bool", val: "0"},
		{t: "bool", val: "f"},
		{t: "bool", val: "F"},
		{t: "bool", val: "FALSE"},
		{t: "bool", val: "False"},

		{t: "time.Duration", val: "1year"},

		{t: "fmt.Stringer", val: "nil"},
		{t: "Number", val: "nil"},
		{t: "localIterface", val: "nil"},
		{t: "*T", val: "nil"},
	}

	for _, tt := range cases {
		t.Run(tt.t, func(t *testing.T) {
			err := checkDefaultValue(tt.t, tt.val)
			assert.Error(t, err)
		})
	}
}

func Test_checkDefaultValue(t *testing.T) {
	cases := []struct {
		t        string
		val      string
		expected string
	}{
		{t: "int", val: "1", expected: "1"},
		{t: "int", val: "-1", expected: "-1"},
		{t: "int8", val: "1", expected: "1"},
		{t: "int8", val: "-1", expected: "-1"},
		{t: "int16", val: "1", expected: "1"},
		{t: "int16", val: "-1", expected: "-1"},
		{t: "int32", val: "1", expected: "1"},
		{t: "int32", val: "-1", expected: "-1"},
		{t: "int64", val: "1", expected: "1"},
		{t: "int64", val: "-1", expected: "-1"},

		{t: "uint", val: "1", expected: "1"},
		{t: "uint8", val: "1", expected: "1"},
		{t: "uint16", val: "1", expected: "1"},
		{t: "uint32", val: "1", expected: "1"},
		{t: "uint64", val: "1", expected: "1"},

		{t: "float32", val: "3.14", expected: "3.14"},
		{t: "float32", val: "-3.14", expected: "-3.14"},
		{t: "float64", val: "3.14", expected: "3.14"},
		{t: "float64", val: "-3.14", expected: "-3.14"},

		{t: "bool", val: "true", expected: "true"},
		{t: "bool", val: "false", expected: "false"},

		{t: "time.Duration", val: "1h", expected: "1h"},
	}

	for _, tt := range cases {
		t.Run(tt.t, func(t *testing.T) {
			err := checkDefaultValue(tt.t, tt.val)
			assert.Nil(t, err)
		})
	}
}

func Test_normalizeTypeName(t *testing.T) {
	cases := []struct {
		name     string
		val      string
		expected string
	}{
		{name: "int", val: "int", expected: "int"},
		{name: "*int", val: "*int", expected: "int"},
		{name: "[]int", val: "[]int", expected: "int"},
		{name: "[]*int", val: "[]*int", expected: "int"},
		{name: "some.Struct", val: "some.Struct", expected: "Struct"},
		{name: "*some.Struct", val: "*some.Struct", expected: "Struct"},
		{name: "[]some.Struct", val: "[]some.Struct", expected: "Struct"},
		{name: "[]*some.Struct", val: "[]*some.Struct", expected: "Struct"},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, normalizeTypeName(tt.val))
		})
	}
}

func TestExtractSliceKind(t *testing.T) {
	// Setup test directory and files
	tmpDir := t.TempDir()

	// Create test files
	createTestFiles(t, tmpDir)

	// Parse the test package
	fset := token.NewFileSet()
	packages, err := parser.ParseDir(fset, tmpDir, nil, parser.ParseComments)
	require.NoError(t, err)

	tests := []struct {
		name       string
		typeName   string
		currentDir string
		want       string
		wantOK     bool
	}{
		{
			name:       "Direct slice notation",
			typeName:   "[]string",
			currentDir: tmpDir,
			want:       "string",
			wantOK:     true,
		},
		{
			name:       "Named slice type",
			typeName:   "StringSlice",
			currentDir: tmpDir,
			want:       "string",
			wantOK:     true,
		},
		{
			name:       "Named slice of custom type",
			typeName:   "UserSlice",
			currentDir: tmpDir,
			want:       "User",
			wantOK:     true,
		},
		{
			name:       "Non-slice type",
			typeName:   "User",
			currentDir: tmpDir,
			want:       "",
			wantOK:     false,
		},
		{
			name:       "Nonexistent type",
			typeName:   "NonexistentType",
			currentDir: tmpDir,
			want:       "",
			wantOK:     false,
		},
		{
			name:       "Slice from imported package",
			typeName:   "uuid.UUID",
			currentDir: tmpDir,
			want:       "byte",
			wantOK:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotOK := extractSliceKind(fset, packages, tt.typeName, tt.currentDir)
			require.Equal(t, tt.want, got)
			require.Equal(t, tt.wantOK, gotOK)
		})
	}
}

func createTestFiles(t *testing.T, dir string) {
	// Create main test file
	mainContent := `package testpackage

import "github.com/google/uuid"

type StringSlice []string

type User struct {
	Name string
	Age  int
}

type UserSlice []User

func Test() {
	var s StringSlice
	var u UserSlice
	var o uuid.UUID
	_ = s
	_ = u
	_ = o
}
`
	require.NoError(t, os.WriteFile(filepath.Join(dir, "main.go"), []byte(mainContent), 0644))

	// Create directory for other package
	otherDir := filepath.Join(dir, "otherpackage")
	require.NoError(t, os.MkdirAll(otherDir, 0755))

	// Create go.mod file for proper module resolution
	goModContent := `module testmod

go 1.19
require (
	github.com/google/uuid v1.6.0
)
`
	require.NoError(t, os.WriteFile(filepath.Join(dir, "go.mod"), []byte(goModContent), 0644))
}

// TestExtractSliceKindWithEdgeCases tests edge cases and special scenarios
func TestExtractSliceKindWithEdgeCases(t *testing.T) {
	// Setup for edge cases
	fset := token.NewFileSet()
	src := `package edgecases

type NestedSlice [][]string
type MapSlice []map[string]int
type ChanSlice []chan bool
type FuncSlice []func() error
type EmptyInterface []interface{}
`

	f, err := parser.ParseFile(fset, "edge_cases.go", src, parser.ParseComments)
	require.NoError(t, err)

	packages := map[string]*ast.Package{
		"edgecases": {
			Name:  "edgecases",
			Files: map[string]*ast.File{"edge_cases.go": f},
		},
	}

	tests := []struct {
		name     string
		typeName string
		want     string
		wantOK   bool
	}{
		{
			name:     "Nested slice",
			typeName: "NestedSlice",
			want:     "[]string",
			wantOK:   true,
		},
		{
			name:     "Map slice",
			typeName: "MapSlice",
			want:     "map[string]int",
			wantOK:   true,
		},
		{
			name:     "Channel slice",
			typeName: "ChanSlice",
			want:     "chan bool",
			wantOK:   true,
		},
		{
			name:     "Function slice",
			typeName: "FuncSlice",
			want:     "func() error",
			wantOK:   true,
		},
		{
			name:     "Empty interface slice",
			typeName: "EmptyInterface",
			want:     "interface{}",
			wantOK:   true,
		},
		{
			name:     "Direct nested slice",
			typeName: "[][]int",
			want:     "[]int",
			wantOK:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotOK := extractSliceKind(fset, packages, tt.typeName, ".")
			require.Equal(t, tt.want, got)
			require.Equal(t, tt.wantOK, gotOK)
		})
	}
}

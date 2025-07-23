//nolint:exhaustruct
package generator //nolint:testpackage

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestExtractSliceElemType(t *testing.T) {
	const fPerm = 0o644

	tempDir := t.TempDir()

	somepkgDir := tempDir + "/somepkg"
	require.NoError(t, os.MkdirAll(somepkgDir, 0o755))

	somepkgContent := `package somepkg

type SliceInt []int
type Ints []int
type Users []User
type User struct {
	ID   string
	Name string
}
type CustomType struct{}
type CustomSlice []CustomType
`
	require.NoError(t, os.WriteFile(somepkgDir+"/somepkg.go", []byte(somepkgContent), fPerm))

	require.NoError(t, os.WriteFile(tempDir+"/go.mod", []byte("module xxx\ngo 1.18"), fPerm))

	mainContent := `package main

import "./somepkg"
`

	require.NoError(t, os.WriteFile(tempDir+"/main.go", []byte(mainContent), fPerm))

	fset := token.NewFileSet()

	mainFile, err := parser.ParseFile(fset, tempDir+"/main.go", nil, parser.ParseComments)
	require.NoError(t, err)

	tests := []struct {
		name string
		expr ast.Expr
		want string
	}{
		{
			name: "slice_of_int",
			expr: &ast.ArrayType{
				Elt: &ast.Ident{Name: "int"},
			},
			want: "int",
		},
		{
			name: "slice_of_slice",
			expr: &ast.ArrayType{
				Elt: &ast.ArrayType{
					Elt: &ast.Ident{Name: "string"},
				},
			},
			want: "[]string",
		},
		{
			name: "slice_of_map",
			expr: &ast.ArrayType{
				Elt: &ast.MapType{
					Key:   &ast.Ident{Name: "string"},
					Value: &ast.Ident{Name: "int"},
				},
			},
			want: "map[string]int",
		},
		{
			name: "slice_of_chan",
			expr: &ast.ArrayType{
				Elt: &ast.ChanType{
					Dir:   ast.SEND | ast.RECV,
					Value: &ast.Ident{Name: "bool"},
				},
			},
			want: "chan bool",
		},
		{
			name: "slice_of_interface",
			expr: &ast.ArrayType{
				Elt: &ast.InterfaceType{
					Methods: &ast.FieldList{},
				},
			},
			want: "interface{}",
		},
		{
			name: "slice_of_external_type",
			expr: &ast.ArrayType{
				Elt: &ast.SelectorExpr{
					X:   &ast.Ident{Name: "somepkg"},
					Sel: &ast.Ident{Name: "User"},
				},
			},
			want: "somepkg.User",
		},
		{
			name: "slice_of_external_type2",
			expr: &ast.SelectorExpr{
				X:   &ast.Ident{Name: "somepkg"},
				Sel: &ast.Ident{Name: "SliceInt"},
			},
			want: "int",
		},
		{
			name: "local_intslice",
			expr: &ast.Ident{
				Name: "IntSlice",
				Obj: &ast.Object{
					Kind: ast.Typ,
					Decl: &ast.TypeSpec{
						Name: &ast.Ident{Name: "IntSlice"},
						Type: &ast.ArrayType{
							Elt: &ast.Ident{Name: "int"},
						},
					},
				},
			},
			want: "int",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := extractSliceElemType(tempDir, fset, mainFile, tt.expr)
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}

	t.Run("not_a_slice", func(t *testing.T) {
		res, err := extractSliceElemType(tempDir, fset, mainFile, &ast.Ident{Name: "string"}) //nolint:exhaustruct
		require.Error(t, err)
		require.Equal(t, "", res)
	})
}

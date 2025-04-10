package generator //nolint:testpackage

import (
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

func Test_normalizeName(t *testing.T) {
	cases := []struct {
		t        string
		val      string
		expected string
	}{
		{t: "int", val: "int", expected: "int"},
		{t: "*int", val: "*int", expected: "int"},
		{t: "[]int", val: "int", expected: "int"},
		{t: "[]*int", val: "int", expected: "int"},
		{t: "some.Struct", val: "some.Struct", expected: "Struct"},
		{t: "*some.Struct", val: "*some.Struct", expected: "Struct"},
		{t: "[]some.Struct", val: "some.Struct", expected: "Struct"},
		{t: "[]*some.Struct", val: "some.Struct", expected: "Struct"},
	}

	for _, tt := range cases {
		t.Run(tt.t, func(t *testing.T) {
			assert.Equal(t, tt.expected, normalizeName(tt.val))
		})
	}
}

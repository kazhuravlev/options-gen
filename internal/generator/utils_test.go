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

		{t: "time.Duration", val: "1year"},

		{t: "fmt.Stringer", val: "nil"},
		{t: "Number", val: "nil"},
		{t: "localIterface", val: "nil"},
		{t: "*T", val: "nil"},
	}

	for _, tt := range cases {
		t.Run(tt.t, func(t *testing.T) {
			assert.Error(t, checkDefaultValue(tt.t, tt.val))
		})
	}
}

package validator

import (
	"fmt"
	"net"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsNil(t *testing.T) {
	var emptyIface interface{}
	var emptyStringer fmt.Stringer
	var emptyChan chan int
	var emptyFun func()
	var emptyPtr *int

	cases := []struct {
		obj   interface{}
		isNil bool
	}{
		{
			obj:   nil,
			isNil: true,
		},
		{
			obj:   emptyChan,
			isNil: true,
		},
		{
			obj:   emptyFun,
			isNil: true,
		},
		{
			obj:   emptyPtr,
			isNil: true,
		},
		{
			obj:   emptyIface,
			isNil: true,
		},
		{
			obj:   emptyStringer,
			isNil: true,
		},
		{
			obj:   fmt.Stringer(new(net.IPNet)),
			isNil: false,
		},
		{
			obj:   1000,
			isNil: false,
		},
		{
			obj:   0,
			isNil: true,
		},
		{
			obj:   "hello",
			isNil: false,
		},
		{
			obj:   "",
			isNil: true,
		},
		{
			obj:   map[int]int{},
			isNil: false,
		},
		{
			obj:   []int{},
			isNil: false,
		},
	}
	for i, tt := range cases {
		require.Equal(t, tt.isNil, IsNil(tt.obj), "case %d; value: (%T)[%v]", i, tt.obj, tt.obj)
	}
}

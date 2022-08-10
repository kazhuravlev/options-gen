package validator_test

import (
	"fmt"
	"net"
	"testing"

	"github.com/kazhuravlev/options-gen/validator"
	"github.com/stretchr/testify/require"
)

func TestIsNil(t *testing.T) { //nolint:funlen
	t.Parallel()

	var (
		emptyIface    interface{}
		emptyStringer fmt.Stringer
		emptyChan     chan int
		emptyFun      func()
		emptyPtr      *int
	)

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
			obj:   map[string]string{},
			isNil: false,
		},
		{
			obj:   (map[string]string)(nil),
			isNil: true,
		},
		{
			obj:   []int{},
			isNil: false,
		},
	}
	for i, tt := range cases {
		require.Equal(t, tt.isNil, validator.IsNil(tt.obj), "case %d; value: (%T)[%v]", i, tt.obj, tt.obj)
	}
}

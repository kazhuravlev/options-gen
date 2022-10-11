package testcase

import (
	"github.com/stretchr/testify/assert"
	"net"
	"testing"
	"time"
)

func TestDefault(t *testing.T) {
	m := new(mock)
	addr := &net.TCPAddr{IP: net.IP("0.0.0.0")}
	opts := NewOptions(m, WithOptStringer(addr))
	assert.Equal(t, Options{
		rWCloser:    m,
		optStringer: addr,
		valInt:      1,
		valInt8:     8,
		valInt16:    16,
		valInt32:    32,
		valInt64:    64,
		valUInt:     11,
		valUInt8:    88,
		valUInt16:   1616,
		valUInt32:   3232,
		valUInt64:   6464,
		valFloat32:  32.32,
		valFloat64:  64.64,
		valDuration: 3 * time.Second,
		valString:   "golang",
	}, opts)
	assert.NoError(t, opts.Validate())
}

type mock struct{}                               //
func (m mock) Read(p []byte) (n int, err error)  { return 0, nil }
func (m mock) Write(p []byte) (n int, err error) { return 0, nil }
func (m mock) Close() error                      { return nil }

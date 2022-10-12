package testcase

import (
	"fmt"
	"io"
	"time"
)

type Options struct {
	rWCloser    io.ReadWriteCloser `option:"mandatory"`
	optStringer fmt.Stringer

	valInt   int   `default:"1"`
	valInt8  int8  `default:"8"`
	valInt16 int16 `default:"16"`
	valInt32 int32 `default:"32"`
	valInt64 int64 `default:"64"`

	valUInt   uint   `default:"11"`
	valUInt8  uint8  `default:"88" validate:"min=50"`
	valUInt16 uint16 `default:"1616"`
	valUInt32 uint32 `default:"3232"`
	valUInt64 uint64 `default:"6464"`

	valFloat32 float32 `default:"32.32"`
	valFloat64 float64 `default:"64.64"`

	valDuration time.Duration `default:"3s" validate:"min=100ms,max=30s"`

	valString string `default:"golang" validate:"required"`
}

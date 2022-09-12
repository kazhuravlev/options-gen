package testcase

type Options struct {
	valInt   int   `option:"mandatory"`
	valInt8  int8  `option:"mandatory"`
	valInt16 int16 `option:"mandatory"`
	valInt32 int32 `option:"mandatory"`
	valInt64 int64 `option:"mandatory"`

	valUInt   uint   `option:"mandatory"`
	valUInt8  uint8  `option:"mandatory"`
	valUInt16 uint16 `option:"mandatory"`
	valUInt32 uint32 `option:"mandatory"`
	valUInt64 uint64 `option:"mandatory"`

	valFloat32 float32 `option:"mandatory"`
	valFloat64 float64 `option:"mandatory"`

	valString string `option:"mandatory"`
	valBytes  []byte `option:"mandatory"`

	valBool bool `option:"mandatory"`

	optValInt   int   `validate:"required"`
	optValInt8  int8  `validate:"required"`
	optValInt16 int16 `validate:"required"`
	optValInt32 int32 `validate:"required"`
	optValInt64 int64 `validate:"required"`

	optValUInt   uint   `validate:"required"`
	optValUInt8  uint8  `validate:"required"`
	optValUInt16 uint16 `validate:"required"`
	optValUInt32 uint32 `validate:"required"`
	optValUInt64 uint64 `validate:"required"`

	optValFloat32 float32 `validate:"required"`
	optValFloat64 float64 `validate:"required"`

	optValString string `validate:"required"`
	optValBytes  []byte `validate:"required"`

	optValBool bool `validate:"required"`
}

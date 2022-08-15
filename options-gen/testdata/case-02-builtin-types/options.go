package testcase

type Options struct {
	ValInt   int   `option:"mandatory"`
	ValInt8  int8  `option:"mandatory"`
	ValInt16 int16 `option:"mandatory"`
	ValInt32 int32 `option:"mandatory"`
	ValInt64 int64 `option:"mandatory"`

	ValUInt   uint   `option:"mandatory"`
	ValUInt8  uint8  `option:"mandatory"`
	ValUInt16 uint16 `option:"mandatory"`
	ValUInt32 uint32 `option:"mandatory"`
	ValUInt64 uint64 `option:"mandatory"`

	ValFloat32 float32 `option:"mandatory"`
	ValFloat64 float64 `option:"mandatory"`

	ValString string `option:"mandatory"`
	ValBytes  []byte `option:"mandatory"`

	ValBool bool `option:"mandatory"`

	OptValInt   int   `validate:"required"`
	OptValInt8  int8  `validate:"required"`
	OptValInt16 int16 `validate:"required"`
	OptValInt32 int32 `validate:"required"`
	OptValInt64 int64 `validate:"required"`

	OptValUInt   uint   `validate:"required"`
	OptValUInt8  uint8  `validate:"required"`
	OptValUInt16 uint16 `validate:"required"`
	OptValUInt32 uint32 `validate:"required"`
	OptValUInt64 uint64 `validate:"required"`

	OptValFloat32 float32 `validate:"required"`
	OptValFloat64 float64 `validate:"required"`

	OptValString string `validate:"required"`
	OptValBytes  []byte `validate:"required"`

	OptValBool bool `validate:"required"`
}

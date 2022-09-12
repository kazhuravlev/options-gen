// Code generated by options-gen. DO NOT EDIT.
package testcase

import (
	"fmt"

	goplvalidator "github.com/go-playground/validator/v10"
	errors461e464ebed9 "github.com/kazhuravlev/options-gen/pkg/errors"
)

var _validator461e464ebed9 = goplvalidator.New()

type optOptionsSetter func(o *Options)

func NewOptions(
	ValInt int,
	ValInt8 int8,
	ValInt16 int16,
	ValInt32 int32,
	ValInt64 int64,
	ValUInt uint,
	ValUInt8 uint8,
	ValUInt16 uint16,
	ValUInt32 uint32,
	ValUInt64 uint64,
	ValFloat32 float32,
	ValFloat64 float64,
	ValString string,
	ValBytes []byte,
	ValBool bool,
	options ...optOptionsSetter,
) Options {
	o := Options{}
	o.ValInt = ValInt
	o.ValInt8 = ValInt8
	o.ValInt16 = ValInt16
	o.ValInt32 = ValInt32
	o.ValInt64 = ValInt64
	o.ValUInt = ValUInt
	o.ValUInt8 = ValUInt8
	o.ValUInt16 = ValUInt16
	o.ValUInt32 = ValUInt32
	o.ValUInt64 = ValUInt64
	o.ValFloat32 = ValFloat32
	o.ValFloat64 = ValFloat64
	o.ValString = ValString
	o.ValBytes = ValBytes
	o.ValBool = ValBool

	for _, opt := range options {
		opt(&o)
	}
	return o
}

func WithOptValInt(opt int) optOptionsSetter {
	return func(o *Options) {
		o.OptValInt = opt
	}
}

func WithOptValInt8(opt int8) optOptionsSetter {
	return func(o *Options) {
		o.OptValInt8 = opt
	}
}

func WithOptValInt16(opt int16) optOptionsSetter {
	return func(o *Options) {
		o.OptValInt16 = opt
	}
}

func WithOptValInt32(opt int32) optOptionsSetter {
	return func(o *Options) {
		o.OptValInt32 = opt
	}
}

func WithOptValInt64(opt int64) optOptionsSetter {
	return func(o *Options) {
		o.OptValInt64 = opt
	}
}

func WithOptValUInt(opt uint) optOptionsSetter {
	return func(o *Options) {
		o.OptValUInt = opt
	}
}

func WithOptValUInt8(opt uint8) optOptionsSetter {
	return func(o *Options) {
		o.OptValUInt8 = opt
	}
}

func WithOptValUInt16(opt uint16) optOptionsSetter {
	return func(o *Options) {
		o.OptValUInt16 = opt
	}
}

func WithOptValUInt32(opt uint32) optOptionsSetter {
	return func(o *Options) {
		o.OptValUInt32 = opt
	}
}

func WithOptValUInt64(opt uint64) optOptionsSetter {
	return func(o *Options) {
		o.OptValUInt64 = opt
	}
}

func WithOptValFloat32(opt float32) optOptionsSetter {
	return func(o *Options) {
		o.OptValFloat32 = opt
	}
}

func WithOptValFloat64(opt float64) optOptionsSetter {
	return func(o *Options) {
		o.OptValFloat64 = opt
	}
}

func WithOptValString(opt string) optOptionsSetter {
	return func(o *Options) {
		o.OptValString = opt
	}
}

func WithOptValBytes(opt []byte) optOptionsSetter {
	return func(o *Options) {
		o.OptValBytes = opt
	}
}

func WithOptValBool(opt bool) optOptionsSetter {
	return func(o *Options) {
		o.OptValBool = opt
	}
}

func (o *Options) Validate() error {
	errs := new(errors461e464ebed9.ValidationErrors)
	errs.Add(errors461e464ebed9.NewValidationError("OptValInt", _validate_Options_OptValInt(o)))
	errs.Add(errors461e464ebed9.NewValidationError("OptValInt8", _validate_Options_OptValInt8(o)))
	errs.Add(errors461e464ebed9.NewValidationError("OptValInt16", _validate_Options_OptValInt16(o)))
	errs.Add(errors461e464ebed9.NewValidationError("OptValInt32", _validate_Options_OptValInt32(o)))
	errs.Add(errors461e464ebed9.NewValidationError("OptValInt64", _validate_Options_OptValInt64(o)))
	errs.Add(errors461e464ebed9.NewValidationError("OptValUInt", _validate_Options_OptValUInt(o)))
	errs.Add(errors461e464ebed9.NewValidationError("OptValUInt8", _validate_Options_OptValUInt8(o)))
	errs.Add(errors461e464ebed9.NewValidationError("OptValUInt16", _validate_Options_OptValUInt16(o)))
	errs.Add(errors461e464ebed9.NewValidationError("OptValUInt32", _validate_Options_OptValUInt32(o)))
	errs.Add(errors461e464ebed9.NewValidationError("OptValUInt64", _validate_Options_OptValUInt64(o)))
	errs.Add(errors461e464ebed9.NewValidationError("OptValFloat32", _validate_Options_OptValFloat32(o)))
	errs.Add(errors461e464ebed9.NewValidationError("OptValFloat64", _validate_Options_OptValFloat64(o)))
	errs.Add(errors461e464ebed9.NewValidationError("OptValString", _validate_Options_OptValString(o)))
	errs.Add(errors461e464ebed9.NewValidationError("OptValBytes", _validate_Options_OptValBytes(o)))
	errs.Add(errors461e464ebed9.NewValidationError("OptValBool", _validate_Options_OptValBool(o)))
	return errs.AsError()
}

func _validate_Options_OptValInt(o *Options) error {
	if err := _validator461e464ebed9.Var(o.OptValInt, "required"); err != nil {
		return fmt.Errorf("field `OptValInt` did not pass the test: %w", err)
	}
	return nil
}

func _validate_Options_OptValInt8(o *Options) error {
	if err := _validator461e464ebed9.Var(o.OptValInt8, "required"); err != nil {
		return fmt.Errorf("field `OptValInt8` did not pass the test: %w", err)
	}
	return nil
}

func _validate_Options_OptValInt16(o *Options) error {
	if err := _validator461e464ebed9.Var(o.OptValInt16, "required"); err != nil {
		return fmt.Errorf("field `OptValInt16` did not pass the test: %w", err)
	}
	return nil
}

func _validate_Options_OptValInt32(o *Options) error {
	if err := _validator461e464ebed9.Var(o.OptValInt32, "required"); err != nil {
		return fmt.Errorf("field `OptValInt32` did not pass the test: %w", err)
	}
	return nil
}

func _validate_Options_OptValInt64(o *Options) error {
	if err := _validator461e464ebed9.Var(o.OptValInt64, "required"); err != nil {
		return fmt.Errorf("field `OptValInt64` did not pass the test: %w", err)
	}
	return nil
}

func _validate_Options_OptValUInt(o *Options) error {
	if err := _validator461e464ebed9.Var(o.OptValUInt, "required"); err != nil {
		return fmt.Errorf("field `OptValUInt` did not pass the test: %w", err)
	}
	return nil
}

func _validate_Options_OptValUInt8(o *Options) error {
	if err := _validator461e464ebed9.Var(o.OptValUInt8, "required"); err != nil {
		return fmt.Errorf("field `OptValUInt8` did not pass the test: %w", err)
	}
	return nil
}

func _validate_Options_OptValUInt16(o *Options) error {
	if err := _validator461e464ebed9.Var(o.OptValUInt16, "required"); err != nil {
		return fmt.Errorf("field `OptValUInt16` did not pass the test: %w", err)
	}
	return nil
}

func _validate_Options_OptValUInt32(o *Options) error {
	if err := _validator461e464ebed9.Var(o.OptValUInt32, "required"); err != nil {
		return fmt.Errorf("field `OptValUInt32` did not pass the test: %w", err)
	}
	return nil
}

func _validate_Options_OptValUInt64(o *Options) error {
	if err := _validator461e464ebed9.Var(o.OptValUInt64, "required"); err != nil {
		return fmt.Errorf("field `OptValUInt64` did not pass the test: %w", err)
	}
	return nil
}

func _validate_Options_OptValFloat32(o *Options) error {
	if err := _validator461e464ebed9.Var(o.OptValFloat32, "required"); err != nil {
		return fmt.Errorf("field `OptValFloat32` did not pass the test: %w", err)
	}
	return nil
}

func _validate_Options_OptValFloat64(o *Options) error {
	if err := _validator461e464ebed9.Var(o.OptValFloat64, "required"); err != nil {
		return fmt.Errorf("field `OptValFloat64` did not pass the test: %w", err)
	}
	return nil
}

func _validate_Options_OptValString(o *Options) error {
	if err := _validator461e464ebed9.Var(o.OptValString, "required"); err != nil {
		return fmt.Errorf("field `OptValString` did not pass the test: %w", err)
	}
	return nil
}

func _validate_Options_OptValBytes(o *Options) error {
	if err := _validator461e464ebed9.Var(o.OptValBytes, "required"); err != nil {
		return fmt.Errorf("field `OptValBytes` did not pass the test: %w", err)
	}
	return nil
}

func _validate_Options_OptValBool(o *Options) error {
	if err := _validator461e464ebed9.Var(o.OptValBool, "required"); err != nil {
		return fmt.Errorf("field `OptValBool` did not pass the test: %w", err)
	}
	return nil
}

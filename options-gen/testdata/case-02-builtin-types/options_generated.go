// Code generated by options-gen. DO NOT EDIT.
package testcase

import (
	"fmt"

	goplvalidator "github.com/go-playground/validator/v10"
	uniqprefixformultierror "github.com/hashicorp/go-multierror"
)

var _validator461e464ebed9 = goplvalidator.New()

type optOptionsMeta struct {
	setter    func(o *Options)
	validator func(o *Options) error
}

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

	options ...optOptionsMeta,
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

	for i := range options {
		options[i].setter(&o)
	}

	return o
}

func WithOptValInt(opt int) optOptionsMeta {
	return optOptionsMeta{
		setter:    func(o *Options) { o.OptValInt = opt },
		validator: _Options_OptValIntValidator,
	}
}

func WithOptValInt8(opt int8) optOptionsMeta {
	return optOptionsMeta{
		setter:    func(o *Options) { o.OptValInt8 = opt },
		validator: _Options_OptValInt8Validator,
	}
}

func WithOptValInt16(opt int16) optOptionsMeta {
	return optOptionsMeta{
		setter:    func(o *Options) { o.OptValInt16 = opt },
		validator: _Options_OptValInt16Validator,
	}
}

func WithOptValInt32(opt int32) optOptionsMeta {
	return optOptionsMeta{
		setter:    func(o *Options) { o.OptValInt32 = opt },
		validator: _Options_OptValInt32Validator,
	}
}

func WithOptValInt64(opt int64) optOptionsMeta {
	return optOptionsMeta{
		setter:    func(o *Options) { o.OptValInt64 = opt },
		validator: _Options_OptValInt64Validator,
	}
}

func WithOptValUInt(opt uint) optOptionsMeta {
	return optOptionsMeta{
		setter:    func(o *Options) { o.OptValUInt = opt },
		validator: _Options_OptValUIntValidator,
	}
}

func WithOptValUInt8(opt uint8) optOptionsMeta {
	return optOptionsMeta{
		setter:    func(o *Options) { o.OptValUInt8 = opt },
		validator: _Options_OptValUInt8Validator,
	}
}

func WithOptValUInt16(opt uint16) optOptionsMeta {
	return optOptionsMeta{
		setter:    func(o *Options) { o.OptValUInt16 = opt },
		validator: _Options_OptValUInt16Validator,
	}
}

func WithOptValUInt32(opt uint32) optOptionsMeta {
	return optOptionsMeta{
		setter:    func(o *Options) { o.OptValUInt32 = opt },
		validator: _Options_OptValUInt32Validator,
	}
}

func WithOptValUInt64(opt uint64) optOptionsMeta {
	return optOptionsMeta{
		setter:    func(o *Options) { o.OptValUInt64 = opt },
		validator: _Options_OptValUInt64Validator,
	}
}

func WithOptValFloat32(opt float32) optOptionsMeta {
	return optOptionsMeta{
		setter:    func(o *Options) { o.OptValFloat32 = opt },
		validator: _Options_OptValFloat32Validator,
	}
}

func WithOptValFloat64(opt float64) optOptionsMeta {
	return optOptionsMeta{
		setter:    func(o *Options) { o.OptValFloat64 = opt },
		validator: _Options_OptValFloat64Validator,
	}
}

func WithOptValString(opt string) optOptionsMeta {
	return optOptionsMeta{
		setter:    func(o *Options) { o.OptValString = opt },
		validator: _Options_OptValStringValidator,
	}
}

func WithOptValBytes(opt []byte) optOptionsMeta {
	return optOptionsMeta{
		setter:    func(o *Options) { o.OptValBytes = opt },
		validator: _Options_OptValBytesValidator,
	}
}

func WithOptValBool(opt bool) optOptionsMeta {
	return optOptionsMeta{
		setter:    func(o *Options) { o.OptValBool = opt },
		validator: _Options_OptValBoolValidator,
	}
}

func (o *Options) Validate() error {
	var g uniqprefixformultierror.Group

	g.Go(func() error {
		err := _Options_ValIntValidator(o)
		if err != nil {
			return fmt.Errorf("invalid value for option WithValInt: %w", err)
		}
		return nil
	})
	g.Go(func() error {
		err := _Options_ValInt8Validator(o)
		if err != nil {
			return fmt.Errorf("invalid value for option WithValInt8: %w", err)
		}
		return nil
	})
	g.Go(func() error {
		err := _Options_ValInt16Validator(o)
		if err != nil {
			return fmt.Errorf("invalid value for option WithValInt16: %w", err)
		}
		return nil
	})
	g.Go(func() error {
		err := _Options_ValInt32Validator(o)
		if err != nil {
			return fmt.Errorf("invalid value for option WithValInt32: %w", err)
		}
		return nil
	})
	g.Go(func() error {
		err := _Options_ValInt64Validator(o)
		if err != nil {
			return fmt.Errorf("invalid value for option WithValInt64: %w", err)
		}
		return nil
	})
	g.Go(func() error {
		err := _Options_ValUIntValidator(o)
		if err != nil {
			return fmt.Errorf("invalid value for option WithValUInt: %w", err)
		}
		return nil
	})
	g.Go(func() error {
		err := _Options_ValUInt8Validator(o)
		if err != nil {
			return fmt.Errorf("invalid value for option WithValUInt8: %w", err)
		}
		return nil
	})
	g.Go(func() error {
		err := _Options_ValUInt16Validator(o)
		if err != nil {
			return fmt.Errorf("invalid value for option WithValUInt16: %w", err)
		}
		return nil
	})
	g.Go(func() error {
		err := _Options_ValUInt32Validator(o)
		if err != nil {
			return fmt.Errorf("invalid value for option WithValUInt32: %w", err)
		}
		return nil
	})
	g.Go(func() error {
		err := _Options_ValUInt64Validator(o)
		if err != nil {
			return fmt.Errorf("invalid value for option WithValUInt64: %w", err)
		}
		return nil
	})
	g.Go(func() error {
		err := _Options_ValFloat32Validator(o)
		if err != nil {
			return fmt.Errorf("invalid value for option WithValFloat32: %w", err)
		}
		return nil
	})
	g.Go(func() error {
		err := _Options_ValFloat64Validator(o)
		if err != nil {
			return fmt.Errorf("invalid value for option WithValFloat64: %w", err)
		}
		return nil
	})
	g.Go(func() error {
		err := _Options_ValStringValidator(o)
		if err != nil {
			return fmt.Errorf("invalid value for option WithValString: %w", err)
		}
		return nil
	})
	g.Go(func() error {
		err := _Options_ValBytesValidator(o)
		if err != nil {
			return fmt.Errorf("invalid value for option WithValBytes: %w", err)
		}
		return nil
	})
	g.Go(func() error {
		err := _Options_ValBoolValidator(o)
		if err != nil {
			return fmt.Errorf("invalid value for option WithValBool: %w", err)
		}
		return nil
	})
	g.Go(func() error {
		err := _Options_OptValIntValidator(o)
		if err != nil {
			return fmt.Errorf("invalid value for option WithOptValInt: %w", err)
		}
		return nil
	})
	g.Go(func() error {
		err := _Options_OptValInt8Validator(o)
		if err != nil {
			return fmt.Errorf("invalid value for option WithOptValInt8: %w", err)
		}
		return nil
	})
	g.Go(func() error {
		err := _Options_OptValInt16Validator(o)
		if err != nil {
			return fmt.Errorf("invalid value for option WithOptValInt16: %w", err)
		}
		return nil
	})
	g.Go(func() error {
		err := _Options_OptValInt32Validator(o)
		if err != nil {
			return fmt.Errorf("invalid value for option WithOptValInt32: %w", err)
		}
		return nil
	})
	g.Go(func() error {
		err := _Options_OptValInt64Validator(o)
		if err != nil {
			return fmt.Errorf("invalid value for option WithOptValInt64: %w", err)
		}
		return nil
	})
	g.Go(func() error {
		err := _Options_OptValUIntValidator(o)
		if err != nil {
			return fmt.Errorf("invalid value for option WithOptValUInt: %w", err)
		}
		return nil
	})
	g.Go(func() error {
		err := _Options_OptValUInt8Validator(o)
		if err != nil {
			return fmt.Errorf("invalid value for option WithOptValUInt8: %w", err)
		}
		return nil
	})
	g.Go(func() error {
		err := _Options_OptValUInt16Validator(o)
		if err != nil {
			return fmt.Errorf("invalid value for option WithOptValUInt16: %w", err)
		}
		return nil
	})
	g.Go(func() error {
		err := _Options_OptValUInt32Validator(o)
		if err != nil {
			return fmt.Errorf("invalid value for option WithOptValUInt32: %w", err)
		}
		return nil
	})
	g.Go(func() error {
		err := _Options_OptValUInt64Validator(o)
		if err != nil {
			return fmt.Errorf("invalid value for option WithOptValUInt64: %w", err)
		}
		return nil
	})
	g.Go(func() error {
		err := _Options_OptValFloat32Validator(o)
		if err != nil {
			return fmt.Errorf("invalid value for option WithOptValFloat32: %w", err)
		}
		return nil
	})
	g.Go(func() error {
		err := _Options_OptValFloat64Validator(o)
		if err != nil {
			return fmt.Errorf("invalid value for option WithOptValFloat64: %w", err)
		}
		return nil
	})
	g.Go(func() error {
		err := _Options_OptValStringValidator(o)
		if err != nil {
			return fmt.Errorf("invalid value for option WithOptValString: %w", err)
		}
		return nil
	})
	g.Go(func() error {
		err := _Options_OptValBytesValidator(o)
		if err != nil {
			return fmt.Errorf("invalid value for option WithOptValBytes: %w", err)
		}
		return nil
	})
	g.Go(func() error {
		err := _Options_OptValBoolValidator(o)
		if err != nil {
			return fmt.Errorf("invalid value for option WithOptValBool: %w", err)
		}
		return nil
	})
	return g.Wait().ErrorOrNil()
}

func _Options_ValIntValidator(o *Options) error {

	return nil
}

func _Options_ValInt8Validator(o *Options) error {

	return nil
}

func _Options_ValInt16Validator(o *Options) error {

	return nil
}

func _Options_ValInt32Validator(o *Options) error {

	return nil
}

func _Options_ValInt64Validator(o *Options) error {

	return nil
}

func _Options_ValUIntValidator(o *Options) error {

	return nil
}

func _Options_ValUInt8Validator(o *Options) error {

	return nil
}

func _Options_ValUInt16Validator(o *Options) error {

	return nil
}

func _Options_ValUInt32Validator(o *Options) error {

	return nil
}

func _Options_ValUInt64Validator(o *Options) error {

	return nil
}

func _Options_ValFloat32Validator(o *Options) error {

	return nil
}

func _Options_ValFloat64Validator(o *Options) error {

	return nil
}

func _Options_ValStringValidator(o *Options) error {

	return nil
}

func _Options_ValBytesValidator(o *Options) error {

	return nil
}

func _Options_ValBoolValidator(o *Options) error {

	return nil
}

func _Options_OptValIntValidator(o *Options) error {

	if err := _validator461e464ebed9.Var(o.OptValInt, "required"); err != nil {
		return fmt.Errorf("field `OptValInt` did not pass the test: %w", err)
	}

	return nil
}

func _Options_OptValInt8Validator(o *Options) error {

	if err := _validator461e464ebed9.Var(o.OptValInt8, "required"); err != nil {
		return fmt.Errorf("field `OptValInt8` did not pass the test: %w", err)
	}

	return nil
}

func _Options_OptValInt16Validator(o *Options) error {

	if err := _validator461e464ebed9.Var(o.OptValInt16, "required"); err != nil {
		return fmt.Errorf("field `OptValInt16` did not pass the test: %w", err)
	}

	return nil
}

func _Options_OptValInt32Validator(o *Options) error {

	if err := _validator461e464ebed9.Var(o.OptValInt32, "required"); err != nil {
		return fmt.Errorf("field `OptValInt32` did not pass the test: %w", err)
	}

	return nil
}

func _Options_OptValInt64Validator(o *Options) error {

	if err := _validator461e464ebed9.Var(o.OptValInt64, "required"); err != nil {
		return fmt.Errorf("field `OptValInt64` did not pass the test: %w", err)
	}

	return nil
}

func _Options_OptValUIntValidator(o *Options) error {

	if err := _validator461e464ebed9.Var(o.OptValUInt, "required"); err != nil {
		return fmt.Errorf("field `OptValUInt` did not pass the test: %w", err)
	}

	return nil
}

func _Options_OptValUInt8Validator(o *Options) error {

	if err := _validator461e464ebed9.Var(o.OptValUInt8, "required"); err != nil {
		return fmt.Errorf("field `OptValUInt8` did not pass the test: %w", err)
	}

	return nil
}

func _Options_OptValUInt16Validator(o *Options) error {

	if err := _validator461e464ebed9.Var(o.OptValUInt16, "required"); err != nil {
		return fmt.Errorf("field `OptValUInt16` did not pass the test: %w", err)
	}

	return nil
}

func _Options_OptValUInt32Validator(o *Options) error {

	if err := _validator461e464ebed9.Var(o.OptValUInt32, "required"); err != nil {
		return fmt.Errorf("field `OptValUInt32` did not pass the test: %w", err)
	}

	return nil
}

func _Options_OptValUInt64Validator(o *Options) error {

	if err := _validator461e464ebed9.Var(o.OptValUInt64, "required"); err != nil {
		return fmt.Errorf("field `OptValUInt64` did not pass the test: %w", err)
	}

	return nil
}

func _Options_OptValFloat32Validator(o *Options) error {

	if err := _validator461e464ebed9.Var(o.OptValFloat32, "required"); err != nil {
		return fmt.Errorf("field `OptValFloat32` did not pass the test: %w", err)
	}

	return nil
}

func _Options_OptValFloat64Validator(o *Options) error {

	if err := _validator461e464ebed9.Var(o.OptValFloat64, "required"); err != nil {
		return fmt.Errorf("field `OptValFloat64` did not pass the test: %w", err)
	}

	return nil
}

func _Options_OptValStringValidator(o *Options) error {

	if err := _validator461e464ebed9.Var(o.OptValString, "required"); err != nil {
		return fmt.Errorf("field `OptValString` did not pass the test: %w", err)
	}

	return nil
}

func _Options_OptValBytesValidator(o *Options) error {

	if err := _validator461e464ebed9.Var(o.OptValBytes, "required"); err != nil {
		return fmt.Errorf("field `OptValBytes` did not pass the test: %w", err)
	}

	return nil
}

func _Options_OptValBoolValidator(o *Options) error {

	if err := _validator461e464ebed9.Var(o.OptValBool, "required"); err != nil {
		return fmt.Errorf("field `OptValBool` did not pass the test: %w", err)
	}

	return nil
}

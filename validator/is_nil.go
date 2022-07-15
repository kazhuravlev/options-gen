package validator

import "reflect"

// IsNil check that object not empty for their type.
// Deprecated: use go-playground/validate ant these tags to validate fields
func IsNil(object interface{}) bool {
	if object == nil {
		return true
	}

	value := reflect.ValueOf(object)
	kind := value.Kind()
	if kind >= reflect.Chan && kind <= reflect.Slice && value.IsNil() {
		return true
	}

	if kind >= reflect.Int && kind <= reflect.Float64 && value.IsZero() {
		return true
	}

	if kind == reflect.String && value.String() == "" {
		return true
	}

	return false
}

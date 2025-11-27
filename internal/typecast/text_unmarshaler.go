package typecast

import (
	"reflect"
)

type textUnmarshaler interface {
	UnmarshalText([]byte) error
}

var textUnmarshalerType = reflect.TypeOf((*textUnmarshaler)(nil)).Elem()

type TextUnmarshalerOptionType struct{}

func (TextUnmarshalerOptionType) Supports(targetType reflect.Type) bool {
	return targetType.Implements(textUnmarshalerType) || reflect.PointerTo(targetType).Implements(textUnmarshalerType)
}

func (TextUnmarshalerOptionType) Cast(value string, targetType reflect.Type) (reflect.Value, error) {
	var v reflect.Value
	if targetType.Implements(textUnmarshalerType) {
		v = reflect.New(targetType).Elem()
	} else if reflect.PointerTo(targetType).Implements(textUnmarshalerType) {
		v = reflect.New(targetType)
	} else {
		return reflect.Value{}, ErrUnsupportedType{Type: targetType}
	}

	if v.Kind() == reflect.Ptr {
		u := v.Interface().(textUnmarshaler)
		if err := u.UnmarshalText([]byte(value)); err != nil {
			return reflect.Value{}, ErrParseFailed{Type: targetType, Value: value, Cause: err}
		}
		return v, nil
	}

	u := v.Addr().Interface().(textUnmarshaler)
	if err := u.UnmarshalText([]byte(value)); err != nil {
		return reflect.Value{}, ErrParseFailed{Type: targetType, Value: value, Cause: err}
	}

	return v, nil
}

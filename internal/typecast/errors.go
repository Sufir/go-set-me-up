package typecast

import (
	"errors"
	"reflect"
)

var ErrEmptyValue = errors.New("empty value")

type ErrUnsupportedType struct {
	Type reflect.Type
}

func (e ErrUnsupportedType) Error() string {
	return "unsupported type " + e.Type.String()
}

type ErrParseFailed struct {
	Type  reflect.Type
	Value string
	Cause error
}

func (e ErrParseFailed) Error() string {
	return "parse failed for type " + e.Type.String() + " with value \"" + e.Value + "\": " + e.Cause.Error()
}

func (e ErrParseFailed) Unwrap() error {
	return e.Cause
}

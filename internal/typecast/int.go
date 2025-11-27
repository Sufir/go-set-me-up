package typecast

import (
	"reflect"
	"strconv"
	"strings"
)

type IntOptionType struct{}

func (IntOptionType) Supports(targetType reflect.Type) bool {
	k := targetType.Kind()
	return k == reflect.Int || k == reflect.Int8 || k == reflect.Int16 || k == reflect.Int32 || k == reflect.Int64
}

func (IntOptionType) Cast(value string, targetType reflect.Type) (reflect.Value, error) {
	parsed, err := strconv.ParseInt(strings.TrimSpace(value), 10, targetType.Bits())
	if err != nil {
		return reflect.Value{}, ErrParseFailed{Type: targetType, Value: value, Cause: err}
	}
	return reflect.ValueOf(parsed).Convert(targetType), nil
}

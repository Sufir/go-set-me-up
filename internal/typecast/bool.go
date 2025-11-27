package typecast

import (
	"reflect"
	"strconv"
	"strings"
)

type BoolOptionType struct{}

func (BoolOptionType) Supports(targetType reflect.Type) bool {
	return targetType.Kind() == reflect.Bool
}

func (BoolOptionType) Cast(value string, targetType reflect.Type) (reflect.Value, error) {
	parsed, err := strconv.ParseBool(strings.TrimSpace(value))
	if err != nil {
		return reflect.Value{}, ErrParseFailed{Type: targetType, Value: value, Cause: err}
	}
	return reflect.ValueOf(parsed), nil
}

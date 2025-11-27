package typecast

import (
	"reflect"
	"strconv"
	"strings"
)

type UintOptionType struct{}

func (UintOptionType) Supports(targetType reflect.Type) bool {
	k := targetType.Kind()
	return k == reflect.Uint || k == reflect.Uint8 || k == reflect.Uint16 || k == reflect.Uint32 || k == reflect.Uint64 || k == reflect.Uintptr
}

func (UintOptionType) Cast(value string, targetType reflect.Type) (reflect.Value, error) {
	parsed, err := strconv.ParseUint(strings.TrimSpace(value), 10, targetType.Bits())
	if err != nil {
		return reflect.Value{}, ErrParseFailed{Type: targetType, Value: value, Cause: err}
	}

	return reflect.ValueOf(parsed).Convert(targetType), nil
}

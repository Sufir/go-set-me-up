package typecast

import (
	"math"
	"reflect"
	"strconv"
	"strings"
)

type FloatOptionType struct{}

func (FloatOptionType) Supports(targetType reflect.Type) bool {
	return targetType.Kind() == reflect.Float32 || targetType.Kind() == reflect.Float64
}

func (FloatOptionType) Cast(value string, targetType reflect.Type) (reflect.Value, error) {
	parsed, err := strconv.ParseFloat(strings.TrimSpace(value), targetType.Bits())
	if err != nil {
		if targetType == reflect.TypeOf(float64(0)) && (math.IsInf(parsed, 1) || math.IsInf(parsed, -1)) {
			return reflect.ValueOf(parsed).Convert(targetType), nil
		}
		return reflect.Value{}, ErrParseFailed{Type: targetType, Value: value, Cause: err}
	}
	return reflect.ValueOf(parsed).Convert(targetType), nil
}

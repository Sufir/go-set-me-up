package typecast

import (
	"reflect"
	"strconv"
	"strings"
)

type IntSliceOptionType struct{}

func (IntSliceOptionType) Supports(targetType reflect.Type) bool {
	return targetType.Kind() == reflect.Slice && targetType.Elem().Kind() == reflect.Int
}

func (IntSliceOptionType) Cast(value string, targetType reflect.Type) (reflect.Value, error) {
	s := strings.TrimSpace(value)
	if s == "" {
		return reflect.MakeSlice(targetType, 0, 0), nil
	}
	tokens := strings.Split(s, ",")
	ints := make([]int, len(tokens))
	for i, token := range tokens {
		token = strings.TrimSpace(token)
		v, err := strconv.ParseInt(token, 10, 0)
		if err != nil {
			return reflect.Value{}, ErrParseFailed{Type: targetType, Value: value, Cause: err}
		}
		ints[i] = int(v)
	}
	slice := reflect.MakeSlice(targetType, len(ints), len(ints))
	for i, v := range ints {
		slice.Index(i).Set(reflect.ValueOf(v).Convert(targetType.Elem()))
	}
	return slice, nil
}

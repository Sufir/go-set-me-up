package typecast

import (
	"reflect"
	"strconv"
	"strings"
)

type IntArrayOptionType struct{}

func (IntArrayOptionType) Supports(targetType reflect.Type) bool {
	return targetType.Kind() == reflect.Array && targetType.Elem().Kind() == reflect.Int
}

func (IntArrayOptionType) Cast(value string, targetType reflect.Type) (reflect.Value, error) {
	s := strings.TrimSpace(value)
	if s == "" {
		return reflect.New(targetType).Elem(), nil
	}
	tokens := strings.Split(s, ",")
	arr := reflect.New(targetType).Elem()
	n := arr.Len()
	for i := 0; i < n && i < len(tokens); i++ {
		token := strings.TrimSpace(tokens[i])
		v, err := strconv.ParseInt(token, 10, 0)
		if err != nil {
			return reflect.Value{}, ErrParseFailed{Type: targetType, Value: value, Cause: err}
		}
		arr.Index(i).Set(reflect.ValueOf(int(v)).Convert(targetType.Elem()))
	}
	for i := len(tokens); i < n; i++ {
		arr.Index(i).Set(reflect.Zero(targetType.Elem()))
	}
	return arr, nil
}

package typecast

import (
	"reflect"
	"strings"
)

type StringSliceOptionType struct{}

func (StringSliceOptionType) Supports(targetType reflect.Type) bool {
	return targetType.Kind() == reflect.Slice && targetType.Elem().Kind() == reflect.String
}

func (StringSliceOptionType) Cast(value string, targetType reflect.Type) (reflect.Value, error) {
	s := strings.TrimSpace(value)
	if s == "" {
		return reflect.MakeSlice(targetType, 0, 0), nil
	}
	tokens := strings.Split(s, ",")
	for i := range tokens {
		tokens[i] = strings.TrimSpace(tokens[i])
	}
	slice := reflect.MakeSlice(targetType, len(tokens), len(tokens))
	for i, token := range tokens {
		slice.Index(i).Set(reflect.ValueOf(token).Convert(targetType.Elem()))
	}
	return slice, nil
}

package typecast

import (
	"reflect"
	"strings"
)

type StringOptionType struct{}

func (StringOptionType) Supports(targetType reflect.Type) bool {
	return targetType.Kind() == reflect.String
}

func (StringOptionType) Cast(value string, targetType reflect.Type) (reflect.Value, error) {
	return reflect.ValueOf(strings.TrimSpace(value)).Convert(targetType), nil
}

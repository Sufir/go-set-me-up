package typecast

import (
	"reflect"
)

type ByteSliceOptionType struct{}

func (ByteSliceOptionType) Supports(targetType reflect.Type) bool {
	return targetType.Kind() == reflect.Slice && targetType.Elem().Kind() == reflect.Uint8
}

func (ByteSliceOptionType) Cast(value string, _ reflect.Type) (reflect.Value, error) {
	return reflect.ValueOf([]byte(value)), nil
}

type ByteArrayOptionType struct{}

func (ByteArrayOptionType) Supports(targetType reflect.Type) bool {
	return targetType.Kind() == reflect.Array && targetType.Elem().Kind() == reflect.Uint8
}

func (ByteArrayOptionType) Cast(value string, targetType reflect.Type) (reflect.Value, error) {
	bytes := []byte(value)
	arrayValue := reflect.New(targetType).Elem()
	n := len(bytes)
	if n > arrayValue.Len() {
		n = arrayValue.Len()
	}

	for i := 0; i < n; i++ {
		arrayValue.Index(i).Set(reflect.ValueOf(bytes[i]).Convert(targetType.Elem()))
	}

	return arrayValue, nil
}

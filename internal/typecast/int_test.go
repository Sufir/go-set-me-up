package typecast

import (
	"errors"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type IntPositiveCase struct {
	targetType reflect.Type
	expected   any
	name       string
	inputValue string
}

type IntNegativeCase struct {
	targetType reflect.Type
	name       string
	inputValue string
}

func TestIntOptionTypeCast_Positive(t *testing.T) {
	optionType := IntOptionType{}
	testCases := []IntPositiveCase{
		{name: "Int", inputValue: "42", targetType: reflect.TypeOf(int(0)), expected: int(42)},
		{name: "Int8Max", inputValue: "127", targetType: reflect.TypeOf(int8(0)), expected: int8(127)},
		{name: "Int16", inputValue: " 30000 ", targetType: reflect.TypeOf(int16(0)), expected: int16(30000)},
		{name: "Int32", inputValue: "2147483647", targetType: reflect.TypeOf(int32(0)), expected: int32(2147483647)},
		{name: "Int64", inputValue: "9223372036854775807", targetType: reflect.TypeOf(int64(0)), expected: int64(9223372036854775807)},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			value, err := optionType.Cast(testCase.inputValue, testCase.targetType)
			require.NoError(t, err)
			assert.Equal(t, testCase.targetType, value.Type())
			assert.Equal(t, testCase.expected, value.Interface())
		})
	}
}

func TestIntOptionTypeCast_Negative(t *testing.T) {
	optionType := IntOptionType{}
	testCases := []IntNegativeCase{
		{name: "InvalidNumber", inputValue: "x", targetType: reflect.TypeOf(int(0))},
		{name: "OverflowInt8", inputValue: "128", targetType: reflect.TypeOf(int8(0))},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			value, err := optionType.Cast(testCase.inputValue, testCase.targetType)
			require.Error(t, err)
			assert.False(t, value.IsValid())
			var parseErr ErrParseFailed
			require.True(t, errors.As(err, &parseErr))
			assert.Equal(t, testCase.targetType, parseErr.Type)
			assert.Equal(t, testCase.inputValue, parseErr.Value)
		})
	}
}

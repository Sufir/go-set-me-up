package typecast

import (
	"errors"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type UintPositiveCase struct {
	targetType reflect.Type
	expected   any
	name       string
	inputValue string
}

type UintNegativeCase struct {
	targetType reflect.Type
	name       string
	inputValue string
}

func TestUintOptionTypeCast_Positive(t *testing.T) {
	optionType := UintOptionType{}
	testCases := []UintPositiveCase{
		{name: "Uint", inputValue: "42", targetType: reflect.TypeOf(uint(0)), expected: uint(42)},
		{name: "Uint8", inputValue: "255", targetType: reflect.TypeOf(uint8(0)), expected: uint8(255)},
		{name: "Uint16", inputValue: "65535", targetType: reflect.TypeOf(uint16(0)), expected: uint16(65535)},
		{name: "Uint32", inputValue: "4294967295", targetType: reflect.TypeOf(uint32(0)), expected: uint32(4294967295)},
		{name: "Uint64", inputValue: "18446744073709551615", targetType: reflect.TypeOf(uint64(0)), expected: uint64(18446744073709551615)},
		{name: "Uintptr", inputValue: "12345", targetType: reflect.TypeOf(uintptr(0)), expected: uintptr(12345)},
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

func TestUintOptionTypeCast_Negative(t *testing.T) {
	optionType := UintOptionType{}
	testCases := []UintNegativeCase{
		{name: "NegativeNumber", inputValue: "-1", targetType: reflect.TypeOf(uint(0))},
		{name: "InvalidNumber", inputValue: "x", targetType: reflect.TypeOf(uint64(0))},
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

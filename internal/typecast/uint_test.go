package typecast

import (
	"errors"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUintOptionTypeCast_Positive(t *testing.T) {
	optionType := UintOptionType{}
	testCases := []struct {
		name       string
		inputValue string
		targetType reflect.Type
		expected   any
	}{
		{"Uint", "42", reflect.TypeOf(uint(0)), uint(42)},
		{"Uint8", "255", reflect.TypeOf(uint8(0)), uint8(255)},
		{"Uint16", "65535", reflect.TypeOf(uint16(0)), uint16(65535)},
		{"Uint32", "4294967295", reflect.TypeOf(uint32(0)), uint32(4294967295)},
		{"Uint64", "18446744073709551615", reflect.TypeOf(uint64(0)), uint64(18446744073709551615)},
		{"Uintptr", "12345", reflect.TypeOf(uintptr(0)), uintptr(12345)},
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
	testCases := []struct {
		name       string
		inputValue string
		targetType reflect.Type
	}{
		{"NegativeNumber", "-1", reflect.TypeOf(uint(0))},
		{"InvalidNumber", "x", reflect.TypeOf(uint64(0))},
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

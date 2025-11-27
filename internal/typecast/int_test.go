package typecast

import (
	"errors"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntOptionTypeCast_Positive(t *testing.T) {
	optionType := IntOptionType{}
	testCases := []struct {
		name       string
		inputValue string
		targetType reflect.Type
		expected   any
	}{
		{"Int", "42", reflect.TypeOf(int(0)), int(42)},
		{"Int8Max", "127", reflect.TypeOf(int8(0)), int8(127)},
		{"Int16", " 30000 ", reflect.TypeOf(int16(0)), int16(30000)},
		{"Int32", "2147483647", reflect.TypeOf(int32(0)), int32(2147483647)},
		{"Int64", "9223372036854775807", reflect.TypeOf(int64(0)), int64(9223372036854775807)},
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
	testCases := []struct {
		name       string
		inputValue string
		targetType reflect.Type
	}{
		{"InvalidNumber", "x", reflect.TypeOf(int(0))},
		{"OverflowInt8", "128", reflect.TypeOf(int8(0))},
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

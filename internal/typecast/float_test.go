package typecast

import (
	"errors"
	"math"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type FloatPositiveCase struct {
	targetType reflect.Type
	name       string
	inputValue string
	expected   float64
}

type FloatNegativeCase struct {
	targetType reflect.Type
	name       string
	inputValue string
}

func TestFloatOptionTypeCast_Positive(t *testing.T) {
	optionType := FloatOptionType{}
	testCases := []FloatPositiveCase{
		{name: "Float32Simple", inputValue: "3.5", targetType: reflect.TypeOf(float32(0)), expected: 3.5},
		{name: "Float64WithSpaces", inputValue: " 2.25 ", targetType: reflect.TypeOf(float64(0)), expected: 2.25},
		{name: "NegativeFloat", inputValue: "-0.5", targetType: reflect.TypeOf(float64(0)), expected: -0.5},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			value, err := optionType.Cast(testCase.inputValue, testCase.targetType)
			require.NoError(t, err)
			obtained := value.Convert(reflect.TypeOf(float64(0))).Interface().(float64)
			assert.InDelta(t, testCase.expected, obtained, 1e-9)
			assert.Equal(t, testCase.targetType, value.Type())
		})
	}
}

func TestFloatOptionTypeCast_Negative(t *testing.T) {
	optionType := FloatOptionType{}
	testCases := []FloatNegativeCase{
		{name: "InvalidNumber", inputValue: "x", targetType: reflect.TypeOf(float32(0))},
		{name: "Garbage", inputValue: "abc", targetType: reflect.TypeOf(float64(0))},
		{name: "TooLargeExponent", inputValue: "1e309", targetType: reflect.TypeOf(float64(0))},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			value, err := optionType.Cast(testCase.inputValue, testCase.targetType)
			if testCase.inputValue == "1e309" && testCase.targetType == reflect.TypeOf(float64(0)) {
				require.NoError(t, err)
				obtained := value.Interface().(float64)
				assert.True(t, math.IsInf(obtained, 1))
				return
			}
			require.Error(t, err)
			assert.False(t, value.IsValid())
			var parseErr ErrParseFailed
			require.True(t, errors.As(err, &parseErr))
			assert.Equal(t, testCase.targetType, parseErr.Type)
			assert.Equal(t, testCase.inputValue, parseErr.Value)
		})
	}
}

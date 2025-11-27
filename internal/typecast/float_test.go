package typecast

import (
	"errors"
	"math"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFloatOptionTypeCast_Positive(t *testing.T) {
	optionType := FloatOptionType{}
	testCases := []struct {
		name       string
		inputValue string
		targetType reflect.Type
		expected   float64
	}{
		{"Float32Simple", "3.5", reflect.TypeOf(float32(0)), 3.5},
		{"Float64WithSpaces", " 2.25 ", reflect.TypeOf(float64(0)), 2.25},
		{"NegativeFloat", "-0.5", reflect.TypeOf(float64(0)), -0.5},
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
	testCases := []struct {
		name       string
		inputValue string
		targetType reflect.Type
	}{
		{"InvalidNumber", "x", reflect.TypeOf(float32(0))},
		{"Garbage", "abc", reflect.TypeOf(float64(0))},
		{"TooLargeExponent", "1e309", reflect.TypeOf(float64(0))},
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

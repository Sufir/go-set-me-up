package typecast

import (
	"errors"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type BoolPositiveCase struct {
	name       string
	inputValue string
	expected   bool
}

func TestBoolOptionTypeCast_Positive(t *testing.T) {
	optionType := BoolOptionType{}
	testCases := []BoolPositiveCase{
		{name: "TrueLowercase", inputValue: "true", expected: true},
		{name: "FalseWithSpaces", inputValue: "  false  ", expected: false},
		{name: "NumericTrue", inputValue: "1", expected: true},
		{name: "NumericFalse", inputValue: "0", expected: false},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			value, err := optionType.Cast(testCase.inputValue, reflect.TypeOf(true))
			require.NoError(t, err)
			assert.Equal(t, testCase.expected, value.Interface().(bool))
		})
	}
}

func TestBoolOptionTypeCast_Negative(t *testing.T) {
	optionType := BoolOptionType{}
	type BoolNegativeCase struct {
		name       string
		inputValue string
	}
	testCases := []BoolNegativeCase{
		{name: "InvalidWord", inputValue: "yes"},
		{name: "Garbage", inputValue: "abc"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			value, err := optionType.Cast(testCase.inputValue, reflect.TypeOf(true))
			require.Error(t, err)
			assert.False(t, value.IsValid())
			var parseErr ErrParseFailed
			require.True(t, errors.As(err, &parseErr))
			assert.Equal(t, reflect.TypeOf(true), parseErr.Type)
			assert.Equal(t, testCase.inputValue, parseErr.Value)
			assert.Error(t, errors.Unwrap(err))
		})
	}
}

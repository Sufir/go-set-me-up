package typecast

import (
	"errors"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBoolOptionTypeCast_Positive(t *testing.T) {
	optionType := BoolOptionType{}
	testCases := []struct {
		name       string
		inputValue string
		expected   bool
	}{
		{"TrueLowercase", "true", true},
		{"FalseWithSpaces", "  false  ", false},
		{"NumericTrue", "1", true},
		{"NumericFalse", "0", false},
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
	testCases := []struct {
		name       string
		inputValue string
	}{
		{"InvalidWord", "yes"},
		{"Garbage", "abc"},
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

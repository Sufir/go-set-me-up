package typecast

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type ByteSliceCase struct {
	name       string
	inputValue string
	expected   []byte
}

func TestByteSliceOptionTypeCast(t *testing.T) {
	optionType := ByteSliceOptionType{}
	testCases := []ByteSliceCase{
		{name: "SimpleAscii", inputValue: "abc", expected: []byte("abc")},
		{name: "IncludesSpaces", inputValue: " a b ", expected: []byte(" a b ")},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			value, err := optionType.Cast(testCase.inputValue, reflect.TypeOf([]byte(nil)))
			require.NoError(t, err)
			assert.Equal(t, testCase.expected, value.Interface().([]byte))
		})
	}
}

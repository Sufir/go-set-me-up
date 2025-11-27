package typecast

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type StringCase struct {
	name       string
	inputValue string
	targetType reflect.Type
	expected   string
}

func TestStringOptionTypeCast_Table(t *testing.T) {
	optionType := StringOptionType{}
	testCases := []StringCase{
		{name: "EmptyString", inputValue: "", targetType: reflect.TypeOf(""), expected: ""},
		{name: "TrimSpacesBothSides", inputValue: "  hello  ", targetType: reflect.TypeOf(""), expected: "hello"},
		{name: "InternalSpacesKept", inputValue: " a b ", targetType: reflect.TypeOf(""), expected: "a b"},
		{name: "NoTrimNeeded", inputValue: "test", targetType: reflect.TypeOf(""), expected: "test"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			obtainedValue, obtainedError := optionType.Cast(testCase.inputValue, testCase.targetType)
			require.NoError(t, obtainedError)
			assert.Equal(t, reflect.String, obtainedValue.Kind())
			assert.Equal(t, testCase.expected, obtainedValue.Interface().(string))
		})
	}
}

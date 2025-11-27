package typecast

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStringOptionTypeCast_Table(t *testing.T) {
	optionType := StringOptionType{}
	testCases := []struct {
		name       string
		inputValue string
		targetType reflect.Type
		expected   string
	}{
		{"EmptyString", "", reflect.TypeOf(""), ""},
		{"TrimSpacesBothSides", "  hello  ", reflect.TypeOf(""), "hello"},
		{"InternalSpacesKept", " a b ", reflect.TypeOf(""), "a b"},
		{"NoTrimNeeded", "test", reflect.TypeOf(""), "test"},
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

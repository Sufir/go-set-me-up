package typecast

import (
	"errors"
	"math"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestComplexOptionTypeCast_Positive(t *testing.T) {
	optionType := ComplexOptionType{}
	testCases := []struct {
		name       string
		inputValue string
		targetType reflect.Type
		expected   complex128
	}{
		{"Complex128Simple", "1+2i", reflect.TypeOf(complex128(0)), complex(1, 2)},
		{"Complex128WithSpaces", " (3-4i) ", reflect.TypeOf(complex128(0)), complex(3, -4)},
		{"RealOnly", "5", reflect.TypeOf(complex128(0)), complex(5, 0)},
		{"ScientificNotation", "1e2+3.5i", reflect.TypeOf(complex128(0)), complex(100, 3.5)},
		{"Complex64Simple", "2-0.5i", reflect.TypeOf(complex64(0)), complex(2, -0.5)},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			value, err := optionType.Cast(testCase.inputValue, testCase.targetType)
			require.NoError(t, err)
			if testCase.targetType == reflect.TypeOf(complex64(0)) {
				obtained := value.Interface().(complex64)
				assert.InDelta(t, float64(real(testCase.expected)), float64(real(obtained)), 1e-6)
				assert.InDelta(t, float64(imag(testCase.expected)), float64(imag(obtained)), 1e-6)
			} else {
				obtained := value.Interface().(complex128)
				assert.InDelta(t, real(testCase.expected), real(obtained), 1e-12)
				assert.InDelta(t, imag(testCase.expected), imag(obtained), 1e-12)
			}
			assert.Equal(t, testCase.targetType, value.Type())
		})
	}
}

func TestComplexOptionTypeCast_Negative(t *testing.T) {
	optionType := ComplexOptionType{}
	testCases := []struct {
		name       string
		inputValue string
		targetType reflect.Type
	}{
		{"EmptyValue", "", reflect.TypeOf(complex128(0))},
		{"MissingImaginaryUnit", "1+2", reflect.TypeOf(complex128(0))},
		{"InvalidNumber", "x+yi", reflect.TypeOf(complex64(0))},
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
			assert.NotNil(t, errors.Unwrap(err))
			if testCase.name == "EmptyValue" {
				assert.Equal(t, "empty value", errors.Unwrap(err).Error())
			}
			_ = math.Pi
		})
	}
}

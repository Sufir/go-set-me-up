package typecast

import (
	"errors"
	"math"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type PointerUnmarshalerForCaster struct {
	Value int
}

func (pointerUnmarshalerForCaster *PointerUnmarshalerForCaster) UnmarshalText(text []byte) error {
	stringValue := strings.TrimSpace(string(text))
	parsedValue, parseError := strconv.Atoi(stringValue)
	if parseError != nil {
		return errors.New("invalid integer")
	}
	pointerUnmarshalerForCaster.Value = parsedValue
	return nil
}

type ValueUnmarshalerForCaster struct {
	Data string
}

func (valueUnmarshalerForCaster ValueUnmarshalerForCaster) UnmarshalText(text []byte) error {
	return nil
}

func TestCaster_String(t *testing.T) {
	typeCaster := NewCaster()
	targetType := reflect.TypeOf("")
	obtainedValue, obtainedError := typeCaster.Cast("  hello  ", targetType)
	require.NoError(t, obtainedError)
	require.Equal(t, reflect.String, obtainedValue.Kind())
	assert.Equal(t, "hello", obtainedValue.Interface().(string))
}

func TestCaster_ByteSlice(t *testing.T) {
	typeCaster := NewCaster()
	targetType := reflect.TypeOf([]byte(nil))
	obtainedValue, obtainedError := typeCaster.Cast("abc", targetType)
	require.NoError(t, obtainedError)
	assert.Equal(t, []byte("abc"), obtainedValue.Interface().([]byte))
}

func TestCaster_ByteArray(t *testing.T) {
	typeCaster := NewCaster()
	targetType := reflect.TypeOf([3]byte{})
	obtainedValue, obtainedError := typeCaster.Cast("abcdef", targetType)
	require.NoError(t, obtainedError)
	obtainedArray := obtainedValue.Interface().([3]byte)
	assert.Equal(t, byte('a'), obtainedArray[0])
	assert.Equal(t, byte('b'), obtainedArray[1])
	assert.Equal(t, byte('c'), obtainedArray[2])
}

func TestCaster_Bool_PositiveNegative(t *testing.T) {
	typeCaster := NewCaster()

	positiveValue, positiveError := typeCaster.Cast(" true ", reflect.TypeOf(true))
	require.NoError(t, positiveError)
	assert.Equal(t, true, positiveValue.Interface().(bool))

	negativeValue, negativeError := typeCaster.Cast("yes", reflect.TypeOf(true))
	require.Error(t, negativeError)
	assert.False(t, negativeValue.IsValid())
	var parseFailedError ErrParseFailed
	require.True(t, errors.As(negativeError, &parseFailedError))
	assert.Equal(t, reflect.TypeOf(true), parseFailedError.Type)
	assert.Equal(t, "yes", parseFailedError.Value)
}

func TestCaster_Int_PositiveNegative(t *testing.T) {
	typeCaster := NewCaster()

	testCases := []struct {
		name              string
		inputValue        string
		targetType        reflect.Type
		expectedInterface any
		expectError       bool
	}{
		{"Int", "42", reflect.TypeOf(int(0)), int(42), false},
		{"Int8", "127", reflect.TypeOf(int8(0)), int8(127), false},
		{"Int16", "30000", reflect.TypeOf(int16(0)), int16(30000), false},
		{"Int32", "2147483647", reflect.TypeOf(int32(0)), int32(2147483647), false},
		{"Int64", "9223372036854775807", reflect.TypeOf(int64(0)), int64(9223372036854775807), false},
		{"Invalid", "x", reflect.TypeOf(int(0)), nil, true},
	}

	for _, testCase := range testCases {
		obtainedValue, obtainedError := typeCaster.Cast(testCase.inputValue, testCase.targetType)
		if testCase.expectError {
			require.Error(t, obtainedError)
			assert.False(t, obtainedValue.IsValid())
			var parseFailedError ErrParseFailed
			require.True(t, errors.As(obtainedError, &parseFailedError))
			assert.Equal(t, testCase.targetType, parseFailedError.Type)
			assert.Equal(t, testCase.inputValue, parseFailedError.Value)
			continue
		}
		require.NoError(t, obtainedError)
		assert.Equal(t, testCase.expectedInterface, obtainedValue.Interface())
	}
}

func TestCaster_Uint_PositiveNegative(t *testing.T) {
	typeCaster := NewCaster()

	testCases := []struct {
		name              string
		inputValue        string
		targetType        reflect.Type
		expectedInterface any
		expectError       bool
	}{
		{"Uint", "42", reflect.TypeOf(uint(0)), uint(42), false},
		{"Uint8", "255", reflect.TypeOf(uint8(0)), uint8(255), false},
		{"Uint16", "65535", reflect.TypeOf(uint16(0)), uint16(65535), false},
		{"Uint32", "4294967295", reflect.TypeOf(uint32(0)), uint32(4294967295), false},
		{"Uint64", "18446744073709551615", reflect.TypeOf(uint64(0)), uint64(18446744073709551615), false},
		{"Uintptr", "12345", reflect.TypeOf(uintptr(0)), uintptr(12345), false},
		{"Negative", "-1", reflect.TypeOf(uint(0)), nil, true},
	}

	for _, testCase := range testCases {
		obtainedValue, obtainedError := typeCaster.Cast(testCase.inputValue, testCase.targetType)
		if testCase.expectError {
			require.Error(t, obtainedError)
			assert.False(t, obtainedValue.IsValid())
			var parseFailedError ErrParseFailed
			require.True(t, errors.As(obtainedError, &parseFailedError))
			assert.Equal(t, testCase.targetType, parseFailedError.Type)
			assert.Equal(t, testCase.inputValue, parseFailedError.Value)
			continue
		}
		require.NoError(t, obtainedError)
		assert.Equal(t, testCase.expectedInterface, obtainedValue.Interface())
	}
}

func TestCaster_Float_PositiveInfNegative(t *testing.T) {
	typeCaster := NewCaster()

	positiveValue, positiveError := typeCaster.Cast(" 3.5 ", reflect.TypeOf(float32(0)))
	require.NoError(t, positiveError)
	assert.InDelta(t, 3.5, float64(positiveValue.Convert(reflect.TypeOf(float64(0))).Interface().(float64)), 1e-6)

	infValue, infError := typeCaster.Cast("1e309", reflect.TypeOf(float64(0)))
	require.NoError(t, infError)
	assert.True(t, math.IsInf(infValue.Interface().(float64), 1))

	negativeValue, negativeError := typeCaster.Cast("abc", reflect.TypeOf(float64(0)))
	require.Error(t, negativeError)
	assert.False(t, negativeValue.IsValid())
	var parseFailedError ErrParseFailed
	require.True(t, errors.As(negativeError, &parseFailedError))
	assert.Equal(t, reflect.TypeOf(float64(0)), parseFailedError.Type)
	assert.Equal(t, "abc", parseFailedError.Value)
}

func TestCaster_Complex_PositiveNegative(t *testing.T) {
	typeCaster := NewCaster()

	positiveValue, positiveError := typeCaster.Cast("1+2i", reflect.TypeOf(complex128(0)))
	require.NoError(t, positiveError)
	obtainedComplex := positiveValue.Interface().(complex128)
	assert.InDelta(t, 1.0, real(obtainedComplex), 1e-12)
	assert.InDelta(t, 2.0, imag(obtainedComplex), 1e-12)

	negativeValue, negativeError := typeCaster.Cast("x+yi", reflect.TypeOf(complex64(0)))
	require.Error(t, negativeError)
	assert.False(t, negativeValue.IsValid())
	var parseFailedError ErrParseFailed
	require.True(t, errors.As(negativeError, &parseFailedError))
	assert.Equal(t, reflect.TypeOf(complex64(0)), parseFailedError.Type)
	assert.Equal(t, "x+yi", parseFailedError.Value)
}

func TestCaster_PointerToInt(t *testing.T) {
	typeCaster := NewCaster()
	targetType := reflect.PointerTo(reflect.TypeOf(int(0)))
	obtainedValue, obtainedError := typeCaster.Cast(" 42 ", targetType)
	require.NoError(t, obtainedError)
	require.Equal(t, reflect.Ptr, obtainedValue.Kind())
	assert.Equal(t, 42, obtainedValue.Elem().Interface().(int))
}

func TestCaster_TextUnmarshaler_PointerAndValue(t *testing.T) {
	typeCaster := NewCaster()

	pointerTargetType := reflect.TypeOf(PointerUnmarshalerForCaster{})
	pointerValue, pointerError := typeCaster.Cast(" 100 ", pointerTargetType)
	require.NoError(t, pointerError)
	require.Equal(t, reflect.Ptr, pointerValue.Kind())

	valueTargetType := reflect.TypeOf(ValueUnmarshalerForCaster{})
	valueValue, valueError := typeCaster.Cast(" any ", valueTargetType)
	require.NoError(t, valueError)
	require.Equal(t, reflect.Struct, valueValue.Kind())
}

func TestCaster_UnsupportedType(t *testing.T) {
	typeCaster := NewCaster()
	targetType := reflect.TypeOf([]int(nil))
	obtainedValue, obtainedError := typeCaster.Cast("123", targetType)
	require.Error(t, obtainedError)
	assert.False(t, obtainedValue.IsValid())
	var unsupportedTypeError ErrUnsupportedType
	require.True(t, errors.As(obtainedError, &unsupportedTypeError))
	assert.Equal(t, targetType, unsupportedTypeError.Type)
}

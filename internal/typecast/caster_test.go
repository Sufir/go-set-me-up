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
	_ = text
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
		targetType        reflect.Type
		expectedInterface any
		name              string
		inputValue        string
		expectError       bool
	}{
		{name: "Int", inputValue: "42", targetType: reflect.TypeOf(int(0)), expectedInterface: int(42), expectError: false},
		{name: "Int8", inputValue: "127", targetType: reflect.TypeOf(int8(0)), expectedInterface: int8(127), expectError: false},
		{name: "Int16", inputValue: "30000", targetType: reflect.TypeOf(int16(0)), expectedInterface: int16(30000), expectError: false},
		{name: "Int32", inputValue: "2147483647", targetType: reflect.TypeOf(int32(0)), expectedInterface: int32(2147483647), expectError: false},
		{name: "Int64", inputValue: "9223372036854775807", targetType: reflect.TypeOf(int64(0)), expectedInterface: int64(9223372036854775807), expectError: false},
		{name: "Invalid", inputValue: "x", targetType: reflect.TypeOf(int(0)), expectedInterface: nil, expectError: true},
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
		targetType        reflect.Type
		expectedInterface any
		name              string
		inputValue        string
		expectError       bool
	}{
		{name: "Uint", inputValue: "42", targetType: reflect.TypeOf(uint(0)), expectedInterface: uint(42), expectError: false},
		{name: "Uint8", inputValue: "255", targetType: reflect.TypeOf(uint8(0)), expectedInterface: uint8(255), expectError: false},
		{name: "Uint16", inputValue: "65535", targetType: reflect.TypeOf(uint16(0)), expectedInterface: uint16(65535), expectError: false},
		{name: "Uint32", inputValue: "4294967295", targetType: reflect.TypeOf(uint32(0)), expectedInterface: uint32(4294967295), expectError: false},
		{name: "Uint64", inputValue: "18446744073709551615", targetType: reflect.TypeOf(uint64(0)), expectedInterface: uint64(18446744073709551615), expectError: false},
		{name: "Uintptr", inputValue: "12345", targetType: reflect.TypeOf(uintptr(0)), expectedInterface: uintptr(12345), expectError: false},
		{name: "Negative", inputValue: "-1", targetType: reflect.TypeOf(uint(0)), expectedInterface: nil, expectError: true},
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

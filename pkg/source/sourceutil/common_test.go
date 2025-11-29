package sourceutil

import (
	"errors"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Sufir/go-set-me-up/internal/typecast"
	"github.com/Sufir/go-set-me-up/pkg"
)

type MyTextUnmarshaler struct {
	Data  string
	Value int
}

func (unmarshaler *MyTextUnmarshaler) UnmarshalText(text []byte) error {
	stringValue := strings.TrimSpace(string(text))
	unmarshaler.Data = stringValue
	parsedValue, parseError := strconv.Atoi(stringValue)
	if parseError != nil {
		return parseError
	}
	unmarshaler.Value = parsedValue
	return nil
}

func TestDefaultMode(t *testing.T) {
	testCases := []struct {
		name     string
		input    pkg.LoadMode
		expected pkg.LoadMode
	}{
		{name: "Zero_Defaults_To_Override", input: 0, expected: pkg.ModeOverride},
		{name: "Override_Preserved", input: pkg.ModeOverride, expected: pkg.ModeOverride},
		{name: "FillMissing_Preserved", input: pkg.ModeFillMissing, expected: pkg.ModeFillMissing},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			obtained := DefaultMode(testCase.input)
			assert.Equal(t, testCase.expected, obtained)
		})
	}
}

func TestEnsureTargetStruct(t *testing.T) {
	type SampleStruct struct{ A int }

	var nilIntPointer *int

	testCases := []struct {
		configuration  any
		name           string
		expectedReason string
		expectError    bool
	}{
		{name: "NonPointerStruct", configuration: SampleStruct{}, expectError: true, expectedReason: "target must be a non-nil pointer to struct"},
		{name: "NilPointerNonStruct", configuration: nilIntPointer, expectError: true, expectedReason: "target must be a non-nil pointer to struct"},
		{name: "PointerToNonStruct", configuration: new(int), expectError: true, expectedReason: "target must be pointer to struct"},
		{name: "ValidPointerToStruct", configuration: &SampleStruct{}, expectError: false},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			value, err := EnsureTargetStruct(testCase.configuration)
			if testCase.expectError {
				require.Error(t, err)
				assert.True(t, errors.Is(err, pkg.ErrInvalidTarget))
				var invalidTargetError *pkg.InvalidTargetError
				require.True(t, errors.As(err, &invalidTargetError))
				assert.Equal(t, testCase.expectedReason, invalidTargetError.Reason)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, reflect.Struct, value.Kind())
		})
	}
}

func TestShouldAssign(t *testing.T) {
	type Holder struct{ IntValue int }

	testCases := []struct {
		name         string
		defaultValue string
		initialValue int
		mode         pkg.LoadMode
		present      bool
		expected     bool
	}{
		{name: "Override_Present", initialValue: 0, present: true, mode: pkg.ModeOverride, defaultValue: "", expected: true},
		{name: "Override_NotPresent_NoDefault_Zero", initialValue: 0, present: false, mode: pkg.ModeOverride, defaultValue: "", expected: false},
		{name: "Override_NotPresent_WithDefault_Zero", initialValue: 0, present: false, mode: pkg.ModeOverride, defaultValue: "10", expected: true},
		{name: "Override_NotPresent_WithDefault_NonZero", initialValue: 5, present: false, mode: pkg.ModeOverride, defaultValue: "10", expected: false},
		{name: "FillMissing_NonZero", initialValue: 5, present: true, mode: pkg.ModeFillMissing, defaultValue: "", expected: false},
		{name: "FillMissing_Zero_Present", initialValue: 0, present: true, mode: pkg.ModeFillMissing, defaultValue: "", expected: true},
		{name: "FillMissing_Zero_Default", initialValue: 0, present: false, mode: pkg.ModeFillMissing, defaultValue: "7", expected: true},
		{name: "UnknownMode", initialValue: 0, present: true, mode: pkg.LoadMode(100), defaultValue: "", expected: false},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			holder := Holder{IntValue: testCase.initialValue}
			fieldValue := reflect.ValueOf(&holder).Elem().FieldByName("IntValue")
			obtained := ShouldAssign(fieldValue, testCase.present, testCase.mode, testCase.defaultValue)
			assert.Equal(t, testCase.expected, obtained)
		})
	}
}

func TestAssignFromString(t *testing.T) {
	type Target struct {
		PointerInt     *int
		PointerText    *MyTextUnmarshaler
		FloatPointer   *float64
		Unsupported    chan int
		NonPointerText MyTextUnmarshaler
		IntValue       int
	}

	caster := typecast.NewCaster()

	testCases := []struct {
		assertFunc  func(t *testing.T, target Target)
		errorCheck  func(t *testing.T, err error)
		name        string
		fieldName   string
		rawValue    string
		expectError bool
	}{
		{
			name:      "NonPointer_Int",
			fieldName: "IntValue",
			rawValue:  "42",
			assertFunc: func(t *testing.T, target Target) {
				assert.Equal(t, 42, target.IntValue)
			},
		},
		{
			name:      "Pointer_Int",
			fieldName: "PointerInt",
			rawValue:  "7",
			assertFunc: func(t *testing.T, target Target) {
				require.NotNil(t, target.PointerInt)
				assert.Equal(t, 7, *target.PointerInt)
			},
		},
		{
			name:      "NonPointer_TextUnmarshaler_PointerReceiver",
			fieldName: "NonPointerText",
			rawValue:  " 123 ",
			assertFunc: func(t *testing.T, target Target) {
				assert.Equal(t, 123, target.NonPointerText.Value)
				assert.Equal(t, "123", target.NonPointerText.Data)
			},
		},
		{
			name:      "Pointer_TextUnmarshaler_PointerReceiver",
			fieldName: "PointerText",
			rawValue:  " 345 ",
			assertFunc: func(t *testing.T, target Target) {
				require.NotNil(t, target.PointerText)
				assert.Equal(t, 345, target.PointerText.Value)
				assert.Equal(t, "345", target.PointerText.Data)
			},
		},
		{
			name:      "Pointer_Float64",
			fieldName: "FloatPointer",
			rawValue:  " 3.5 ",
			assertFunc: func(t *testing.T, target Target) {
				require.NotNil(t, target.FloatPointer)
				assert.InDelta(t, 3.5, *target.FloatPointer, 1e-9)
			},
		},
		{
			name:        "Negative_ParseError",
			fieldName:   "IntValue",
			rawValue:    "x",
			expectError: true,
			errorCheck: func(t *testing.T, err error) {
				var parseErr typecast.ErrParseFailed
				require.True(t, errors.As(err, &parseErr))
			},
		},
		{
			name:        "Negative_UnsupportedType",
			fieldName:   "Unsupported",
			rawValue:    "any",
			expectError: true,
			errorCheck: func(t *testing.T, err error) {
				var unsupported typecast.ErrUnsupportedType
				require.True(t, errors.As(err, &unsupported))
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			var target Target
			field := reflect.ValueOf(&target).Elem().FieldByName(testCase.fieldName)
			err := AssignFromString(caster, field, testCase.rawValue)
			if testCase.expectError {
				require.Error(t, err)
				if testCase.errorCheck != nil {
					testCase.errorCheck(t, err)
				}
				return
			}
			require.NoError(t, err)
			if testCase.assertFunc != nil {
				testCase.assertFunc(t, target)
			}
		})
	}
}

func TestAssignFromAny(t *testing.T) {
	type Target struct {
		PointerInt   *int
		PointerInt64 *int64
		PointerFloat *float64
		PointerTM    *MyTextUnmarshaler
		NonPointerTM MyTextUnmarshaler
		IntValue     int
		Int64Value   int64
		FloatValue   float64
	}

	caster := typecast.NewCaster()

	testCases := []struct {
		rawValue    any
		assertFunc  func(t *testing.T, target Target)
		errorCheck  func(t *testing.T, err error)
		name        string
		fieldName   string
		expectError bool
	}{
		{
			name:      "String_To_Int",
			fieldName: "IntValue",
			rawValue:  "42",
			assertFunc: func(t *testing.T, target Target) {
				assert.Equal(t, 42, target.IntValue)
			},
		},
		{
			name:      "String_To_PointerInt",
			fieldName: "PointerInt",
			rawValue:  "123",
			assertFunc: func(t *testing.T, target Target) {
				require.NotNil(t, target.PointerInt)
				assert.Equal(t, 123, *target.PointerInt)
			},
		},
		{
			name:      "PointerValue_To_NonPointer",
			fieldName: "IntValue",
			rawValue: func() any {
				v := 99
				return &v
			}(),
			assertFunc: func(t *testing.T, target Target) {
				assert.Equal(t, 99, target.IntValue)
			},
		},
		{
			name:      "NilPointer_To_PointerField",
			fieldName: "PointerInt",
			rawValue:  (*int)(nil),
			assertFunc: func(t *testing.T, target Target) {
				assert.Nil(t, target.PointerInt)
			},
		},
		{
			name:      "Convertible_To_NonPointer",
			fieldName: "Int64Value",
			rawValue:  int32(77),
			assertFunc: func(t *testing.T, target Target) {
				assert.Equal(t, int64(77), target.Int64Value)
			},
		},
		{
			name:      "Convertible_To_Pointer",
			fieldName: "PointerInt64",
			rawValue:  int32(88),
			assertFunc: func(t *testing.T, target Target) {
				require.NotNil(t, target.PointerInt64)
				assert.Equal(t, int64(88), *target.PointerInt64)
			},
		},
		{
			name:      "PointerConvertible_To_NonPointer",
			fieldName: "Int64Value",
			rawValue: func() any {
				v := int32(66)
				return &v
			}(),
			assertFunc: func(t *testing.T, target Target) {
				assert.Equal(t, int64(66), target.Int64Value)
			},
		},
		{
			name:      "SameTypeAssign_PointerField",
			fieldName: "PointerInt",
			rawValue: func() any {
				v := 5
				return &v
			}(),
			assertFunc: func(t *testing.T, target Target) {
				require.NotNil(t, target.PointerInt)
				assert.Equal(t, 5, *target.PointerInt)
			},
		},
		{
			name:      "SameTypeAssign_NonPointer",
			fieldName: "IntValue",
			rawValue:  6,
			assertFunc: func(t *testing.T, target Target) {
				assert.Equal(t, 6, target.IntValue)
			},
		},
		{
			name:        "Unsupported_Nil_To_NonPointer",
			fieldName:   "IntValue",
			rawValue:    nil,
			expectError: true,
			errorCheck: func(t *testing.T, err error) {
				var unsupported typecast.ErrUnsupportedType
				require.True(t, errors.As(err, &unsupported))
			},
		},
		{
			name:        "Unsupported_NonConvertible_To_Pointer",
			fieldName:   "PointerInt",
			rawValue:    struct{}{},
			expectError: true,
			errorCheck: func(t *testing.T, err error) {
				var unsupported typecast.ErrUnsupportedType
				require.True(t, errors.As(err, &unsupported))
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			var target Target
			field := reflect.ValueOf(&target).Elem().FieldByName(testCase.fieldName)
			err := AssignFromAny(caster, field, testCase.rawValue)
			if testCase.expectError {
				require.Error(t, err)
				if testCase.errorCheck != nil {
					testCase.errorCheck(t, err)
				}
				return
			}
			require.NoError(t, err)
			if testCase.assertFunc != nil {
				testCase.assertFunc(t, target)
			}
		})
	}
}

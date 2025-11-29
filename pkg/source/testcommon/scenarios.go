package testcommon

import (
	"errors"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Sufir/go-set-me-up/internal/typecast"
	"github.com/Sufir/go-set-me-up/pkg"
)

type DataEntry struct {
	Value string
	Path  []string
}

type Scenario struct {
	CreateConfig func() any
	PreInit      func(any)
	AssertResult func(*testing.T, any)
	AssertError  func(*testing.T, error)
	Name         string
	Input        []DataEntry
	Mode         pkg.LoadMode
}

func RunScenario(t *testing.T, scenario Scenario, executeFunction func(*testing.T, any, pkg.LoadMode, []DataEntry) error) {
	configuration := scenario.CreateConfig()
	if scenario.PreInit != nil {
		scenario.PreInit(configuration)
	}
	loadError := executeFunction(t, configuration, scenario.Mode, scenario.Input)
	if scenario.AssertError != nil {
		scenario.AssertError(t, loadError)
	} else {
		require.NoError(t, loadError)
	}
	scenario.AssertResult(t, configuration)
}

type BasicTypesConfiguration struct {
	Name  string `env:"NAME" flag:"name" flagShort:"n"`
	Port  int    `env:"PORT" flag:"port" flagShort:"p"`
	Debug bool   `env:"DEBUG" flag:"debug" flagShort:"d"`
}

type CastConfiguration struct {
	IntPointer   *int       `env:"INT_POINTER" flag:"int_pointer" flagShort:"ip"`
	ByteSlice    []byte     `env:"BYTE_SLICE" flag:"byte_slice" flagShort:"bs"`
	ComplexValue complex128 `env:"COMPLEX_VALUE" flag:"complex_value" flagShort:"cv"`
	IntValue     int        `env:"INT_VALUE" flag:"int_value" flagShort:"iv"`
	FloatValue   float64    `env:"FLOAT_VALUE" flag:"float_value" flagShort:"fv"`
	ByteArray    [5]byte    `env:"BYTE_ARRAY" flag:"byte_array" flagShort:"ba"`
	BoolValue    bool       `env:"BOOL_VALUE" flag:"bool_value" flagShort:"bv"`
}

type ModeBehaviorConfiguration struct {
	B *int `env:"B" flag:"b" flagShort:"B" flagDefault:"20"`
	A int  `env:"A" flag:"a" flagShort:"A" flagDefault:"10"`
}

type NestedInner struct {
	Value int `env:"VALUE" flag:"value" flagShort:"v"`
}

type NestedOuter struct {
	Inner NestedInner `envSegment:"inner"`
}

type RootNested struct {
	Outer NestedOuter `envSegment:"outer"`
}

type PNestedInner struct {
	Value int `env:"VALUE" flag:"value" flagShort:"v"`
}

type PNestedOuter struct {
	Inner *PNestedInner `envSegment:"inner"`
}

type PRootNested struct {
	Outer *PNestedOuter `envSegment:"outer"`
}

type PointerConfiguration struct {
	NumberPointer *int `env:"P" flag:"p" flagShort:"np"`
}

type CustomUnmarshaler struct {
	Value int
}

func (c *CustomUnmarshaler) UnmarshalText(text []byte) error {
	s := strings.TrimSpace(string(text))
	n, err := strconv.Atoi(s)
	if err != nil {
		return errors.New("invalid integer")
	}
	c.Value = n
	return nil
}

type TextUnmarshalerConfiguration struct {
	PointerType *CustomUnmarshaler `env:"U2" flag:"u2" flagShort:"u2s"`
	ValueType   CustomUnmarshaler  `env:"U1" flag:"u1" flagShort:"u1s"`
}

type AggregationConfiguration struct {
	B     []int `env:"B" flag:"b" flagShort:"B"`
	A     int   `env:"A" flag:"a" flagShort:"A"`
	C     int   `env:"C" flag:"c" flagShort:"C"`
	Outer struct {
		Inner struct {
			Value int `env:"VALUE" flag:"value" flagShort:"v"`
		} `envSegment:"inner"`
	} `envSegment:"outer"`
}

type EmptyValuesConfiguration struct {
	S string `env:"S" flag:"s" flagShort:"S"`
	I int    `env:"I" flag:"i" flagShort:"I"`
	B bool   `env:"B" flag:"b" flagShort:"B"`
}

func IntPointer(value int) *int {
	x := value
	return &x
}

func BuildBasicPrimitivesScenarios() []Scenario {
	return []Scenario{
		{
			Name:         "Basic_Primitives_From_String",
			Mode:         pkg.ModeOverride,
			CreateConfig: func() any { return &BasicTypesConfiguration{} },
			Input: []DataEntry{
				{Path: []string{"Port"}, Value: "8080"},
				{Path: []string{"Debug"}, Value: "true"},
				{Path: []string{"Name"}, Value: "  hello  "},
			},
			AssertResult: func(t *testing.T, configuration any) {
				configurationTyped := configuration.(*BasicTypesConfiguration)
				assert.Equal(t, 8080, configurationTyped.Port)
				assert.Equal(t, true, configurationTyped.Debug)
				assert.Equal(t, "hello", configurationTyped.Name)
			},
		},
		{
			Name:         "Float_And_Complex_From_String",
			Mode:         pkg.ModeOverride,
			CreateConfig: func() any { return &CastConfiguration{} },
			Input: []DataEntry{
				{Path: []string{"FloatValue"}, Value: "3.5"},
				{Path: []string{"ComplexValue"}, Value: "1+2i"},
			},
			AssertResult: func(t *testing.T, configuration any) {
				configurationTyped := configuration.(*CastConfiguration)
				assert.Equal(t, 3.5, configurationTyped.FloatValue)
				assert.Equal(t, complex(1, 2), configurationTyped.ComplexValue)
			},
		},
		{
			Name:         "Boolean_From_String",
			Mode:         pkg.ModeOverride,
			CreateConfig: func() any { return &CastConfiguration{} },
			Input: []DataEntry{
				{Path: []string{"BoolValue"}, Value: " true "},
			},
			AssertResult: func(t *testing.T, configuration any) {
				configurationTyped := configuration.(*CastConfiguration)
				assert.Equal(t, true, configurationTyped.BoolValue)
			},
		},
	}
}

func BuildPointerLeafScenarios() []Scenario {
	return []Scenario{
		{
			Name:         "Pointer_Int_From_String",
			Mode:         pkg.ModeOverride,
			CreateConfig: func() any { return &PointerConfiguration{} },
			Input: []DataEntry{
				{Path: []string{"NumberPointer"}, Value: "42"},
			},
			AssertResult: func(t *testing.T, configuration any) {
				configurationTyped := configuration.(*PointerConfiguration)
				require.NotNil(t, configurationTyped.NumberPointer)
				assert.Equal(t, 42, *configurationTyped.NumberPointer)
			},
		},
		{
			Name:         "Pointer_Field_In_Cast_Configuration",
			Mode:         pkg.ModeOverride,
			CreateConfig: func() any { return &CastConfiguration{} },
			Input: []DataEntry{
				{Path: []string{"IntPointer"}, Value: "100"},
			},
			AssertResult: func(t *testing.T, configuration any) {
				configurationTyped := configuration.(*CastConfiguration)
				require.NotNil(t, configurationTyped.IntPointer)
				assert.Equal(t, 100, *configurationTyped.IntPointer)
			},
		},
	}
}

func BuildBytesScenarios() []Scenario {
	return []Scenario{
		{
			Name:         "Bytes_Slice_And_Array_From_String",
			Mode:         pkg.ModeOverride,
			CreateConfig: func() any { return &CastConfiguration{} },
			Input: []DataEntry{
				{Path: []string{"ByteSlice"}, Value: " a b "},
				{Path: []string{"ByteArray"}, Value: "abc"},
			},
			AssertResult: func(t *testing.T, configuration any) {
				configurationTyped := configuration.(*CastConfiguration)
				assert.Equal(t, []byte(" a b "), configurationTyped.ByteSlice)
				assert.Equal(t, [5]byte{'a', 'b', 'c', 0, 0}, configurationTyped.ByteArray)
			},
		},
	}
}

func BuildNestedValueScenarios() []Scenario {
	return []Scenario{
		{
			Name:         "Nested_Value_Structs",
			Mode:         pkg.ModeOverride,
			CreateConfig: func() any { return &RootNested{} },
			Input: []DataEntry{
				{Path: []string{"Outer", "Inner", "Value"}, Value: "123"},
			},
			AssertResult: func(t *testing.T, configuration any) {
				configurationTyped := configuration.(*RootNested)
				assert.Equal(t, 123, configurationTyped.Outer.Inner.Value)
			},
		},
	}
}

func BuildNestedPointerScenarios() []Scenario {
	return []Scenario{
		{
			Name:         "Nested_Pointer_Structs",
			Mode:         pkg.ModeOverride,
			CreateConfig: func() any { return &PRootNested{} },
			Input: []DataEntry{
				{Path: []string{"Outer", "Inner", "Value"}, Value: "321"},
			},
			AssertResult: func(t *testing.T, configuration any) {
				configurationTyped := configuration.(*PRootNested)
				require.NotNil(t, configurationTyped.Outer)
				require.NotNil(t, configurationTyped.Outer.Inner)
				assert.Equal(t, 321, configurationTyped.Outer.Inner.Value)
			},
		},
	}
}

func BuildModeScenarios() []Scenario {
	return []Scenario{
		{
			Name:         "Mode_Override_Replaces_Values",
			Mode:         pkg.ModeOverride,
			CreateConfig: func() any { return &ModeBehaviorConfiguration{} },
			PreInit: func(configuration any) {
				configurationTyped := configuration.(*ModeBehaviorConfiguration)
				configurationTyped.A = 5
				configurationTyped.B = IntPointer(7)
			},
			Input: []DataEntry{
				{Path: []string{"A"}, Value: "10"},
				{Path: []string{"B"}, Value: "20"},
			},
			AssertResult: func(t *testing.T, configuration any) {
				configurationTyped := configuration.(*ModeBehaviorConfiguration)
				require.NotNil(t, configurationTyped.B)
				assert.Equal(t, 10, configurationTyped.A)
				assert.Equal(t, 20, *configurationTyped.B)
			},
		},
		{
			Name:         "Mode_FillMissing_Does_Not_Override_NonZero",
			Mode:         pkg.ModeFillMissing,
			CreateConfig: func() any { return &ModeBehaviorConfiguration{} },
			PreInit: func(configuration any) {
				configurationTyped := configuration.(*ModeBehaviorConfiguration)
				configurationTyped.A = 5
				configurationTyped.B = IntPointer(7)
			},
			Input: []DataEntry{
				{Path: []string{"A"}, Value: "10"},
				{Path: []string{"B"}, Value: "20"},
			},
			AssertResult: func(t *testing.T, configuration any) {
				configurationTyped := configuration.(*ModeBehaviorConfiguration)
				require.NotNil(t, configurationTyped.B)
				assert.Equal(t, 5, configurationTyped.A)
				assert.Equal(t, 7, *configurationTyped.B)
			},
		},
		{
			Name:         "Mode_FillMissing_Fills_Zero_Values",
			Mode:         pkg.ModeFillMissing,
			CreateConfig: func() any { return &ModeBehaviorConfiguration{} },
			Input: []DataEntry{
				{Path: []string{"A"}, Value: "10"},
				{Path: []string{"B"}, Value: "20"},
			},
			AssertResult: func(t *testing.T, configuration any) {
				configurationTyped := configuration.(*ModeBehaviorConfiguration)
				require.NotNil(t, configurationTyped.B)
				assert.Equal(t, 10, configurationTyped.A)
				assert.Equal(t, 20, *configurationTyped.B)
			},
		},
		{
			Name:         "Default_Mode_Zero_Is_Override",
			Mode:         0,
			CreateConfig: func() any { return &ModeBehaviorConfiguration{} },
			PreInit: func(configuration any) {
				configurationTyped := configuration.(*ModeBehaviorConfiguration)
				configurationTyped.A = 1
			},
			Input: []DataEntry{
				{Path: []string{"A"}, Value: "10"},
			},
			AssertResult: func(t *testing.T, configuration any) {
				configurationTyped := configuration.(*ModeBehaviorConfiguration)
				assert.Equal(t, 10, configurationTyped.A)
			},
		},
	}
}

func BuildAggregatedErrorScenarios() []Scenario {
	return []Scenario{
		{
			Name:         "Aggregated_Errors_Multiple_Fields",
			Mode:         pkg.ModeOverride,
			CreateConfig: func() any { return &AggregationConfiguration{} },
			Input: []DataEntry{
				{Path: []string{"A"}, Value: "x"},
				{Path: []string{"B"}, Value: "a,b,c"},
				{Path: []string{"C"}, Value: ""},
				{Path: []string{"Outer", "Inner", "Value"}, Value: "y"},
			},
			AssertError: func(t *testing.T, err error) {
				require.Error(t, err)
				var parseErr typecast.ErrParseFailed
				assert.True(t, errors.As(err, &parseErr))
			},
			AssertResult: func(_ *testing.T, cfg any) {
				_ = cfg
			},
		},
	}
}

func BuildUnknownKeysScenarios() []Scenario {
	return []Scenario{
		{
			Name:         "Unknown_Keys_Ignored",
			Mode:         pkg.ModeOverride,
			CreateConfig: func() any { return &BasicTypesConfiguration{} },
			Input:        []DataEntry{},
			AssertResult: func(t *testing.T, configuration any) {
				configurationTyped := configuration.(*BasicTypesConfiguration)
				assert.Equal(t, 0, configurationTyped.Port)
				assert.Equal(t, false, configurationTyped.Debug)
				assert.Equal(t, "", configurationTyped.Name)
			},
		},
	}
}

func BuildEmptyValuesScenarios() []Scenario {
	return []Scenario{
		{
			Name:         "Empty_Values_From_String",
			Mode:         pkg.ModeOverride,
			CreateConfig: func() any { return &EmptyValuesConfiguration{} },
			Input: []DataEntry{
				{Path: []string{"I"}, Value: ""},
				{Path: []string{"B"}, Value: ""},
				{Path: []string{"S"}, Value: ""},
			},
			AssertError: func(t *testing.T, err error) {
				require.Error(t, err)
				var parseErr typecast.ErrParseFailed
				assert.True(t, errors.As(err, &parseErr))
			},
			AssertResult: func(t *testing.T, configuration any) {
				configurationTyped := configuration.(*EmptyValuesConfiguration)
				assert.Equal(t, "", configurationTyped.S)
			},
		},
	}
}

func BuildTextUnmarshalerScenarios() []Scenario {
	return []Scenario{
		{
			Name:         "Text_Unmarshaler_Value_And_Pointer",
			Mode:         pkg.ModeOverride,
			CreateConfig: func() any { return &TextUnmarshalerConfiguration{} },
			Input: []DataEntry{
				{Path: []string{"ValueType"}, Value: "100"},
				{Path: []string{"PointerType"}, Value: "200"},
			},
			AssertResult: func(t *testing.T, configuration any) {
				configurationTyped := configuration.(*TextUnmarshalerConfiguration)
				assert.Equal(t, 100, configurationTyped.ValueType.Value)
				require.NotNil(t, configurationTyped.PointerType)
				assert.Equal(t, 200, configurationTyped.PointerType.Value)
			},
		},
	}
}

func BuildInvalidPrimitiveCastScenarios() []Scenario {
	return []Scenario{
		{
			Name:         "Invalid_Int_From_String",
			Mode:         pkg.ModeOverride,
			CreateConfig: func() any { return &CastConfiguration{} },
			Input: []DataEntry{
				{Path: []string{"IntValue"}, Value: "x"},
			},
			AssertError: func(t *testing.T, err error) {
				require.Error(t, err)
				var parseErr typecast.ErrParseFailed
				assert.True(t, errors.As(err, &parseErr))
			},
			AssertResult: func(_ *testing.T, configuration any) {
				_ = configuration.(*CastConfiguration)
			},
		},
		{
			Name:         "Invalid_Bool_From_String",
			Mode:         pkg.ModeOverride,
			CreateConfig: func() any { return &CastConfiguration{} },
			Input: []DataEntry{
				{Path: []string{"BoolValue"}, Value: "yes"},
			},
			AssertError: func(t *testing.T, err error) {
				require.Error(t, err)
				var parseErr typecast.ErrParseFailed
				assert.True(t, errors.As(err, &parseErr))
			},
			AssertResult: func(_ *testing.T, configuration any) {
				_ = configuration.(*CastConfiguration)
			},
		},
		{
			Name:         "Invalid_Complex_From_String",
			Mode:         pkg.ModeOverride,
			CreateConfig: func() any { return &CastConfiguration{} },
			Input: []DataEntry{
				{Path: []string{"ComplexValue"}, Value: "x+yi"},
			},
			AssertError: func(t *testing.T, err error) {
				require.Error(t, err)
				var parseErr typecast.ErrParseFailed
				assert.True(t, errors.As(err, &parseErr))
			},
			AssertResult: func(_ *testing.T, configuration any) {
				_ = configuration.(*CastConfiguration)
			},
		},
	}
}

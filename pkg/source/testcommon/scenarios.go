package testcommon

import (
	"errors"
	"strconv"
	"strings"
	"testing"

	"github.com/Sufir/go-set-me-up/internal/typecast"
	"github.com/Sufir/go-set-me-up/pkg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type DataEntry struct {
	Path  []string
	Value string
}

type Scenario struct {
	Name         string
	Mode         pkg.LoadMode
	CreateConfig func() any
	PreInit      func(any)
	Input        []DataEntry
	AssertResult func(*testing.T, any)
	AssertError  func(*testing.T, error)
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
	Port  int    `env:"PORT"`
	Debug bool   `env:"DEBUG"`
	Name  string `env:"NAME"`
}

type CastConfiguration struct {
	IntValue     int        `env:"INT_VALUE"`
	IntPointer   *int       `env:"INT_POINTER"`
	BoolValue    bool       `env:"BOOL_VALUE"`
	FloatValue   float64    `env:"FLOAT_VALUE"`
	ComplexValue complex128 `env:"COMPLEX_VALUE"`
	ByteSlice    []byte     `env:"BYTE_SLICE"`
	ByteArray    [5]byte    `env:"BYTE_ARRAY"`
}

type ModeBehaviorConfiguration struct {
	A int  `env:"A"`
	B *int `env:"B"`
}

type NestedInner struct {
	Value int `env:"VALUE"`
}

type NestedOuter struct {
	Inner NestedInner `envSegment:"inner"`
}

type RootNested struct {
	Outer NestedOuter `envSegment:"outer"`
}

type PNestedInner struct {
	Value int `env:"VALUE"`
}

type PNestedOuter struct {
	Inner *PNestedInner `envSegment:"inner"`
}

type PRootNested struct {
	Outer *PNestedOuter `envSegment:"outer"`
}

type PointerConfiguration struct {
	NumberPointer *int `env:"P"`
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
	ValueType   CustomUnmarshaler  `env:"U1"`
	PointerType *CustomUnmarshaler `env:"U2"`
}

type AggregationConfiguration struct {
	A     int   `env:"A"`
	B     []int `env:"B"`
	C     int   `env:"C"`
	Outer struct {
		Inner struct {
			Value int `env:"VALUE"`
		} `envSegment:"inner"`
	} `envSegment:"outer"`
}

type EmptyValuesConfiguration struct {
	I int    `env:"I"`
	B bool   `env:"B"`
	S string `env:"S"`
}

func In
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
			AssertResult: func(t *testing.T, cfg any) {
				c := cfg.(*BasicTypesConfiguration)
				assert.Equal(t, 8080, c.Port)
				assert.Equal(t, true, c.Debug)
				assert.Equal(t, "hello", c.Name)
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
			AssertResult: func(t *testing.T, cfg any) {
				c := cfg.(*CastConfiguration)
				assert.Equal(t, 3.5, c.FloatValue)
				assert.Equal(t, complex(1, 2), c.ComplexValue)
			},
		},
		{
			Name:         "Boolean_From_String",
			Mode:         pkg.ModeOverride,
			CreateConfig: func() any { return &CastConfiguration{} },
			Input: []DataEntry{
				{Path: []string{"BoolValue"}, Value: " true "},
			},
			AssertResult: func(t *testing.T, cfg any) {
				c := cfg.(*CastConfiguration)
				assert.Equal(t, true, c.BoolValue)
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
			AssertResult: func(t *testing.T, cfg any) {
				c := cfg.(*PointerConfiguration)
				require.NotNil(t, c.NumberPointer)
				assert.Equal(t, 42, *c.NumberPointer)
			},
		},
		{
			Name:         "Pointer_Field_In_Cast_Configuration",
			Mode:         pkg.ModeOverride,
			CreateConfig: func() any { return &CastConfiguration{} },
			Input: []DataEntry{
				{Path: []string{"IntPointer"}, Value: "100"},
			},
			AssertResult: func(t *testing.T, cfg any) {
				c := cfg.(*CastConfiguration)
				require.NotNil(t, c.IntPointer)
				assert.Equal(t, 100, *c.IntPointer)
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
			AssertResult: func(t *testing.T, cfg any) {
				c := cfg.(*CastConfiguration)
				assert.Equal(t, []byte(" a b "), c.ByteSlice)
				assert.Equal(t, [5]byte{'a', 'b', 'c', 0, 0}, c.ByteArray)
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
			AssertResult: func(t *testing.T, cfg any) {
				c := cfg.(*RootNested)
				assert.Equal(t, 123, c.Outer.Inner.Value)
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
			AssertResult: func(t *testing.T, cfg any) {
				c := cfg.(*PRootNested)
				require.NotNil(t, c.Outer)
				require.NotNil(t, c.Outer.Inner)
				assert.Equal(t, 321, c.Outer.Inner.Value)
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
			PreInit: func(a any) {
				c := a.(*ModeBehaviorConfiguration)
				c.A = 5
				c.B = IntPointer(7)
			},
			Input: []DataEntry{
				{Path: []string{"A"}, Value: "10"},
				{Path: []string{"B"}, Value: "20"},
			},
			AssertResult: func(t *testing.T, cfg any) {
				c := cfg.(*ModeBehaviorConfiguration)
				require.NotNil(t, c.B)
				assert.Equal(t, 10, c.A)
				assert.Equal(t, 20, *c.B)
			},
		},
		{
			Name:         "Mode_FillMissing_Does_Not_Override_NonZero",
			Mode:         pkg.ModeFillMissing,
			CreateConfig: func() any { return &ModeBehaviorConfiguration{} },
			PreInit: func(a any) {
				c := a.(*ModeBehaviorConfiguration)
				c.A = 5
				c.B = IntPointer(7)
			},
			Input: []DataEntry{
				{Path: []string{"A"}, Value: "10"},
				{Path: []string{"B"}, Value: "20"},
			},
			AssertResult: func(t *testing.T, cfg any) {
				c := cfg.(*ModeBehaviorConfiguration)
				require.NotNil(t, c.B)
				assert.Equal(t, 5, c.A)
				assert.Equal(t, 7, *c.B)
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
			AssertResult: func(t *testing.T, cfg any) {
				c := cfg.(*ModeBehaviorConfiguration)
				require.NotNil(t, c.B)
				assert.Equal(t, 10, c.A)
				assert.Equal(t, 20, *c.B)
			},
		},
		{
			Name:         "Default_Mode_Zero_Is_Override",
			Mode:         0,
			CreateConfig: func() any { return &ModeBehaviorConfiguration{} },
			PreInit: func(a any) {
				c := a.(*ModeBehaviorConfiguration)
				c.A = 1
			},
			Input: []DataEntry{
				{Path: []string{"A"}, Value: "10"},
			},
			AssertResult: func(t *testing.T, cfg any) {
				c := cfg.(*ModeBehaviorConfiguration)
				assert.Equal(t, 10, c.A)
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
				{Path: []string{"B"}, Value: "1,2,3"},
				{Path: []string{"C"}, Value: ""},
				{Path: []string{"Outer", "Inner", "Value"}, Value: "y"},
			},
			AssertError: func(t *testing.T, err error) {
				require.Error(t, err)
				var parseErr typecast.ErrParseFailed
				var unsupportedErr typecast.ErrUnsupportedType
				assert.True(t, errors.As(err, &parseErr))
				assert.True(t, errors.As(err, &unsupportedErr))
			},
			AssertResult: func(t *testing.T, cfg any) {
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
			AssertResult: func(t *testing.T, cfg any) {
				c := cfg.(*BasicTypesConfiguration)
				assert.Equal(t, 0, c.Port)
				assert.Equal(t, false, c.Debug)
				assert.Equal(t, "", c.Name)
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
			AssertResult: func(t *testing.T, cfg any) {
				c := cfg.(*TextUnmarshalerConfiguration)
				assert.Equal(t, 100, c.ValueType.Value)
				require.NotNil(t, c.PointerType)
				assert.Equal(t, 200, c.PointerType.Value)
			},
		},
	}
}

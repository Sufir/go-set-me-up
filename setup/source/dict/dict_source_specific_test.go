package dict

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Sufir/go-set-me-up/setup"
)

func intPointer(v int) *int {
	x := v
	return &x
}

type KeyResolutionConfig struct {
	SomeVar int
}

func TestDictSource_KeyResolution_Table(t *testing.T) {
	testCases := []struct {
		input    map[string]any
		name     string
		expected int
	}{
		{name: "FieldNamePriority", input: map[string]any{"SomeVar": 1, "some_var": 9, "SOME_VAR": 9}, expected: 1},
		{name: "SnakeLowerUsed", input: map[string]any{"some_var": 2}, expected: 2},
		{name: "SnakeUpperUsed", input: map[string]any{"SOME_VAR": 3}, expected: 3},
		{name: "NoKeyNoChange", input: map[string]any{"OTHER": 7}, expected: 0},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			source := NewSource(tc.input, setup.ModeOverride)
			cfg := KeyResolutionConfig{}
			err := source.Load(&cfg)
			require.NoError(t, err)
			assert.Equal(t, tc.expected, cfg.SomeVar)
		})
	}
}

type CollectionsConfig struct {
	Ints  []int
	Bytes [3]byte
}

func TestDictSource_Collections_StringParsing_Supported(t *testing.T) {
	sourceGood := NewSource(map[string]any{
		"Ints":  []int{1, 2, 3},
		"Bytes": "xyz",
	}, setup.ModeOverride)
	cfgGood := CollectionsConfig{}
	err := sourceGood.Load(&cfgGood)
	require.NoError(t, err)
	assert.Equal(t, []int{1, 2, 3}, cfgGood.Ints)
	assert.Equal(t, [3]byte{'x', 'y', 'z'}, cfgGood.Bytes)

	sourceParsed := NewSource(map[string]any{
		"Ints": "1,2,3",
	}, setup.ModeOverride)
	cfgParsed := CollectionsConfig{}
	err = sourceParsed.Load(&cfgParsed)
	require.NoError(t, err)
	assert.Equal(t, []int{1, 2, 3}, cfgParsed.Ints)
}

func TestDictSource_ConvertibleNumericTypes(t *testing.T) {
	type ConvertibleConfig struct {
		I int
		U uint64
		F float64
	}
	source := NewSource(map[string]any{
		"I": float64(3.5),
		"U": int64(7),
		"F": int32(2),
	}, setup.ModeOverride)
	cfg := ConvertibleConfig{}
	err := source.Load(&cfg)
	require.NoError(t, err)
	assert.Equal(t, 3, cfg.I)
	assert.Equal(t, uint64(7), cfg.U)
	assert.InDelta(t, 2.0, cfg.F, 1e-9)
}

type NilAssignConfig struct {
	MP map[string]int
	PP *int
	SP []int
	VP int
}

func TestDictSource_NilAssignments(t *testing.T) {
	sourceOverride := NewSource(map[string]any{"SP": nil, "MP": nil, "VP": nil, "PP": nil}, setup.ModeOverride)
	cfgOverride := NilAssignConfig{SP: []int{1}, MP: map[string]int{"x": 1}, VP: 5, PP: intPointer(1)}
	err := sourceOverride.Load(&cfgOverride)
	require.Error(t, err)
	assert.Nil(t, cfgOverride.SP)
	assert.Nil(t, cfgOverride.MP)
	assert.Nil(t, cfgOverride.PP)

	sourceFill := NewSource(map[string]any{"SP": nil, "MP": nil, "PP": nil}, setup.ModeFillMissing)
	cfgFill := NilAssignConfig{SP: []int{1}, MP: map[string]int{"x": 1}, PP: intPointer(2)}
	err = sourceFill.Load(&cfgFill)
	require.NoError(t, err)
	assert.NotNil(t, cfgFill.SP)
	assert.NotNil(t, cfgFill.MP)
	assert.NotNil(t, cfgFill.PP)
}

func TestDictSource_NilForNonNilType_IsErrorWithFieldInfo(t *testing.T) {
	source := NewSource(map[string]any{"X": nil}, setup.ModeOverride)
	type Simple struct{ X int }
	cfg := Simple{}
	err := source.Load(&cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "dict field X: unsupported type int")
}

func TestDictSource_MapToStruct_NonStructField_Ignored(t *testing.T) {
	source := NewSource(map[string]any{"X": map[string]any{"A": 1}}, setup.ModeOverride)
	type Simple struct{ X int }
	cfg := Simple{}
	err := source.Load(&cfg)
	require.NoError(t, err)
	assert.Equal(t, 0, cfg.X)
}

type PointerMixConfig struct {
	IntPointer *int
	IntValue   int
}

func TestDictSource_PointerAutoWrapUnwrap(t *testing.T) {
	p := intPointer(77)
	source := NewSource(map[string]any{
		"IntValue":   p,
		"IntPointer": 88,
	}, setup.ModeOverride)
	cfg := PointerMixConfig{}
	err := source.Load(&cfg)
	require.NoError(t, err)
	require.NotNil(t, cfg.IntPointer)
	assert.Equal(t, 77, cfg.IntValue)
	assert.Equal(t, 88, *cfg.IntPointer)
}

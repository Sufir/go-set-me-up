package dict

import (
	"testing"

	"github.com/Sufir/go-set-me-up/pkg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
		name     string
		input    map[string]any
		expected int
	}{
		{"FieldNamePriority", map[string]any{"SomeVar": 1, "some_var": 9, "SOME_VAR": 9}, 1},
		{"SnakeLowerUsed", map[string]any{"some_var": 2}, 2},
		{"SnakeUpperUsed", map[string]any{"SOME_VAR": 3}, 3},
		{"NoKeyNoChange", map[string]any{"OTHER": 7}, 0},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			source := NewDictSource(tc.input)
			cfg := KeyResolutionConfig{}
			err := source.Load(&cfg, pkg.ModeOverride)
			require.NoError(t, err)
			assert.Equal(t, tc.expected, cfg.SomeVar)
		})
	}
}

type CollectionsConfig struct {
	Ints  []int
	Bytes [3]byte
}

func TestDictSource_Collections_AssignWhole_NoElementCast(t *testing.T) {
	sourceGood := NewDictSource(map[string]any{
		"Ints":  []int{1, 2, 3},
		"Bytes": "xyz",
	})
	cfgGood := CollectionsConfig{}
	err := sourceGood.Load(&cfgGood, pkg.ModeOverride)
	require.NoError(t, err)
	assert.Equal(t, []int{1, 2, 3}, cfgGood.Ints)
	assert.Equal(t, [3]byte{'x', 'y', 'z'}, cfgGood.Bytes)

	sourceBad := NewDictSource(map[string]any{
		"Ints": "1,2,3",
	})
	cfgBad := CollectionsConfig{}
	err = sourceBad.Load(&cfgBad, pkg.ModeOverride)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "field Ints")
}

func TestDictSource_ConvertibleNumericTypes(t *testing.T) {
	type ConvertibleConfig struct {
		I int
		U uint64
		F float64
	}
	source := NewDictSource(map[string]any{
		"I": float64(3.5),
		"U": int64(7),
		"F": int32(2),
	})
	cfg := ConvertibleConfig{}
	err := source.Load(&cfg, pkg.ModeOverride)
	require.NoError(t, err)
	assert.Equal(t, 3, cfg.I)
	assert.Equal(t, uint64(7), cfg.U)
	assert.InDelta(t, 2.0, cfg.F, 1e-9)
}

type NilAssignConfig struct {
	SP []int
	MP map[string]int
	VP int
	PP *int
}

func TestDictSource_NilAssignments(t *testing.T) {
	sourceOverride := NewDictSource(map[string]any{"SP": nil, "MP": nil, "VP": nil, "PP": nil})
	cfgOverride := NilAssignConfig{SP: []int{1}, MP: map[string]int{"x": 1}, VP: 5, PP: intPointer(1)}
	err := sourceOverride.Load(&cfgOverride, pkg.ModeOverride)
	require.Error(t, err)
	assert.Nil(t, cfgOverride.SP)
	assert.Nil(t, cfgOverride.MP)
	assert.Nil(t, cfgOverride.PP)

	sourceFill := NewDictSource(map[string]any{"SP": nil, "MP": nil, "PP": nil})
	cfgFill := NilAssignConfig{SP: []int{1}, MP: map[string]int{"x": 1}, PP: intPointer(2)}
	err = sourceFill.Load(&cfgFill, pkg.ModeFillMissing)
	require.NoError(t, err)
	assert.NotNil(t, cfgFill.SP)
	assert.NotNil(t, cfgFill.MP)
	assert.NotNil(t, cfgFill.PP)
}

func TestDictSource_NilForNonNilType_IsErrorWithFieldInfo(t *testing.T) {
	source := NewDictSource(map[string]any{"X": nil})
	type Simple struct{ X int }
	cfg := Simple{}
	err := source.Load(&cfg, pkg.ModeOverride)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "field X (type int)")
}

func TestDictSource_MapToStruct_NonStructField_Ignored(t *testing.T) {
	source := NewDictSource(map[string]any{"X": map[string]any{"A": 1}})
	type Simple struct{ X int }
	cfg := Simple{}
	err := source.Load(&cfg, pkg.ModeOverride)
	require.NoError(t, err)
	assert.Equal(t, 0, cfg.X)
}

type PointerMixConfig struct {
	IntValue   int
	IntPointer *int
}

func TestDictSource_PointerAutoWrapUnwrap(t *testing.T) {
	p := intPointer(77)
	source := NewDictSource(map[string]any{
		"IntValue":   p,
		"IntPointer": 88,
	})
	cfg := PointerMixConfig{}
	err := source.Load(&cfg, pkg.ModeOverride)
	require.NoError(t, err)
	require.NotNil(t, cfg.IntPointer)
	assert.Equal(t, 77, cfg.IntValue)
	assert.Equal(t, 88, *cfg.IntPointer)
}

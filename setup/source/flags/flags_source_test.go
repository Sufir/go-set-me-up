package flags

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Sufir/go-set-me-up/setup"
)

type Example struct {
	Value int `flag:"value"`
}

type ExampleConfig struct {
	Name            string `flag:"name"`
	Sub             Example
	AdditionalValue int  `flag:"additional_value" flagShort:"a"`
	Debug           bool `flag:"debug"`
	Neg             bool `flag:"neg"`
}

func TestFlagsSource_Primitives_Short_And_BoolForms(t *testing.T) {
	oldArgs := os.Args
	os.Args = append([]string{"app"}, "--value", "12", "-a=18", "--debug", "--no-neg", "--name=  hello  ")
	defer func() { os.Args = oldArgs }()

	cfg := ExampleConfig{}
	source := NewSource(setup.ModeOverride)
	err := source.Load(&cfg)
	require.NoError(t, err)
	assert.Equal(t, 12, cfg.Sub.Value)
	assert.Equal(t, 18, cfg.AdditionalValue)
	assert.Equal(t, true, cfg.Debug)
	assert.Equal(t, false, cfg.Neg)
	assert.Equal(t, "hello", cfg.Name)
}

type ModeConfig struct {
	B *int `flag:"b" flagDefault:"20"`
	A int  `flag:"a" flagDefault:"10"`
}

func TestFlagsSource_Default_And_ModeFillMissing(t *testing.T) {
	oldArgs := os.Args
	os.Args = append([]string{"app"}, "--unused", "42")
	defer func() { os.Args = oldArgs }()

	cfg := ModeConfig{}
	source := NewSource(setup.ModeOverride)
	err := source.Load(&cfg)
	require.NoError(t, err)
	require.NotNil(t, cfg.B)
	assert.Equal(t, 10, cfg.A)
	assert.Equal(t, 20, *cfg.B)

	cfg2 := ModeConfig{A: 5}
	x := 7
	cfg2.B = &x
	os.Args = append([]string{"app"}, "--a", "30", "--b", "40")
	source = NewSource(setup.ModeFillMissing)
	err = source.Load(&cfg2)
	require.NoError(t, err)
	require.NotNil(t, cfg2.B)
	assert.Equal(t, 5, cfg2.A)
	assert.Equal(t, 7, *cfg2.B)
}

type CastConfig struct {
	ByteSlice []byte  `flag:"byte_slice"`
	ByteArray [5]byte `flag:"byte_array"`
}

func TestFlagsSource_ByteSlice_And_Array(t *testing.T) {
	oldArgs := os.Args
	os.Args = append([]string{"app"}, "--byte_slice= a b ", "--byte_array", "abc")
	defer func() { os.Args = oldArgs }()

	cfg := CastConfig{}
	source := NewSource(setup.ModeOverride)
	err := source.Load(&cfg)
	require.NoError(t, err)
	assert.Equal(t, []byte(" a b "), cfg.ByteSlice)
	assert.Equal(t, [5]byte{'a', 'b', 'c', 0, 0}, cfg.ByteArray)
}

func TestFlagsSource_BoolExplicitFalse(t *testing.T) {
	oldArgs := os.Args
	os.Args = append([]string{"app"}, "--debug=false")
	defer func() { os.Args = oldArgs }()

	type BoolConfig struct {
		Debug bool `flag:"debug"`
	}
	cfg := BoolConfig{}
	source := NewSource(setup.ModeOverride)
	err := source.Load(&cfg)
	require.NoError(t, err)
	assert.Equal(t, false, cfg.Debug)
}

package env

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Sufir/go-set-me-up/setup"
)

type EnvDelimiterConfig struct {
	Ints     []int    `env:"INTS"`
	Strings  []string `env:"STRS"`
	IntArray [3]int   `env:"ARR"`
}

type EnvTagDelimiterConfig struct {
	Ints    []int    `env:"INTS" envDelim:";"`
	Strings []string `env:"STRS" envDelim:"|"`
}

func TestEnvSource_Delimiter_DefaultFromSource(t *testing.T) {
	source := NewSource("app", ":", setup.ModeOverride)
	t.Setenv("APP_INTS", "1:2:3")
	t.Setenv("APP_ARR", "10:20:30")
	t.Setenv("APP_STRS", "a:b:c")
	cfg := &EnvDelimiterConfig{}
	err := source.Load(cfg)
	require.NoError(t, err)
	assert.Equal(t, []int{1, 2, 3}, cfg.Ints)
	assert.Equal(t, [3]int{10, 20, 30}, cfg.IntArray)
	assert.Equal(t, []string{"a", "b", "c"}, cfg.Strings)
}

func TestEnvSource_Delimiter_TagOverridesSource(t *testing.T) {
	source := NewSource("app", ":", setup.ModeOverride)
	t.Setenv("APP_INTS", "1;2;3")
	t.Setenv("APP_STRS", "a|b|c")
	cfg := &EnvTagDelimiterConfig{}
	err := source.Load(cfg)
	require.NoError(t, err)
	assert.Equal(t, []int{1, 2, 3}, cfg.Ints)
	assert.Equal(t, []string{"a", "b", "c"}, cfg.Strings)
}

package flags

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Sufir/go-set-me-up/setup"
)

type FlagsDelimiterConfig struct {
	Ints     []int    `flag:"ints"`
	Strings  []string `flag:"strs"`
	IntArray [3]int   `flag:"arr"`
}

type FlagsTagDelimiterConfig struct {
	Ints    []int    `flag:"ints" flagDelim:";"`
	Strings []string `flag:"strs" flagDelim:"|"`
}

func TestFlagsSource_Delimiter_DefaultFromSource(t *testing.T) {
	cfg := &FlagsDelimiterConfig{}
	args := []string{"app", "--ints", "1:2:3", "--arr", "10:20:30", "--strs", "a:b:c"}
	old := osArgsSwap(args)
	defer osArgsSwap(old)
	source := NewSourceWithDelimiter(setup.ModeOverride, ":")
	err := source.Load(cfg)
	require.NoError(t, err)
	assert.Equal(t, []int{1, 2, 3}, cfg.Ints)
	assert.Equal(t, [3]int{10, 20, 30}, cfg.IntArray)
	assert.Equal(t, []string{"a", "b", "c"}, cfg.Strings)
}

func TestFlagsSource_Delimiter_TagOverridesSource(t *testing.T) {
	cfg := &FlagsTagDelimiterConfig{}
	args := []string{"app", "--ints", "1;2;3", "--strs", "a|b|c"}
	old := osArgsSwap(args)
	defer osArgsSwap(old)
	source := NewSourceWithDelimiter(setup.ModeOverride, ":")
	err := source.Load(cfg)
	require.NoError(t, err)
	assert.Equal(t, []int{1, 2, 3}, cfg.Ints)
	assert.Equal(t, []string{"a", "b", "c"}, cfg.Strings)
}

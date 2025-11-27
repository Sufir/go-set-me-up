package env

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Sufir/go-set-me-up/pkg"
)

type SkipConfig struct {
	Value int `env:"-"`
}

type UnusedConfig struct {
	NoTag int
}

func TestEnvSource_Load_SkipFieldWithDashTag(t *testing.T) {
	source := NewSource("app", ",")

	t.Setenv("APP_VALUE", "123")

	cfg := SkipConfig{}
	err := source.Load(&cfg, pkg.ModeOverride)
	require.NoError(t, err)
	assert.Equal(t, 0, cfg.Value)
}

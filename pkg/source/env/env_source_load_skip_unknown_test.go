package env

import (
	"testing"

	"github.com/Sufir/go-set-me-up/pkg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type SkipConfig struct {
	Value int `env:"-"`
}

type UnusedConfig struct {
	NoTag int
}

func TestEnvSource_Load_SkipFieldWithDashTag(t *testing.T) {
	source := NewEnvSource("app", ",")

	t.Setenv("APP_VALUE", "123")

	cfg := SkipConfig{}
	err := source.Load(&cfg, pkg.ModeOverride)
	require.NoError(t, err)
	assert.Equal(t, 0, cfg.Value)
}

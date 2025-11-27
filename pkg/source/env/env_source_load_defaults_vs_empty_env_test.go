package env

import (
	"errors"
	"testing"

	"github.com/Sufir/go-set-me-up/internal/typecast"
	"github.com/Sufir/go-set-me-up/pkg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type DefaultsConflictConfig struct {
	S string `env:"S" envDefault:"fallback"`
	I int    `env:"I" envDefault:"7"`
	B bool   `env:"B" envDefault:"true"`
}

func TestEnvSource_Load_Override_EmptyEnvWinsOverDefault(t *testing.T) {
	source := NewEnvSource("app", ",")

	t.Setenv("APP_S", "")
	t.Setenv("APP_I", "")
	t.Setenv("APP_B", "")

	cfg := DefaultsConflictConfig{}
	err := source.Load(&cfg, pkg.ModeOverride)
	require.Error(t, err)

	var parseErr typecast.ErrParseFailed
	require.True(t, errors.As(err, &parseErr))
	assert.Equal(t, "", cfg.S)
	assert.Equal(t, 0, cfg.I)
	assert.Equal(t, false, cfg.B)
}

func TestEnvSource_Load_FillMissing_EmptyEnvWinsOverDefault(t *testing.T) {
	source := NewEnvSource("app", ",")

	t.Setenv("APP_S", "")
	t.Setenv("APP_I", "")
	t.Setenv("APP_B", "")

	cfg := DefaultsConflictConfig{}
	err := source.Load(&cfg, pkg.ModeFillMissing)
	require.Error(t, err)

	var parseErr typecast.ErrParseFailed
	require.True(t, errors.As(err, &parseErr))
	assert.Equal(t, "", cfg.S)
	assert.Equal(t, 0, cfg.I)
	assert.Equal(t, false, cfg.B)
}

func TestEnvSource_Load_DefaultsUsedWhenEnvMissing(t *testing.T) {
	source := NewEnvSource("app", ",")

	cfg := DefaultsConflictConfig{}
	err := source.Load(&cfg, pkg.ModeOverride)
	require.NoError(t, err)
	assert.Equal(t, "fallback", cfg.S)
	assert.Equal(t, 7, cfg.I)
	assert.Equal(t, true, cfg.B)
}

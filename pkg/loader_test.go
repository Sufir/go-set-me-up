package pkg_test

import (
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	pkg "github.com/Sufir/go-set-me-up/pkg"
	"github.com/Sufir/go-set-me-up/pkg/source/dict"
	"github.com/Sufir/go-set-me-up/pkg/source/env"
	"github.com/Sufir/go-set-me-up/pkg/source/flags"
	jsonfile "github.com/Sufir/go-set-me-up/pkg/source/json-file"
	"github.com/Sufir/go-set-me-up/pkg/source/testcommon"
)

func TestLoader_EnvSource_DefaultModeOverride_AssignsValues(t *testing.T) {
	environmentSource := env.NewSource("app", ",", pkg.ModeOverride)
	loader := pkg.NewLoader(environmentSource)
	configuration := &testcommon.BasicTypesConfiguration{}

	t.Setenv("APP_PORT", "8080")
	t.Setenv("APP_DEBUG", "true")
	t.Setenv("APP_NAME", "  hello  ")

	loadError := loader.Load(configuration)
	require.NoError(t, loadError)
	assert.Equal(t, 8080, configuration.Port)
	assert.Equal(t, true, configuration.Debug)
	assert.Equal(t, "hello", configuration.Name)
}

func TestLoader_CombineDictAndFlags_ModeOverride_OrderAndAssignments(t *testing.T) {
	dictionarySource := dict.NewSource(map[string]any{
		"IntValue":     "5",
		"ComplexValue": "1+2i",
	}, pkg.ModeOverride)
	flagsSource := flags.NewSource(pkg.ModeOverride)
	loader := pkg.NewLoader(dictionarySource, flagsSource)
	configuration := &testcommon.CastConfiguration{}

	previousArgs := os.Args
	os.Args = append([]string{"app"}, "--int_value", "10", "--byte_slice", " a b ", "--byte_array", "abc")
	defer func() { os.Args = previousArgs }()

	loadError := loader.Load(configuration)
	require.NoError(t, loadError)
	assert.Equal(t, 10, configuration.IntValue)
	assert.Equal(t, []byte(" a b "), configuration.ByteSlice)
	assert.Equal(t, [5]byte{'a', 'b', 'c', 0, 0}, configuration.ByteArray)
	assert.Equal(t, complex(1, 2), configuration.ComplexValue)
}

func TestLoader_JSONAndEnv_Override_NestedAssignments(t *testing.T) {
	jsonContent := []byte(`{"Outer":{"Inner":{"Value":123}},"Port":8080}`)
	tempFile, tempFileError := os.CreateTemp("", "loader-json-*.json")
	require.NoError(t, tempFileError)
	tempFilePath := tempFile.Name()
	require.NoError(t, tempFile.Close())
	require.NoError(t, os.WriteFile(tempFilePath, jsonContent, 0o644))
	defer os.Remove(tempFilePath)

	jsonFileSource := jsonfile.NewSource(tempFilePath, pkg.ModeOverride)
	environmentSource := env.NewSource("app", ",", pkg.ModeOverride)
	loader := pkg.NewLoader(jsonFileSource, environmentSource)

	type IntegratedConfiguration struct {
		Name  string `env:"NAME" json:"Name"`
		Port  int    `env:"PORT" json:"Port"`
		Outer struct {
			Inner struct {
				Value int `env:"VALUE" json:"Value"`
			} `envSegment:"inner" json:"Inner"`
		} `envSegment:"outer" json:"Outer"`
	}

	configuration := &IntegratedConfiguration{}
	t.Setenv("APP_OUTER_INNER_VALUE", "321")

	loadError := loader.Load(configuration)
	require.NoError(t, loadError)
	assert.Equal(t, 8080, configuration.Port)
	assert.Equal(t, 321, configuration.Outer.Inner.Value)
}

func TestLoader_AggregatesErrors_FromRealSources(t *testing.T) {
	environmentSource := env.NewSource("app", ",", pkg.ModeOverride)
	dictionarySource := dict.NewSource(map[string]any{
		"BoolValue": "yes",
	}, pkg.ModeOverride)
	flagsSource := flags.NewSource(pkg.ModeOverride)
	loader := pkg.NewLoader(environmentSource, dictionarySource, flagsSource)
	configuration := &testcommon.CastConfiguration{}

	t.Setenv("APP_INT_VALUE", "x")

	previousArgs := os.Args
	os.Args = append([]string{"app"}, "--complex_value", "x+yi")
	defer func() { os.Args = previousArgs }()

	loadError := loader.Load(configuration)
	require.Error(t, loadError)
	assert.True(t, errors.Is(loadError, pkg.ErrLoadAggregatedFailed))
	var aggregatedError *pkg.AggregatedLoadFailedError
	require.True(t, errors.As(loadError, &aggregatedError))
	var loaderSourceFailedError *pkg.LoaderSourceFailedError
	require.True(t, errors.As(loadError, &loaderSourceFailedError))
	errorMessage := loadError.Error()
	assert.Contains(t, errorMessage, "index 0")
	assert.Contains(t, errorMessage, "index 1")
	assert.Contains(t, errorMessage, "index 2")
	assert.Contains(t, errorMessage, "env.Source")
	assert.Contains(t, errorMessage, "dict.Source")
	assert.Contains(t, errorMessage, "flags.Source")
}

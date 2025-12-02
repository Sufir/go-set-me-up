package jsonfile

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Sufir/go-set-me-up/setup"
	"github.com/Sufir/go-set-me-up/setup/source/testcommon"
)

func writeJSONFile(t *testing.T, data any) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	bytes, err := json.Marshal(data)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(path, bytes, 0o600))
	return path
}

func buildJSONMapFromInput(configuration any, input []testcommon.DataEntry) map[string]any {
	root := map[string]any{}
	for _, entry := range input {
		current := root
		cfgVal := reflect.ValueOf(configuration)
		if cfgVal.Kind() == reflect.Ptr {
			cfgVal = cfgVal.Elem()
		}
		cfgType := cfgVal.Type()
		for i := 0; i < len(entry.Path)-1; i++ {
			name := entry.Path[i]
			field, ok := cfgType.FieldByName(name)
			if ok {
				tag := field.Tag.Get("json")
				if idx := strings.IndexByte(tag, ','); idx >= 0 {
					tag = tag[:idx]
				}
				if tag == "" || tag == "-" {
					tag = name
				}
				next, exists := current[tag]
				if !exists {
					next = map[string]any{}
					current[tag] = next
				}
				current = next.(map[string]any)
				ft := field.Type
				if ft.Kind() == reflect.Ptr {
					ft = ft.Elem()
				}
				cfgType = ft
				continue
			}
			next, exists := current[name]
			if !exists {
				next = map[string]any{}
				current[name] = next
			}
			current = next.(map[string]any)
		}
		leaf := entry.Path[len(entry.Path)-1]
		field, ok := cfgType.FieldByName(leaf)
		key := leaf
		if ok {
			tag := field.Tag.Get("json")
			if idx := strings.IndexByte(tag, ','); idx >= 0 {
				tag = tag[:idx]
			}
			if tag != "" && tag != "-" {
				key = tag
			}
			t := field.Type
			if t.Kind() == reflect.Ptr {
				t = t.Elem()
			}
			sv := strings.TrimSpace(entry.Value)
			switch t.Kind() {
			case reflect.String:
				current[key] = sv
			case reflect.Bool:
				if v, err := strconv.ParseBool(sv); err == nil {
					current[key] = v
				} else {
					current[key] = sv
				}
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				if v, err := strconv.ParseInt(sv, 10, 64); err == nil {
					current[key] = int(v)
				} else {
					current[key] = sv
				}
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
				if v, err := strconv.ParseUint(sv, 10, 64); err == nil {
					current[key] = uint(v)
				} else {
					current[key] = sv
				}
			case reflect.Float32, reflect.Float64:
				if v, err := strconv.ParseFloat(sv, 64); err == nil {
					current[key] = v
				} else {
					current[key] = sv
				}
			default:
				current[key] = sv
			}
			continue
		}
		current[key] = strings.TrimSpace(entry.Value)
	}
	return root
}

func executeJSONScenario(t *testing.T, configuration any, mode setup.LoadMode, input []testcommon.DataEntry) error {
	root := buildJSONMapFromInput(configuration, input)
	path := writeJSONFile(t, root)
	source := NewSource(path, mode)
	return source.Load(configuration)
}

func TestJSON_BasicPrimitives(t *testing.T) {
	type C struct {
		Name  string `json:"Name"`
		Port  int    `json:"Port"`
		Debug bool   `json:"Debug"`
	}
	err := executeJSONScenario(t, func() any { return &C{} }(), setup.ModeOverride, []testcommon.DataEntry{
		{Path: []string{"Port"}, Value: "8080"},
		{Path: []string{"Debug"}, Value: "true"},
		{Path: []string{"Name"}, Value: "  hello  "},
	})
	require.NoError(t, err)
}

func TestJSON_PointerLeaf(t *testing.T) {
	type C struct {
		IntPointer *int `json:"int_pointer"`
	}
	root := map[string]any{"int_pointer": 100}
	path := writeJSONFile(t, root)
	cfg := &C{}
	err := NewSource(path, setup.ModeOverride).Load(cfg)
	require.NoError(t, err)
	require.NotNil(t, cfg.IntPointer)
	assert.Equal(t, 100, *cfg.IntPointer)
}

func TestJSON_Bytes(t *testing.T) {
	type C struct {
		ByteSlice []byte  `json:"byte_slice"`
		ByteArray [5]byte `json:"byte_array"`
	}
	b64 := base64.StdEncoding.EncodeToString([]byte(" a b "))
	root := map[string]any{
		"byte_slice": b64,
		"byte_array": []int{'a', 'b', 'c', 0, 0},
	}
	path := writeJSONFile(t, root)
	cfg := &C{}
	err := NewSource(path, setup.ModeOverride).Load(cfg)
	require.NoError(t, err)
	assert.Equal(t, []byte(" a b "), cfg.ByteSlice)
	assert.Equal(t, [5]byte{'a', 'b', 'c', 0, 0}, cfg.ByteArray)
}

func TestJSON_NestedValue(t *testing.T) {
	type Inner struct {
		Value int `json:"Value"`
	}
	type Outer struct {
		Inner Inner `json:"Inner"`
	}
	type Root struct {
		Outer Outer `json:"Outer"`
	}
	cfg := &Root{}
	err := executeJSONScenario(t, cfg, setup.ModeOverride, []testcommon.DataEntry{{Path: []string{"Outer", "Inner", "Value"}, Value: "123"}})
	require.NoError(t, err)
	assert.Equal(t, 123, cfg.Outer.Inner.Value)
}

func TestJSON_NestedPointer(t *testing.T) {
	type Inner struct {
		Value int `json:"Value"`
	}
	type Outer struct {
		Inner *Inner `json:"Inner"`
	}
	type Root struct {
		Outer *Outer `json:"Outer"`
	}
	cfg := &Root{}
	err := executeJSONScenario(t, cfg, setup.ModeOverride, []testcommon.DataEntry{{Path: []string{"Outer", "Inner", "Value"}, Value: "321"}})
	require.NoError(t, err)
	require.NotNil(t, cfg.Outer)
	require.NotNil(t, cfg.Outer.Inner)
	assert.Equal(t, 321, cfg.Outer.Inner.Value)
}

func TestJSON_Mode(t *testing.T) {
	type C struct {
		B *int `json:"B"`
		A int  `json:"A"`
	}
	cfg := &C{A: 5, B: func(x int) *int { return &x }(7)}
	require.NoError(t, executeJSONScenario(t, cfg, setup.ModeOverride, []testcommon.DataEntry{{Path: []string{"A"}, Value: "10"}, {Path: []string{"B"}, Value: "20"}}))
	require.NotNil(t, cfg.B)
	assert.Equal(t, 10, cfg.A)
	assert.Equal(t, 20, *cfg.B)

	cfg2 := &C{A: 5, B: func(x int) *int { return &x }(7)}
	require.NoError(t, executeJSONScenario(t, cfg2, setup.ModeFillMissing, []testcommon.DataEntry{{Path: []string{"A"}, Value: "10"}, {Path: []string{"B"}, Value: "20"}}))
	require.NotNil(t, cfg2.B)
	assert.Equal(t, 5, cfg2.A)
	assert.Equal(t, 7, *cfg2.B)

	cfg3 := &C{}
	require.NoError(t, executeJSONScenario(t, cfg3, setup.ModeFillMissing, []testcommon.DataEntry{{Path: []string{"A"}, Value: "10"}, {Path: []string{"B"}, Value: "20"}}))
	require.NotNil(t, cfg3.B)
	assert.Equal(t, 10, cfg3.A)
	assert.Equal(t, 20, *cfg3.B)
}

func TestJSON_AggregatedErrors(t *testing.T) {
	type Root struct {
		B     []int `json:"b"`
		A     int   `json:"a"`
		C     int   `json:"c"`
		Outer struct {
			Inner struct {
				Value int `json:"value"`
			} `json:"inner"`
		} `json:"outer"`
	}
	root := map[string]any{
		"a":     "x",
		"b":     "1,2,3",
		"c":     nil,
		"outer": map[string]any{"inner": map[string]any{"value": "y"}},
	}
	path := writeJSONFile(t, root)
	cfg := &Root{}
	err := NewSource(path, setup.ModeOverride).Load(cfg)
	require.Error(t, err)
}

func TestJSON_UnknownKeysIgnored(t *testing.T) {
	type C struct {
		Name  string `json:"Name"`
		Port  int    `json:"Port"`
		Debug bool   `json:"Debug"`
	}
	cfg4 := &C{}
	require.NoError(t, executeJSONScenario(t, cfg4, setup.ModeOverride, []testcommon.DataEntry{}))
	assert.Equal(t, 0, cfg4.Port)
	assert.Equal(t, false, cfg4.Debug)
	assert.Equal(t, "", cfg4.Name)
}

package env

import (
	"reflect"
	"testing"

	"github.com/Sufir/go-set-me-up/pkg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConvertToEnvironmentVariableName(t *testing.T) {
	tests := []struct {
		name string
		in   string
		out  string
	}{
		{"simple_lower", "simple", "SIMPLE"},
		{"simple_upper", "SIMPLE", "SIMPLE"},
		{"camelCase", "fooBar", "FOO_BAR"},
		{"PascalCase", "FooBar", "FOO_BAR"},
		{"HTTPServer", "HTTPServer", "HTTP_SERVER"},
		{"digits_inside", "lower123Upper", "LOWER123_UPPER"},
		{"start_digit", "123start", "123START"},
		{"only_digits", "123456", "123456"},
		{"dash", "user-id", "USER_ID"},
		{"space", "user id", "USER_ID"},
		{"mixed separators", "user-id name", "USER_ID_NAME"},
		{"multiple_nonalpha", "user---name##id", "USER_NAME_ID"},
		{"leading_separators", "  leading", "LEADING"},
		{"trailing_separators", "trailing---", "TRAILING"},
		{"surrounded", "---value---", "VALUE"},
		{"double__underscore", "multiple__sep", "MULTIPLE_SEP"},
		{"empty", "", ""},
		{"only_symbols", "-_-", ""},
		{"spaces_only", "     ", ""},
		{"unicode_ignore", "пример", "ПРИМЕР"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.out, convertToEnvVar(tc.in))
		})
	}
}

func TestBuildKey(t *testing.T) {
	assert.Equal(t, "FOO", buildKey(nil, "FOO"))
	assert.Equal(t, "APP", buildKey([]string{"APP"}, ""))
	assert.Equal(t, "APP_SERVER_PORT", buildKey([]string{"APP", "SERVER"}, "PORT"))
}

func TestSplitIntoTokens(t *testing.T) {
	assert.Equal(t, []string{"1", "2", "3"}, splitIntoTokens("1,2,3", ","))
	assert.Equal(t, []string{"abc"}, splitIntoTokens("abc", ""))
}

func TestShouldSetField(t *testing.T) {
	var x int
	vx := reflect.ValueOf(&x).Elem()
	assert.True(t, (EnvSource{}).shouldSetField(vx, true, pkg.ModeOverride, ""))
	assert.True(t, (EnvSource{}).shouldSetField(vx, false, pkg.ModeOverride, "10"))
	assert.False(t, (EnvSource{}).shouldSetField(vx, false, pkg.ModeOverride, ""))
	x = 5
	vx = reflect.ValueOf(&x).Elem()
	assert.False(t, (EnvSource{}).shouldSetField(vx, true, pkg.ModeFillMissing, ""))
	x = 0
	vx = reflect.ValueOf(&x).Elem()
	assert.True(t, (EnvSource{}).shouldSetField(vx, true, pkg.ModeFillMissing, ""))
	assert.True(t, (EnvSource{}).shouldSetField(vx, false, pkg.ModeFillMissing, "7"))
}

func TestSetFieldValue(t *testing.T) {
	source := NewEnvSource("app", ",")
	type Holder struct {
		IntSlice          []int  `envDelim:":"`
		IntArray          [3]int `envDelim:","`
		ByteSlice         []byte
		BytesArray        [5]byte `envDelim:","`
		BytesArrayNoDelim [5]byte `envDelim:","`
		IntValue          int
	}
	var h Holder
	hv := reflect.ValueOf(&h).Elem()
	ht := hv.Type()

	field := hv.FieldByName("IntSlice")
	fi, _ := ht.FieldByName("IntSlice")
	require.Error(t, source.setFieldValue(field, "1:2:3", fi, ""))

	field = hv.FieldByName("IntArray")
	fi, _ = ht.FieldByName("IntArray")
	require.Error(t, source.setFieldValue(field, "10,20,30,40", fi, ""))

	field = hv.FieldByName("ByteSlice")
	fi, _ = ht.FieldByName("ByteSlice")
	require.NoError(t, source.setFieldValue(field, " a b ", fi, ""))
	assert.Equal(t, []byte(" a b "), h.ByteSlice)

	field = hv.FieldByName("BytesArray")
	fi, _ = ht.FieldByName("BytesArray")
	require.NoError(t, source.setFieldValue(field, "abcde", fi, ""))
	assert.Equal(t, [5]byte{'a', 'b', 'c', 'd', 'e'}, h.BytesArray)

	field = hv.FieldByName("BytesArrayNoDelim")
	fi, _ = ht.FieldByName("BytesArrayNoDelim")
	require.NoError(t, source.setFieldValue(field, "abc", fi, ""))
	assert.Equal(t, [5]byte{'a', 'b', 'c', 0, 0}, h.BytesArrayNoDelim)

	field = hv.FieldByName("IntValue")
	fi, _ = ht.FieldByName("IntValue")
	require.NoError(t, source.setFieldValue(field, "42", fi, ""))
	assert.Equal(t, 42, h.IntValue)
}

func TestGetEnv_ReadsSetVariables(t *testing.T) {
	t.Setenv("APP_TEST_X", "A")
	t.Setenv("APP_TEST_Y", "")
	m := getEnv()
	assert.Equal(t, "A", m["APP_TEST_X"])
	assert.Equal(t, "", m["APP_TEST_Y"])
}

func TestLoadStruct_SetsOnlyTaggedFields(t *testing.T) {
	source := NewEnvSource("APP", ",")
	type Sub struct {
		Value int `env:"VALUE"`
	}
	type Root struct {
		Sub   Sub `envSegment:"sub"`
		Skip  int `env:"-"`
		NoTag int
	}
	var r Root
	env := map[string]string{
		"APP_SUB_VALUE": "123",
		"APP_SKIP":      "1",
		"APP_NOT_USED":  "9",
	}
	var errs []error
	source.loadStruct(reflect.ValueOf(&r).Elem(), []string{"APP"}, env, pkg.ModeOverride, &errs)
	require.Empty(t, errs)
	assert.Equal(t, 123, r.Sub.Value)
	assert.Equal(t, 0, r.Skip)
	assert.Equal(t, 0, r.NoTag)
}

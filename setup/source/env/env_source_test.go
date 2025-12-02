package env

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Sufir/go-set-me-up/setup"
	"github.com/Sufir/go-set-me-up/setup/source/sourceutil"
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
			require.Equal(t, tc.out, sourceutil.ConvertToEnvVar(tc.in))
		})
	}
}

func TestBuildKey(t *testing.T) {
	assert.Equal(t, "FOO", buildKey(nil, "FOO"))
	assert.Equal(t, "APP", buildKey([]string{"APP"}, ""))
	assert.Equal(t, "APP_SERVER_PORT", buildKey([]string{"APP", "SERVER"}, "PORT"))
}

func TestNormalizeDelimited(t *testing.T) {
	assert.Equal(t, "1,2,3", sourceutil.NormalizeDelimited("1,2,3", ","))
	assert.Equal(t, "abc", sourceutil.NormalizeDelimited("abc", ""))
}

func TestShouldSetField(t *testing.T) {
	var x int
	vx := reflect.ValueOf(&x).Elem()
	assert.True(t, sourceutil.ShouldAssign(vx, true, setup.ModeOverride, ""))
	assert.True(t, sourceutil.ShouldAssign(vx, false, setup.ModeOverride, "10"))
	assert.False(t, sourceutil.ShouldAssign(vx, false, setup.ModeOverride, ""))
	x = 5
	vx = reflect.ValueOf(&x).Elem()
	assert.False(t, sourceutil.ShouldAssign(vx, true, setup.ModeFillMissing, ""))
	x = 0
	vx = reflect.ValueOf(&x).Elem()
	assert.True(t, sourceutil.ShouldAssign(vx, true, setup.ModeFillMissing, ""))
	assert.True(t, sourceutil.ShouldAssign(vx, false, setup.ModeFillMissing, "7"))
}

func TestSetFieldValue(t *testing.T) {
	source := NewSource("app", ",", setup.ModeOverride)
	type Holder struct {
		IntSlice          []int `envDelim:":"`
		ByteSlice         []byte
		IntArray          [3]int `envDelim:","`
		IntValue          int
		BytesArray        [5]byte `envDelim:","`
		BytesArrayNoDelim [5]byte `envDelim:","`
	}
	var h Holder
	hv := reflect.ValueOf(&h).Elem()

	field := hv.FieldByName("IntSlice")

	require.NoError(t, sourceutil.AssignFromString(source.caster, field, sourceutil.NormalizeDelimited("1:2:3", ":")))
	assert.Equal(t, []int{1, 2, 3}, h.IntSlice)

	field = hv.FieldByName("IntArray")

	require.NoError(t, sourceutil.AssignFromString(source.caster, field, sourceutil.NormalizeDelimited("10,20,30,40", ",")))
	assert.Equal(t, [3]int{10, 20, 30}, h.IntArray)

	field = hv.FieldByName("ByteSlice")

	require.NoError(t, sourceutil.AssignFromString(source.caster, field, " a b "))
	assert.Equal(t, []byte(" a b "), h.ByteSlice)

	field = hv.FieldByName("BytesArray")

	require.NoError(t, sourceutil.AssignFromString(source.caster, field, "abcde"))
	assert.Equal(t, [5]byte{'a', 'b', 'c', 'd', 'e'}, h.BytesArray)

	field = hv.FieldByName("BytesArrayNoDelim")

	require.NoError(t, sourceutil.AssignFromString(source.caster, field, "abc"))
	assert.Equal(t, [5]byte{'a', 'b', 'c', 0, 0}, h.BytesArrayNoDelim)

	field = hv.FieldByName("IntValue")
	require.NoError(t, sourceutil.AssignFromString(source.caster, field, "42"))
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
	source := NewSource("APP", ",", setup.ModeOverride)
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
	source.loadStruct(reflect.ValueOf(&r).Elem(), []string{"APP"}, env, setup.ModeOverride, &errs, "")
	require.Empty(t, errs)
	assert.Equal(t, 123, r.Sub.Value)
	assert.Equal(t, 0, r.Skip)
	assert.Equal(t, 0, r.NoTag)
}

package typecast

import (
	"errors"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type PointerUnmarshaler struct {
	Value int
}

func (pointerUnmarshaler *PointerUnmarshaler) UnmarshalText(text []byte) error {
	stringValue := strings.TrimSpace(string(text))
	parsedValue, parseError := strconv.Atoi(stringValue)
	if parseError != nil {
		return errors.New("invalid integer")
	}
	pointerUnmarshaler.Value = parsedValue
	return nil
}

type ValueUnmarshaler struct {
	Data string
}

func (valueUnmarshaler ValueUnmarshaler) UnmarshalText(text []byte) error {
	_ = text
	return nil
}

type BadPointerUnmarshaler struct{}

func (badPointerUnmarshaler *BadPointerUnmarshaler) UnmarshalText(text []byte) error {
	return errors.New("failed to unmarshal")
}

func TestTextUnmarshalerOptionTypeCast_PointerImplements(t *testing.T) {
	optionType := TextUnmarshalerOptionType{}
	targetType := reflect.TypeOf(PointerUnmarshaler{})
	value, err := optionType.Cast(" 123 ", targetType)
	require.NoError(t, err)
	require.Equal(t, reflect.Ptr, value.Kind())
	obtained := value.Elem().Interface().(PointerUnmarshaler)
	assert.Equal(t, 123, obtained.Value)
}

func TestTextUnmarshalerOptionTypeCast_ValueImplements(t *testing.T) {
	optionType := TextUnmarshalerOptionType{}
	targetType := reflect.TypeOf(ValueUnmarshaler{})
	value, err := optionType.Cast(" abc ", targetType)
	require.NoError(t, err)
	require.Equal(t, reflect.Struct, value.Kind())
}

func TestTextUnmarshalerOptionTypeCast_Negative(t *testing.T) {
	optionType := TextUnmarshalerOptionType{}
	targetType := reflect.TypeOf(BadPointerUnmarshaler{})
	value, err := optionType.Cast("any", targetType)
	require.Error(t, err)
	assert.False(t, value.IsValid())
	var parseErr ErrParseFailed
	require.True(t, errors.As(err, &parseErr))
	assert.Equal(t, targetType, parseErr.Type)
	assert.Equal(t, "any", parseErr.Value)
	assert.Equal(t, "failed to unmarshal", errors.Unwrap(err).Error())
}

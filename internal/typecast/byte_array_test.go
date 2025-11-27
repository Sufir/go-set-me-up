package typecast

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestByteArrayOptionTypeCast_TruncationAndZeroPadding(t *testing.T) {
	optionType := ByteArrayOptionType{}

	t.Run("ShorterThanArrayLength", func(t *testing.T) {
		targetType := reflect.TypeOf([5]byte{})
		value, err := optionType.Cast("abc", targetType)
		require.NoError(t, err)
		obtained := value.Interface().([5]byte)
		assert.Equal(t, byte('a'), obtained[0])
		assert.Equal(t, byte('b'), obtained[1])
		assert.Equal(t, byte('c'), obtained[2])
		assert.Equal(t, byte(0), obtained[3])
		assert.Equal(t, byte(0), obtained[4])
	})

	t.Run("LongerThanArrayLength", func(t *testing.T) {
		targetType := reflect.TypeOf([3]byte{})
		value, err := optionType.Cast("abcdef", targetType)
		require.NoError(t, err)
		obtained := value.Interface().([3]byte)
		assert.Equal(t, byte('a'), obtained[0])
		assert.Equal(t, byte('b'), obtained[1])
		assert.Equal(t, byte('c'), obtained[2])
	})
}

package setup

import (
	"reflect"

	"github.com/Sufir/go-set-me-up/internal/typecast"
)

type (
	TypeCaster         = typecast.TypeCaster
	ErrUnsupportedType = typecast.ErrUnsupportedType
	ErrParseFailed     = typecast.ErrParseFailed
)

var ErrEmptyValue = typecast.ErrEmptyValue

type TypeCasterOption interface {
	Supports(targetType reflect.Type) bool
	Cast(value string, targetType reflect.Type) (reflect.Value, error)
}

func NewTypeCaster(optionTypes ...TypeCasterOption) typecast.TypeCaster {
	opts := make([]typecast.OptionType, 0, len(optionTypes))
	for _, o := range optionTypes {
		if o == nil {
			continue
		}
		opts = append(opts, optionAdapter{opt: o})
	}
	return typecast.NewCaster(opts...)
}

type optionAdapter struct {
	opt TypeCasterOption
}

func (a optionAdapter) Supports(targetType reflect.Type) bool {
	return a.opt.Supports(targetType)
}

func (a optionAdapter) Cast(value string, targetType reflect.Type) (reflect.Value, error) {
	return a.opt.Cast(value, targetType)
}

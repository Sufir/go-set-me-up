package typecast

import (
	"reflect"
	"sync"
)

type Caster struct {
	byType  map[reflect.Type]OptionType
	options []OptionType
	mu      sync.RWMutex
}

type OptionType interface {
	Supports(targetType reflect.Type) bool
	Cast(value string, targetType reflect.Type) (reflect.Value, error)
}

type TypeCaster interface {
	Cast(value string, targetType reflect.Type) (reflect.Value, error)
}

func NewCaster(types ...OptionType) TypeCaster {
	options := []OptionType{
		TextUnmarshalerOptionType{},
		ByteArrayOptionType{},
		StringOptionType{},
		ByteSliceOptionType{},
		BoolOptionType{},
		IntOptionType{},
		UintOptionType{},
		FloatOptionType{},
		ComplexOptionType{},
	}

	options = append(options, types...)

	return &Caster{
		byType:  make(map[reflect.Type]OptionType),
		options: options,
	}
}

func (c *Caster) Cast(value string, targetType reflect.Type) (reflect.Value, error) {
	if targetType.Kind() == reflect.Ptr {
		elementType := targetType.Elem()

		c.mu.RLock()
		cachedElementOptionType, found := c.byType[elementType]
		c.mu.RUnlock()
		if !found {
			for _, optionType := range c.options {
				if optionType.Supports(elementType) {
					cachedElementOptionType = optionType
					c.mu.Lock()
					c.byType[elementType] = cachedElementOptionType
					c.mu.Unlock()
					break
				}
			}
		}

		if cachedElementOptionType != nil {
			castedValue, castError := cachedElementOptionType.Cast(value, elementType)
			if castError != nil {
				return reflect.Value{}, castError
			}
			if !castedValue.Type().AssignableTo(elementType) && castedValue.Type().ConvertibleTo(elementType) {
				castedValue = castedValue.Convert(elementType)
			}
			pointerValue := reflect.New(elementType)
			pointerValue.Elem().Set(castedValue)
			return pointerValue, nil
		}
	}

	c.mu.RLock()
	cachedOptionType, found := c.byType[targetType]
	c.mu.RUnlock()
	if found {
		return cachedOptionType.Cast(value, targetType)
	}

	var selectedOptionType OptionType
	for _, optionType := range c.options {
		if optionType.Supports(targetType) {
			selectedOptionType = optionType
			break
		}
	}

	if selectedOptionType != nil {
		c.mu.Lock()
		c.byType[targetType] = selectedOptionType
		c.mu.Unlock()
		v, err := selectedOptionType.Cast(value, targetType)
		if err != nil {
			return reflect.Value{}, err
		}
		if !v.Type().AssignableTo(targetType) && v.Type().ConvertibleTo(targetType) {
			v = v.Convert(targetType)
		}
		return v, nil
	}

	return reflect.Value{}, ErrUnsupportedType{Type: targetType}
}

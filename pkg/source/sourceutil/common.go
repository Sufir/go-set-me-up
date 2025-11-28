package sourceutil

import (
	"reflect"

	"github.com/Sufir/go-set-me-up/internal/typecast"
	"github.com/Sufir/go-set-me-up/pkg"
)

func DefaultMode(mode pkg.LoadMode) pkg.LoadMode {
	if mode == 0 {
		return pkg.ModeOverride
	}

	return mode
}

func EnsureTargetStruct(configuration any) (reflect.Value, error) {
	value := reflect.ValueOf(configuration)
	if value.Kind() != reflect.Ptr || value.IsNil() {
		return reflect.Value{}, pkg.NewInvalidTargetError("target must be a non-nil pointer to struct")
	}

	elem := value.Elem()
	if elem.Kind() != reflect.Struct {
		return reflect.Value{}, pkg.NewInvalidTargetError("target must be pointer to struct")
	}

	return elem, nil
}

func ShouldAssign(fieldValue reflect.Value, present bool, mode pkg.LoadMode, defaultValue string) bool {
	if mode == pkg.ModeOverride {
		if present {
			return true
		}
		if defaultValue != "" && fieldValue.IsZero() {
			return true
		}
		return false
	}

	if mode == pkg.ModeFillMissing {
		if !fieldValue.IsZero() {
			return false
		}
		if present || defaultValue != "" {
			return true
		}
		return false
	}

	return false
}

func AssignFromString(caster typecast.TypeCaster, field reflect.Value, raw string) error {
	t := field.Type()
	if t.Kind() == reflect.Ptr {
		v, err := caster.Cast(raw, t.Elem())
		if err != nil {
			return err
		}

		if v.Type() == t {
			field.Set(v)
			return nil
		}

		if v.Kind() == reflect.Ptr && v.Type().Elem() == t.Elem() {
			field.Set(v)
			return nil
		}

		if v.Type() == t.Elem() {
			ptr := reflect.New(t.Elem())
			ptr.Elem().Set(v)
			field.Set(ptr)
			return nil
		}

		if v.Type().ConvertibleTo(t.Elem()) {
			ptr := reflect.New(t.Elem())
			ptr.Elem().Set(v.Convert(t.Elem()))
			field.Set(ptr)
			return nil
		}

		return typecast.ErrUnsupportedType{Type: t}
	}

	v, err := caster.Cast(raw, t)
	if err != nil {
		return err
	}

	if v.Type() == t {
		field.Set(v)
		return nil
	}

	if v.Kind() == reflect.Ptr && v.Type().Elem() == t {
		field.Set(v.Elem())
		return nil
	}

	return typecast.ErrUnsupportedType{Type: t}
}

func AssignFromAny(caster typecast.TypeCaster, field reflect.Value, raw any) error {
	t := field.Type()
	if raw == nil {
		if isNilAssignableKind(t.Kind()) {
			field.Set(reflect.Zero(t))
			return nil
		}
		return typecast.ErrUnsupportedType{Type: t}
	}

	rv := reflect.ValueOf(raw)
	if t.Kind() == reflect.Ptr {
		elem := t.Elem()
		if rv.Kind() == reflect.String {
			v, err := caster.Cast(rv.String(), elem)
			if err != nil {
				return err
			}
			if v.Type() == t {
				field.Set(v)
				return nil
			}
			if v.Kind() == reflect.Ptr && v.Type().Elem() == elem {
				field.Set(v)
				return nil
			}
			if v.Type() == elem {
				p := reflect.New(elem)
				p.Elem().Set(v)
				field.Set(p)
				return nil
			}
			if v.Type().ConvertibleTo(elem) {
				p := reflect.New(elem)
				p.Elem().Set(v.Convert(elem))
				field.Set(p)
				return nil
			}
			return typecast.ErrUnsupportedType{Type: t}
		}
		if rv.Type() == t {
			field.Set(rv)
			return nil
		}
		if rv.Kind() == reflect.Ptr && rv.Type().Elem() == elem {
			field.Set(rv)
			return nil
		}
		if rv.Type() == elem {
			p := reflect.New(elem)
			p.Elem().Set(rv)
			field.Set(p)
			return nil
		}
		if rv.Type().ConvertibleTo(elem) {
			p := reflect.New(elem)
			p.Elem().Set(rv.Convert(elem))
			field.Set(p)
			return nil
		}
		return typecast.ErrUnsupportedType{Type: t}
	}

	if rv.Kind() == reflect.String {
		v, err := caster.Cast(rv.String(), t)
		if err != nil {
			return err
		}
		if v.Type() == t {
			field.Set(v)
			return nil
		}
		if v.Kind() == reflect.Ptr && v.Type().Elem() == t {
			field.Set(v.Elem())
			return nil
		}
		if v.Type().ConvertibleTo(t) {
			field.Set(v.Convert(t))
			return nil
		}
		return typecast.ErrUnsupportedType{Type: t}
	}

	if rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			if isNilAssignableKind(t.Kind()) {
				field.Set(reflect.Zero(t))
				return nil
			}
			return typecast.ErrUnsupportedType{Type: t}
		}
		if rv.Type().Elem() == t {
			field.Set(rv.Elem())
			return nil
		}
		if rv.Elem().Type().ConvertibleTo(t) {
			field.Set(rv.Elem().Convert(t))
			return nil
		}
	}

	if rv.Type() == t {
		field.Set(rv)
		return nil
	}

	if rv.Type().ConvertibleTo(t) {
		field.Set(rv.Convert(t))
		return nil
	}

	return typecast.ErrUnsupportedType{Type: t}
}

func isNilAssignableKind(k reflect.Kind) bool {
	switch k {
	case reflect.Ptr, reflect.Slice, reflect.Map, reflect.Func, reflect.Interface, reflect.Chan:
		return true
	default:
		return false
	}
}

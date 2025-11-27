package dict

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/Sufir/go-set-me-up/internal/typecast"
	"github.com/Sufir/go-set-me-up/pkg"
)

type Source struct {
	dict   map[string]any
	caster typecast.TypeCaster
}

func NewSource(dict map[string]any) *Source {
	if dict == nil {
		dict = map[string]any{}
	}
	return &Source{dict: dict, caster: typecast.NewCaster()}
}

func (source *Source) Load(cfg any, mode pkg.LoadMode) error {
	if mode == 0 {
		mode = pkg.ModeOverride
	}
	v := reflect.ValueOf(cfg)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		return errors.New("target must be a non-nil pointer to struct")
	}
	e := v.Elem()
	if e.Kind() != reflect.Struct {
		return errors.New("target must be pointer to struct")
	}
	var errs []error
	source.loadStruct(e, source.dict, mode, &errs, "")
	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

func (source Source) loadStruct(structValue reflect.Value, dict map[string]any, mode pkg.LoadMode, errs *[]error, prefix string) {
	t := structValue.Type()
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if f.PkgPath != "" {
			continue
		}
		fv := structValue.Field(i)
		raw, ok := source.lookupValue(dict, f.Name)
		if !ok {
			continue
		}
		if m, isMap := asMapStringAny(raw); isMap {
			if fv.Kind() == reflect.Struct {
				source.loadStruct(fv, m, mode, errs, makePath(prefix, f.Name))
				continue
			}
			if fv.Kind() == reflect.Ptr && fv.Type().Elem().Kind() == reflect.Struct {
				if fv.IsNil() {
					fv.Set(reflect.New(fv.Type().Elem()))
				}
				source.loadStruct(fv.Elem(), m, mode, errs, makePath(prefix, f.Name))
				continue
			}
			continue
		}
		if !source.shouldSetField(fv, mode) {
			continue
		}
		if err := source.setFieldValue(fv, raw); err != nil {
			*errs = append(*errs, fmt.Errorf("field %s (type %v): %w", makePath(prefix, f.Name), fv.Type(), err))
		}
	}
}

func (source Source) lookupValue(dict map[string]any, fieldName string) (any, bool) {
	if v, ok := dict[fieldName]; ok {
		return v, true
	}
	upper := convertToUpperSnake(fieldName)
	lower := strings.ToLower(upper)
	if v, ok := dict[lower]; ok {
		return v, true
	}
	if v, ok := dict[upper]; ok {
		return v, true
	}
	return nil, false
}

func asMapStringAny(v any) (map[string]any, bool) {
	if v == nil {
		return nil, false
	}
	m, ok := v.(map[string]any)
	return m, ok
}

func (source Source) shouldSetField(fieldValue reflect.Value, mode pkg.LoadMode) bool {
	if mode == pkg.ModeOverride {
		return true
	}
	if mode == pkg.ModeFillMissing {
		return fieldValue.IsZero()
	}
	return false
}

func (source Source) setFieldValue(field reflect.Value, raw any) error {
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
			v, err := source.caster.Cast(rv.String(), elem)
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
		v, err := source.caster.Cast(rv.String(), t)
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

func makePath(prefix, name string) string {
	if prefix == "" {
		return name
	}
	return prefix + "." + name
}

func convertToUpperSnake(name string) string {
	var b strings.Builder
	b.Grow(len(name))
	lastUnderscore := false
	wroteAny := false
	prevLowerOrDigit := false
	prevUpper := false
	runes := []rune(name)
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		if r == '-' || r == ' ' || r == '_' {
			if !lastUnderscore && wroteAny {
				b.WriteByte('_')
				lastUnderscore = true
			}
			prevLowerOrDigit = false
			prevUpper = false
			continue
		}
		isUpper := r >= 'A' && r <= 'Z'
		isLower := r >= 'a' && r <= 'z'
		isDigit := r >= '0' && r <= '9'
		if isUpper {
			nextLower := false
			if i+1 < len(runes) {
				rr := runes[i+1]
				nextLower = rr >= 'a' && rr <= 'z'
			}
			if (prevLowerOrDigit || (prevUpper && nextLower)) && !lastUnderscore && wroteAny {
				b.WriteByte('_')
			}
			b.WriteRune(r)
			lastUnderscore = false
			wroteAny = true
			prevLowerOrDigit = false
			prevUpper = true
			continue
		}
		if isLower {
			b.WriteRune(r - ('a' - 'A'))
			lastUnderscore = false
			wroteAny = true
			prevLowerOrDigit = true
			prevUpper = false
			continue
		}
		if isDigit {
			b.WriteRune(r)
			lastUnderscore = false
			wroteAny = true
			prevLowerOrDigit = true
			prevUpper = false
			continue
		}
		if !lastUnderscore && wroteAny {
			b.WriteByte('_')
			lastUnderscore = true
		}
		prevLowerOrDigit = false
		prevUpper = false
	}
	s := b.String()
	if len(s) > 0 && s[len(s)-1] == '_' {
		s = s[:len(s)-1]
	}
	return s
}

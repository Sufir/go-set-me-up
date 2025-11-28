package dict

import (
	"errors"
	"reflect"
	"strings"

	"github.com/Sufir/go-set-me-up/internal/typecast"
	"github.com/Sufir/go-set-me-up/pkg"
	"github.com/Sufir/go-set-me-up/pkg/source/sourceutil"
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

func (source Source) Load(cfg any, mode pkg.LoadMode) error {
	mode = sourceutil.DefaultMode(mode)
	e, err := sourceutil.EnsureTargetStruct(cfg)
	if err != nil {
		return err
	}
	var errs []error
	source.loadStruct(e, source.dict, mode, &errs, "")
	if len(errs) > 0 {
		return pkg.NewAggregatedLoadFailedError(errors.Join(errs...))
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
		if err := sourceutil.AssignFromAny(source.caster, fv, raw); err != nil {
			*errs = append(*errs, pkg.NewDictFieldFailedError(makePath(prefix, f.Name), err))
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

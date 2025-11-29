package dict

import (
	"errors"
	"reflect"
	"strings"

	"github.com/Sufir/go-set-me-up/pkg"
	"github.com/Sufir/go-set-me-up/pkg/source/sourceutil"
)

type Source struct {
	dict   map[string]any
	caster pkg.TypeCaster
	mode   pkg.LoadMode
}

func NewSource(dict map[string]any, mode pkg.LoadMode) *Source {
	if dict == nil {
		dict = map[string]any{}
	}
	return &Source{dict: dict, caster: pkg.NewTypeCaster(), mode: sourceutil.DefaultMode(mode)}
}

func NewSourceWithCaster(dict map[string]any, mode pkg.LoadMode, caster pkg.TypeCaster) *Source {
	if dict == nil {
		dict = map[string]any{}
	}
	if caster == nil {
		caster = pkg.NewTypeCaster()
	}
	return &Source{dict: dict, caster: caster, mode: sourceutil.DefaultMode(mode)}
}

func (source Source) Load(cfg any) error {
	e, err := sourceutil.EnsureTargetStruct(cfg)
	if err != nil {
		return err
	}
	var errs []error
	source.loadStruct(e, source.dict, source.mode, &errs, "")
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
				source.loadStruct(fv, m, mode, errs, sourceutil.MakePath(prefix, f.Name))
				continue
			}
			if fv.Kind() == reflect.Ptr && fv.Type().Elem().Kind() == reflect.Struct {
				if fv.IsNil() {
					fv.Set(reflect.New(fv.Type().Elem()))
				}
				source.loadStruct(fv.Elem(), m, mode, errs, sourceutil.MakePath(prefix, f.Name))
				continue
			}
			continue
		}
		if !sourceutil.ShouldAssign(fv, true, mode, "") {
			continue
		}
		if err := sourceutil.AssignFromAny(source.caster, fv, raw); err != nil {
			*errs = append(*errs, pkg.NewDictFieldFailedError(sourceutil.MakePath(prefix, f.Name), err))
		}
	}
}

func (source Source) lookupValue(dict map[string]any, fieldName string) (any, bool) {
	if v, ok := dict[fieldName]; ok {
		return v, true
	}
	upper := sourceutil.ConvertToUpperSnake(fieldName)
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

// removed local shouldSetField; using sourceutil.ShouldAssign for consistent mode semantics

// makePath proxy removed; use sourceutil.MakePath directly where needed

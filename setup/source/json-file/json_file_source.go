package jsonfile

import (
	"encoding/json"
	"os"
	"reflect"
	"strings"

	"github.com/Sufir/go-set-me-up/setup"
	"github.com/Sufir/go-set-me-up/setup/source/sourceutil"
)

type Source struct {
	path string
	mode setup.LoadMode
}

func NewSource(path string, mode setup.LoadMode) *Source {
	return &Source{path: path, mode: sourceutil.DefaultMode(mode)}
}

func (source Source) Load(cfg any) error {
	elem, err := sourceutil.EnsureTargetStruct(cfg)
	if err != nil {
		return err
	}

	data, readErr := os.ReadFile(source.path)
	if readErr != nil {
		return setup.NewAggregatedLoadFailedError(readErr)
	}

	var root map[string]json.RawMessage
	if unmarshalErr := json.Unmarshal(data, &root); unmarshalErr != nil {
		return setup.NewAggregatedLoadFailedError(unmarshalErr)
	}

	holder := reflect.New(elem.Type())
	if err := json.Unmarshal(data, holder.Interface()); err != nil {
		return setup.NewAggregatedLoadFailedError(err)
	}
	shadow := holder.Elem()
	source.copyStructValues(elem, shadow, root, source.mode, nil, "")
	return nil
}

func (source Source) copyStructValues(dest reflect.Value, shadow reflect.Value, raw map[string]json.RawMessage, mode setup.LoadMode, _ *[]error, prefix string) {
	structType := dest.Type()
	for i := 0; i < structType.NumField(); i++ {
		fieldInfo := structType.Field(i)
		if fieldInfo.PkgPath != "" {
			continue
		}
		jsonTag := fieldInfo.Tag.Get("json")
		name := parseJSONTagName(jsonTag)
		if name == "" {
			continue
		}
		present := raw != nil
		var childRaw map[string]json.RawMessage
		if present {
			v, ok := raw[name]
			present = ok
			if ok {
				var nested map[string]json.RawMessage
				if json.Unmarshal(v, &nested) == nil {
					childRaw = nested
				}
			}
		}
		destField := dest.Field(i)
		shadowField := shadow.Field(i)
		t := fieldInfo.Type
		if t.Kind() == reflect.Struct {
			source.copyStructValues(destField, shadowField, childRaw, mode, nil, sourceutil.MakePath(prefix, fieldInfo.Name))
			continue
		}
		if t.Kind() == reflect.Ptr && t.Elem().Kind() == reflect.Struct {
			if destField.IsNil() {
				destField.Set(reflect.New(t.Elem()))
			}
			var shadowStruct reflect.Value
			if shadowField.IsNil() {
				shadowStruct = reflect.New(t.Elem()).Elem()
			} else {
				shadowStruct = shadowField.Elem()
			}
			source.copyStructValues(destField.Elem(), shadowStruct, childRaw, mode, nil, sourceutil.MakePath(prefix, fieldInfo.Name))
			continue
		}
		if !sourceutil.ShouldAssign(destField, present, mode, "") {
			continue
		}
		if destField.Type() == shadowField.Type() {
			destField.Set(shadowField)
			continue
		}
		if destField.Type().Kind() == reflect.Ptr && shadowField.Type() == destField.Type().Elem() {
			p := reflect.New(destField.Type().Elem())
			p.Elem().Set(shadowField)
			destField.Set(p)
			continue
		}
		if shadowField.Type().Kind() == reflect.Ptr && shadowField.Type().Elem() == destField.Type() {
			if shadowField.IsNil() {
				destField.Set(reflect.Zero(destField.Type()))
			} else {
				destField.Set(shadowField.Elem())
			}
			continue
		}
	}
}

func parseJSONTagName(tag string) string {
	if tag == "" || tag == "-" {
		return ""
	}
	if idx := strings.IndexByte(tag, ','); idx >= 0 {
		return tag[:idx]
	}
	return tag
}

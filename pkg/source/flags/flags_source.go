package flags

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/Sufir/go-set-me-up/pkg"
	"github.com/Sufir/go-set-me-up/pkg/source/sourceutil"
)

type Source struct {
	caster    pkg.TypeCaster
	delimiter string
	mode      pkg.LoadMode
}

func NewSource(mode pkg.LoadMode) *Source {
	return &Source{caster: pkg.NewTypeCaster(), mode: sourceutil.DefaultMode(mode), delimiter: ","}
}

func NewSourceWithCaster(mode pkg.LoadMode, caster pkg.TypeCaster) *Source {
	if caster == nil {
		caster = pkg.NewTypeCaster()
	}
	return &Source{caster: caster, mode: sourceutil.DefaultMode(mode), delimiter: ","}
}

func NewSourceWithDelimiter(mode pkg.LoadMode, delimiter string) *Source {
	if delimiter == "" {
		delimiter = ","
	}
	return &Source{caster: pkg.NewTypeCaster(), mode: sourceutil.DefaultMode(mode), delimiter: delimiter}
}

func NewSourceWithCasterAndDelimiter(mode pkg.LoadMode, delimiter string, caster pkg.TypeCaster) *Source {
	if delimiter == "" {
		delimiter = ","
	}
	if caster == nil {
		caster = pkg.NewTypeCaster()
	}
	return &Source{caster: caster, mode: sourceutil.DefaultMode(mode), delimiter: delimiter}
}

func (source Source) Load(cfg any) error {
	elem, err := sourceutil.EnsureTargetStruct(cfg)
	if err != nil {
		return err
	}

	argsMap := parseArguments(os.Args[1:])
	var collected []error
	source.loadStruct(elem, argsMap, source.mode, &collected)
	if len(collected) > 0 {
		return pkg.NewAggregatedLoadFailedError(errors.Join(collected...))
	}
	return nil
}

func parseArguments(args []string) map[string]string {
	result := make(map[string]string)
	i := 0
	for i < len(args) {
		token := args[i]
		if strings.HasPrefix(token, "--") {
			name := token[2:]
			if name == "" {
				i++
				continue
			}
			if eq := strings.IndexByte(name, '='); eq >= 0 {
				key := name[:eq]
				value := name[eq+1:]
				if strings.HasPrefix(key, "no-") {
					k := key[3:]
					result[k] = "false"
				} else {
					result[key] = value
				}
				i++
				continue
			}
			if strings.HasPrefix(name, "no-") {
				k := name[3:]
				result[k] = "false"
				i++
				continue
			}
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				result[name] = args[i+1]
				i += 2
				continue
			}
			result[name] = ""
			i++
			continue
		}
		if strings.HasPrefix(token, "-") && len(token) >= 2 {
			name := token[1:]
			if name == "" {
				i++
				continue
			}
			if eq := strings.IndexByte(name, '='); eq >= 0 {
				key := name[:eq]
				value := name[eq+1:]
				result[key] = value
				i++
				continue
			}
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				result[name] = args[i+1]
				i += 2
				continue
			}
			result[name] = ""
			i++
			continue
		}
		i++
	}
	return result
}

func (source Source) loadStruct(structValue reflect.Value, args map[string]string, mode pkg.LoadMode, errs *[]error) {
	structType := structValue.Type()
	for i := 0; i < structType.NumField(); i++ {
		fieldInfo := structType.Field(i)
		if fieldInfo.PkgPath != "" {
			continue
		}
		fieldValue := structValue.Field(i)
		if source.processLeafField(fieldValue, fieldInfo, args, mode, errs) {
			continue
		}
		t := fieldInfo.Type
		if t.Kind() == reflect.Struct {
			source.loadStruct(fieldValue, args, mode, errs)
			continue
		}
		if t.Kind() == reflect.Ptr && t.Elem().Kind() == reflect.Struct {
			if fieldValue.IsNil() {
				fieldValue.Set(reflect.New(t.Elem()))
			}
			source.loadStruct(fieldValue.Elem(), args, mode, errs)
			continue
		}
	}
}

func (source Source) processLeafField(fieldValue reflect.Value, fieldInfo reflect.StructField, args map[string]string, mode pkg.LoadMode, errs *[]error) bool {
	tagFlag := fieldInfo.Tag.Get("flag")
	if tagFlag == "" || tagFlag == "-" {
		return false
	}
	tagShort := fieldInfo.Tag.Get("flagShort")
	tagDefault := fieldInfo.Tag.Get("flagDefault")
	v, ok := args[tagFlag]
	usedShort := false
	if !ok && tagShort != "" {
		v, ok = args[tagShort]
		usedShort = ok
	}
	if !sourceutil.ShouldAssign(fieldValue, ok, mode, tagDefault) {
		return true
	}
	raw := ""
	if ok {
		raw = v
		if raw == "" {
			t := fieldValue.Type()
			if t.Kind() == reflect.Ptr {
				t = t.Elem()
			}
			if t.Kind() == reflect.Bool {
				raw = "true"
			} else {
				name := tagFlag
				if usedShort {
					name = tagShort
				}
				parseErr := pkg.ErrParseFailed{Type: t, Value: raw, Cause: pkg.ErrEmptyValue}
				*errs = append(*errs, fmt.Errorf("%s=%s: %w", name, raw, parseErr))
				return true
			}
		}
	} else {
		if tagDefault == "" {
			return true
		}
		raw = tagDefault
	}
	t := fieldInfo.Type
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() == reflect.Slice || (t.Kind() == reflect.Array && t.Elem().Kind() == reflect.Int) {
		elemKind := t.Elem().Kind()
		if elemKind == reflect.String || elemKind == reflect.Int {
			delim := sourceutil.ResolveDelimiter(fieldInfo.Tag.Get("flagDelim"), source.delimiter)
			raw = sourceutil.NormalizeDelimited(raw, delim)
		}
	}
	if err := sourceutil.AssignFromString(source.caster, fieldValue, raw); err != nil {
		name := tagFlag
		if usedShort {
			name = tagShort
		}
		path := fieldInfo.Name
		*errs = append(*errs, pkg.NewFlagsFieldFailedError(name, raw, path, err))
	}
	return true
}

// Removed local shouldSetField and setFieldValue in favor of common utilities.

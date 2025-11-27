package env

import (
	"errors"
	"os"
	"reflect"
	"strings"

	"github.com/Sufir/go-set-me-up/internal/typecast"
	"github.com/Sufir/go-set-me-up/pkg"
)

type Source struct {
	caster    typecast.TypeCaster
	prefix    string
	delimiter string
}

func NewSource(prefix string, delimiter string) *Source {
	normalized := ""
	if prefix != "" {
		normalized = convertToEnvVar(prefix)
	}

	if delimiter == "" {
		delimiter = ","
	}

	return &Source{
		prefix:    normalized,
		delimiter: delimiter,
		caster:    typecast.NewCaster(),
	}
}

func (source Source) Load(cfg any, mode pkg.LoadMode) error {
	if mode == 0 {
		mode = pkg.ModeOverride
	}

	value := reflect.ValueOf(cfg)
	if value.Kind() != reflect.Ptr || value.IsNil() {
		return errors.New("target must be a non-nil pointer to struct")
	}

	elem := value.Elem()
	if elem.Kind() != reflect.Struct {
		return errors.New("target must be pointer to struct")
	}

	environment := getEnv()
	var collected []error
	segments := []string{}
	if source.prefix != "" {
		segments = append(segments, source.prefix)
	}

	source.loadStruct(elem, segments, environment, mode, &collected)

	if len(collected) > 0 {
		return errors.Join(collected...)
	}

	return nil
}

func getEnv() map[string]string {
	environment := os.Environ()
	result := make(map[string]string, len(environment))

	for _, pair := range environment {
		index := strings.IndexByte(pair, '=')
		if index <= 0 {
			continue
		}
		key := pair[:index]
		value := pair[index+1:]
		result[key] = value
	}

	return result
}

func (source Source) loadStruct(structValue reflect.Value, segments []string, env map[string]string, mode pkg.LoadMode, errs *[]error) {
	structType := structValue.Type()
	for i := 0; i < structType.NumField(); i++ {
		fieldInfo := structType.Field(i)
		if fieldInfo.PkgPath != "" {
			continue
		}
		if fieldInfo.Tag.Get("env") == "-" {
			continue
		}
		fieldValue := structValue.Field(i)
		if source.processLeafField(fieldValue, fieldInfo, segments, env, mode, errs) {
			continue
		}
		nestedValue, nextSegments, ok := source.resolveNestedStruct(fieldValue, fieldInfo, segments)
		if !ok {
			continue
		}
		source.loadStruct(nestedValue, nextSegments, env, mode, errs)
	}
}

func appendIfNotEmpty(segments []string, name string) []string {
	if name == "" {
		return segments
	}
	return append(segments, name)
}

func (source Source) segmentForField(fieldInfo reflect.StructField) string {
	segmentName := fieldInfo.Tag.Get("envSegment")
	if segmentName == "" {
		segmentName = fieldInfo.Name
	}
	return convertToEnvVar(segmentName)
}

func (source Source) resolveNestedStruct(fieldValue reflect.Value, fieldInfo reflect.StructField, segments []string) (reflect.Value, []string, bool) {
	t := fieldInfo.Type
	switch t.Kind() {
	case reflect.Struct:
		segmentName := source.segmentForField(fieldInfo)
		nextSegments := appendIfNotEmpty(segments, segmentName)
		return fieldValue, nextSegments, true
	case reflect.Ptr:
		if t.Elem().Kind() != reflect.Struct {
			return reflect.Value{}, nil, false
		}
		segmentName := source.segmentForField(fieldInfo)
		nextSegments := appendIfNotEmpty(segments, segmentName)
		if fieldValue.IsNil() {
			fieldValue.Set(reflect.New(t.Elem()))
		}
		return fieldValue.Elem(), nextSegments, true
	default:
		return reflect.Value{}, nil, false
	}
}

func (source Source) processLeafField(fieldValue reflect.Value, fieldInfo reflect.StructField, segments []string, env map[string]string, mode pkg.LoadMode, errs *[]error) bool {
	tagEnv := fieldInfo.Tag.Get("env")
	if tagEnv == "" {
		return false
	}
	leaf := convertToEnvVar(tagEnv)
	key := buildKey(segments, leaf)
	val, ok := env[key]
	defaultValue := fieldInfo.Tag.Get("envDefault")
	if !source.shouldSetField(fieldValue, ok, mode, defaultValue) {
		return true
	}
	setValue := ""
	if ok {
		setValue = val
	} else {
		if defaultValue == "" {
			return true
		}
		setValue = defaultValue
	}
	if err := source.setFieldValue(fieldValue, setValue); err != nil {
		*errs = append(*errs, err)
	}
	return true
}

func buildKey(segments []string, leaf string) string {
	if len(segments) == 0 {
		return leaf
	}

	if leaf == "" {
		return strings.Join(segments, "_")
	}

	return strings.Join(append(segments, leaf), "_")
}

func (source Source) shouldSetField(fieldValue reflect.Value, envPresent bool, mode pkg.LoadMode, defaultValue string) bool {
	if mode == pkg.ModeOverride {
		if envPresent {
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

		if envPresent || defaultValue != "" {
			return true
		}

		return false
	}

	return false
}

func splitIntoTokens(value string, delimiter string) []string {
	if delimiter == "" {
		return []string{value}
	}

	return strings.Split(value, delimiter)
}

func (source Source) setFieldValue(field reflect.Value, raw string) error {
	t := field.Type()

	if t.Kind() == reflect.Ptr {
		v, err := source.caster.Cast(raw, t.Elem())
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

	v, err := source.caster.Cast(raw, t)
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

func convertToEnvVar(name string) string {
	var builder strings.Builder
	builder.Grow(len(name))

	lastUnderscore := false
	wroteAny := false
	prevLowerOrDigit := false
	prevUpper := false

	runes := []rune(name)
	for i := 0; i < len(runes); i++ {
		r := runes[i]

		if r == '-' || r == ' ' {
			if !lastUnderscore && wroteAny {
				builder.WriteByte('_')
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
				builder.WriteByte('_')
			}
			builder.WriteRune(r)
			lastUnderscore = false
			wroteAny = true
			prevLowerOrDigit = false
			prevUpper = true
			continue
		}

		if isLower {
			builder.WriteRune(r - ('a' - 'A'))
			lastUnderscore = false
			wroteAny = true
			prevLowerOrDigit = true
			prevUpper = false
			continue
		}

		if isDigit {
			builder.WriteRune(r)
			lastUnderscore = false
			wroteAny = true
			prevLowerOrDigit = true
			prevUpper = false
			continue
		}

		if !lastUnderscore && wroteAny {
			builder.WriteByte('_')
			lastUnderscore = true
		}
		prevLowerOrDigit = false
		prevUpper = false
	}

	s := builder.String()
	if len(s) > 0 && s[len(s)-1] == '_' {
		s = s[:len(s)-1]
	}

	return s
}

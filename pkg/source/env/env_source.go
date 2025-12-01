package env

import (
	"errors"
	"os"
	"reflect"
	"strings"

	"github.com/Sufir/go-set-me-up/pkg"
	"github.com/Sufir/go-set-me-up/pkg/source/sourceutil"
)

type Source struct {
	caster    pkg.TypeCaster
	prefix    string
	delimiter string
	mode      pkg.LoadMode
}

func NewSource(prefix string, delimiter string, mode pkg.LoadMode) *Source {
	normalized := ""
	if prefix != "" {
		normalized = sourceutil.ConvertToEnvVar(prefix)
	}

	if delimiter == "" {
		delimiter = ","
	}

	return &Source{
		prefix:    normalized,
		delimiter: delimiter,
		caster:    pkg.NewTypeCaster(),
		mode:      sourceutil.DefaultMode(mode),
	}
}

func NewSourceWithCaster(prefix string, delimiter string, mode pkg.LoadMode, caster pkg.TypeCaster) *Source {
	normalized := ""
	if prefix != "" {
		normalized = sourceutil.ConvertToEnvVar(prefix)
	}
	if delimiter == "" {
		delimiter = ","
	}
	if caster == nil {
		caster = pkg.NewTypeCaster()
	}
	return &Source{
		prefix:    normalized,
		delimiter: delimiter,
		caster:    caster,
		mode:      sourceutil.DefaultMode(mode),
	}
}

func (source Source) Load(cfg any) error {
	elem, err := sourceutil.EnsureTargetStruct(cfg)
	if err != nil {
		return err
	}

	environment := getEnv()
	var collected []error
	segments := []string{}
	if source.prefix != "" {
		segments = append(segments, source.prefix)
	}

	source.loadStruct(elem, segments, environment, source.mode, &collected, "")

	if len(collected) > 0 {
		return pkg.NewAggregatedLoadFailedError(errors.Join(collected...))
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

func (source Source) loadStruct(structValue reflect.Value, segments []string, env map[string]string, mode pkg.LoadMode, errs *[]error, prefix string) {
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
		if source.processLeafField(fieldValue, fieldInfo, segments, env, mode, errs, prefix) {
			continue
		}
		nestedValue, nextSegments, ok := source.resolveNestedStruct(fieldValue, fieldInfo, segments)
		if !ok {
			continue
		}
		source.loadStruct(nestedValue, nextSegments, env, mode, errs, sourceutil.MakePath(prefix, fieldInfo.Name))
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
	return sourceutil.ConvertToEnvVar(segmentName)
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

func (source Source) processLeafField(fieldValue reflect.Value, fieldInfo reflect.StructField, segments []string, env map[string]string, mode pkg.LoadMode, errs *[]error, prefix string) bool {
	tagEnv := fieldInfo.Tag.Get("env")
	if tagEnv == "" {
		return false
	}
	leaf := sourceutil.ConvertToEnvVar(tagEnv)
	key := buildKey(segments, leaf)
	val, ok := env[key]
	defaultValue := fieldInfo.Tag.Get("envDefault")
	if !sourceutil.ShouldAssign(fieldValue, ok, mode, defaultValue) {
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
	t := fieldInfo.Type
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() == reflect.Slice || (t.Kind() == reflect.Array && t.Elem().Kind() == reflect.Int) {
		elemKind := t.Elem().Kind()
		if elemKind == reflect.String || elemKind == reflect.Int {
			delim := sourceutil.ResolveDelimiter(fieldInfo.Tag.Get("envDelim"), source.delimiter)
			setValue = sourceutil.NormalizeDelimited(setValue, delim)
		}
	}
	if err := sourceutil.AssignFromString(source.caster, fieldValue, setValue); err != nil {
		path := sourceutil.MakePath(prefix, fieldInfo.Name)
		*errs = append(*errs, pkg.NewEnvFieldFailedError(key, setValue, path, err))
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

// convertToEnvVar proxy removed; use sourceutil.ConvertToEnvVar directly

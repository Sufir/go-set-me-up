package sourceutil

import (
	"reflect"
	"strings"
	"unicode"

	"github.com/Sufir/go-set-me-up/internal/typecast"
	"github.com/Sufir/go-set-me-up/pkg"
)

// DefaultMode returns the effective load mode.
// mode: requested mode; when zero, ModeOverride is used as the default.
func DefaultMode(mode pkg.LoadMode) pkg.LoadMode {
	if mode == 0 {
		return pkg.ModeOverride
	}

	return mode
}

// EnsureTargetStruct verifies that configuration is a non-nil pointer to a struct
// and returns the underlying struct value.
// configuration: configuration object that must be a pointer to a struct.
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

// ShouldAssign decides whether a destination field should be assigned based on
// the presence of a value, the load mode, and a non-empty default.
// fieldValue: destination field
// present: whether the source contains a value for the field
// mode: load mode controlling override/fill semantics
// defaultValue: non-empty default value used when source does not contain a value
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

// AssignFromString converts a raw string to the field's type using the provided TypeCaster
// and assigns the result into the destination field.
// caster: TypeCaster used for string-to-type conversion
// field: destination field to set
// raw: input string value
func AssignFromString(caster pkg.TypeCaster, field reflect.Value, raw string) error {
	t := field.Type()
	if t.Kind() == reflect.Ptr {
		v, err := caster.Cast(raw, t.Elem())
		if err != nil {
			return err
		}
		if assignExactType(field, v) {
			return nil
		}
		if v.Kind() == reflect.Ptr && v.Type().Elem() == t.Elem() {
			field.Set(v)
			return nil
		}
		if wrapPointer(field, v) {
			return nil
		}
		return pkg.ErrUnsupportedType{Type: t}
	}

	v, err := caster.Cast(raw, t)
	if err != nil {
		return err
	}
	if assignExactType(field, v) {
		return nil
	}
	if unwrapPointerExact(field, v) {
		return nil
	}
	if assignConvertible(field, v) {
		return nil
	}
	return pkg.ErrUnsupportedType{Type: t}
}

// AssignFromAny assigns an arbitrary typed value into the destination field, performing
// pointer wrapping/unwrapping and type conversions where supported. String inputs use
// the provided TypeCaster for conversion.
// caster: TypeCaster used when converting from string values
// field: destination field to set
// raw: input value (may be string, pointer, nil, or concrete type)
func AssignFromAny(caster pkg.TypeCaster, field reflect.Value, raw any) error {
	t := field.Type()
	if raw == nil {
		if isNilAssignableKind(t.Kind()) {
			field.Set(reflect.Zero(t))
			return nil
		}
		return pkg.ErrUnsupportedType{Type: t}
	}

	rv := reflect.ValueOf(raw)
	if t.Kind() == reflect.Ptr {
		elem := t.Elem()
		if rv.Kind() == reflect.String {
			v, err := caster.Cast(rv.String(), elem)
			if err != nil {
				return err
			}
			if assignExactType(field, v) {
				return nil
			}
			if v.Kind() == reflect.Ptr && v.Type().Elem() == elem {
				field.Set(v)
				return nil
			}
			if wrapPointer(field, v) {
				return nil
			}
			return pkg.ErrUnsupportedType{Type: t}
		}
		if assignExactType(field, rv) {
			return nil
		}
		if rv.Kind() == reflect.Ptr && rv.Type().Elem() == elem {
			field.Set(rv)
			return nil
		}
		if wrapPointer(field, rv) {
			return nil
		}
		return typecast.ErrUnsupportedType{Type: t}
	}

	if rv.Kind() == reflect.String {
		v, err := caster.Cast(rv.String(), t)
		if err != nil {
			return err
		}
		if assignExactType(field, v) {
			return nil
		}
		if unwrapPointer(field, v) {
			return nil
		}
		if assignConvertible(field, v) {
			return nil
		}
		return pkg.ErrUnsupportedType{Type: t}
	}

	if rv.Type() == t {
		field.Set(rv)
		return nil
	}
	if rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			if isNilAssignableKind(t.Kind()) {
				field.Set(reflect.Zero(t))
				return nil
			}
			return pkg.ErrUnsupportedType{Type: t}
		}
		if unwrapPointer(field, rv) {
			return nil
		}
	}

	if assignExactType(field, rv) {
		return nil
	}
	if assignConvertible(field, rv) {
		return nil
	}
	return typecast.ErrUnsupportedType{Type: t}
}

func assignExactType(field reflect.Value, v reflect.Value) bool {
	if v.Type() == field.Type() {
		field.Set(v)
		return true
	}
	return false
}

func assignConvertible(field reflect.Value, v reflect.Value) bool {
	t := field.Type()
	if v.Type().ConvertibleTo(t) {
		field.Set(v.Convert(t))
		return true
	}
	return false
}

func wrapPointer(field reflect.Value, v reflect.Value) bool {
	t := field.Type()
	if t.Kind() != reflect.Ptr {
		return false
	}
	elem := t.Elem()
	if v.Type() == elem {
		p := reflect.New(elem)
		p.Elem().Set(v)
		field.Set(p)
		return true
	}
	if v.Type().ConvertibleTo(elem) {
		p := reflect.New(elem)
		p.Elem().Set(v.Convert(elem))
		field.Set(p)
		return true
	}
	return false
}

func unwrapPointer(field reflect.Value, v reflect.Value) bool {
	if v.Kind() == reflect.Ptr {
		if v.Type().Elem() == field.Type() {
			field.Set(v.Elem())
			return true
		}
		if v.Elem().Type().ConvertibleTo(field.Type()) {
			field.Set(v.Elem().Convert(field.Type()))
			return true
		}
	}
	return false
}

func unwrapPointerExact(field reflect.Value, v reflect.Value) bool {
	if v.Kind() == reflect.Ptr && v.Type().Elem() == field.Type() {
		field.Set(v.Elem())
		return true
	}
	return false
}

func isNilAssignableKind(k reflect.Kind) bool {
	switch k {
	case reflect.Ptr, reflect.Slice, reflect.Map, reflect.Func, reflect.Interface, reflect.Chan:
		return true
	default:
		return false
	}
}

func MakePath(prefix, name string) string {
	if prefix == "" {
		return name
	}
	return prefix + "." + name
}

func ConvertToUpperSnake(name string) string {
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

func NormalizeDelimited(input string, delim string) string {
	if delim == "" {
		return input
	}
	if !strings.Contains(input, delim) {
		return input
	}
	tokens := strings.Split(input, delim)
	for i := range tokens {
		tokens[i] = strings.TrimSpace(tokens[i])
	}
	return strings.Join(tokens, ",")
}

func ResolveDelimiter(tagDelimiter string, defaultDelimiter string) string {
	if tagDelimiter != "" {
		return tagDelimiter
	}
	return defaultDelimiter
}

func ConvertToEnvVar(name string) string {
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

		isUpper := unicode.IsUpper(r)
		isLower := unicode.IsLower(r)
		isDigit := unicode.IsDigit(r)

		if isUpper {
			nextLower := false
			if i+1 < len(runes) {
				rr := runes[i+1]
				nextLower = unicode.IsLower(rr)
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
			builder.WriteRune(unicode.ToUpper(r))
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

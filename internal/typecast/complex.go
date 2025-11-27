package typecast

import (
	"reflect"
	"strconv"
	"strings"
)

type ComplexOptionType struct{}

func (ComplexOptionType) Supports(targetType reflect.Type) bool {
	return targetType.Kind() == reflect.Complex64 || targetType.Kind() == reflect.Complex128
}

func (ComplexOptionType) Cast(value string, targetType reflect.Type) (reflect.Value, error) {
	c, err := parseComplex(strings.TrimSpace(value), targetType.Bits())
	if err != nil {
		return reflect.Value{}, ErrParseFailed{Type: targetType, Value: value, Cause: err}
	}
	return reflect.ValueOf(c).Convert(targetType), nil
}

func parseComplex(s string, bits int) (complex128, error) {
	if s == "" {
		return 0, ErrEmptyValue
	}
	if s[0] == '(' && s[len(s)-1] == ')' {
		s = s[1 : len(s)-1]
	}

	if !strings.ContainsAny(s, "iI") {
		r, err := strconv.ParseFloat(s, bits)
		if err != nil {
			return 0, err
		}
		return complex(r, 0), nil
	}

	last := s
	if last[len(last)-1] == 'i' || last[len(last)-1] == 'I' {
		last = last[:len(last)-1]
	} else {
		idx := strings.LastIndexAny(last, "iI")
		if idx >= 0 {
			last = last[:idx]
		}
	}

	split := -1
	for i := len(last) - 1; i >= 1; i-- {
		c := last[i]
		if c == '+' || c == '-' {
			p := last[i-1]
			if p != 'e' && p != 'E' {
				split = i
				break
			}
		}
	}

	var rs, is string
	if split == -1 {
		rs = "0"
		is = last
	} else {
		rs = last[:split]
		is = last[split:]
	}

	r, err := strconv.ParseFloat(strings.TrimSpace(rs), bits)
	if err != nil {
		return 0, err
	}
	i, err := strconv.ParseFloat(strings.TrimSpace(is), bits)
	if err != nil {
		return 0, err
	}
	return complex(r, i), nil
}

package convert

import (
	"encoding"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// ConvertFunc converts a raw cell string to a typed value.
type ConvertFunc func(raw string) (any, error)

// Registry holds named converters registered for a read operation.
type Registry map[string]ConvertFunc

// BoolTrueValues are accepted as true when parsing bool cells.
var BoolTrueValues = map[string]struct{}{
	"true": {}, "1": {}, "是": {}, "yes": {}, "y": {},
}

// BoolFalseValues are accepted as false when parsing bool cells.
var BoolFalseValues = map[string]struct{}{
	"false": {}, "0": {}, "否": {}, "no": {}, "n": {},
}

var commonTimeLayouts = []string{
	time.RFC3339,
	"2006-01-02 15:04:05",
	"2006-01-02",
	"2006/01/02",
	"01-02-06",
	"2006/1/2",
}

// To assigns raw into dst based on dst's type, optional named converter, and time layout.
func To(raw string, dst reflect.Value, named string, timeLayout string, reg Registry) error {
	raw = strings.TrimSpace(raw)
	if !dst.CanSet() {
		return fmt.Errorf("convert: cannot set field")
	}
	if named != "" {
		fn, ok := reg[named]
		if !ok {
			return fmt.Errorf("convert: unknown converter %q", named)
		}
		v, err := fn(raw)
		if err != nil {
			return err
		}

		return assignValue(dst, v)
	}

	return builtin(raw, dst, timeLayout)
}

func builtin(raw string, dst reflect.Value, timeLayout string) error {
	if dst.Kind() == reflect.Ptr {
		if raw == "" {
			dst.Set(reflect.Zero(dst.Type()))

			return nil
		}
		if dst.IsNil() {
			dst.Set(reflect.New(dst.Type().Elem()))
		}

		return builtin(raw, dst.Elem(), timeLayout)
	}
	switch dst.Kind() {
	case reflect.String:
		dst.SetString(raw)

		return nil
	case reflect.Bool:
		v, err := parseBool(raw)
		if err != nil {
			return err
		}
		dst.SetBool(v)

		return nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if raw == "" {
			dst.SetInt(0)

			return nil
		}
		n, err := strconv.ParseInt(raw, 10, dst.Type().Bits())
		if err != nil {
			return fmt.Errorf("convert: invalid integer %q: %w", raw, err)
		}
		dst.SetInt(n)

		return nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if raw == "" {
			dst.SetUint(0)

			return nil
		}
		n, err := strconv.ParseUint(raw, 10, dst.Type().Bits())
		if err != nil {
			return fmt.Errorf("convert: invalid unsigned integer %q: %w", raw, err)
		}
		dst.SetUint(n)

		return nil
	case reflect.Float32, reflect.Float64:
		if raw == "" {
			dst.SetFloat(0)

			return nil
		}
		n, err := strconv.ParseFloat(raw, dst.Type().Bits())
		if err != nil {
			return fmt.Errorf("convert: invalid float %q: %w", raw, err)
		}
		dst.SetFloat(n)

		return nil
	case reflect.Struct:
		if dst.Type() == reflect.TypeOf(time.Time{}) {
			t, err := parseTime(raw, timeLayout)
			if err != nil {
				return err
			}
			dst.Set(reflect.ValueOf(t))

			return nil
		}
	}
	if dst.CanAddr() {
		if u, ok := dst.Addr().Interface().(encoding.TextUnmarshaler); ok {
			return u.UnmarshalText([]byte(raw))
		}
	}

	return fmt.Errorf("convert: unsupported type %s", dst.Type())
}

func parseBool(raw string) (bool, error) {
	if raw == "" {
		return false, nil
	}
	lower := strings.ToLower(raw)
	if _, ok := BoolTrueValues[lower]; ok {
		return true, nil
	}
	if _, ok := BoolFalseValues[lower]; ok {
		return false, nil
	}

	return false, fmt.Errorf("convert: invalid bool %q", raw)
}

func parseTime(raw, layout string) (time.Time, error) {
	if raw == "" {
		return time.Time{}, nil
	}
	if layout != "" {
		t, err := time.ParseInLocation(layout, raw, time.Local)
		if err != nil {
			return time.Time{}, fmt.Errorf("convert: invalid time %q with layout %q: %w", raw, layout, err)
		}

		return t, nil
	}
	for _, l := range commonTimeLayouts {
		if t, err := time.ParseInLocation(l, raw, time.Local); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("convert: invalid time %q", raw)
}

func assignValue(dst reflect.Value, v any) error {
	if v == nil {
		dst.Set(reflect.Zero(dst.Type()))

		return nil
	}
	rv := reflect.ValueOf(v)
	if !rv.IsValid() {
		dst.Set(reflect.Zero(dst.Type()))

		return nil
	}
	if rv.Type().AssignableTo(dst.Type()) {
		dst.Set(rv)

		return nil
	}
	if rv.Type().ConvertibleTo(dst.Type()) {
		dst.Set(rv.Convert(dst.Type()))

		return nil
	}

	return fmt.Errorf("convert: cannot assign %T to %s", v, dst.Type())
}

// ConvertTo registers a typed named converter helper.
func ConvertTo[T any](reg Registry, name string, fn func(string) (T, error)) {
	reg[name] = func(raw string) (any, error) {
		return fn(raw)
	}
}

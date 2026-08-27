package convert

import (
	"encoding"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

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

// To assigns raw into dst based on dst's type and optional time layout.
func To(raw string, dst reflect.Value, timeLayout string) error {
	raw = strings.TrimSpace(raw)
	if !dst.CanSet() {
		return fmt.Errorf("convert: cannot set field")
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
	if dst.Kind() == reflect.Struct && dst.Type() == reflect.TypeOf(time.Time{}) {
		t, err := parseTime(raw, timeLayout)
		if err != nil {
			return err
		}
		dst.Set(reflect.ValueOf(t))

		return nil
	}
	if u, ok := textUnmarshaler(dst); ok {
		return u.UnmarshalText([]byte(raw))
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
		return fmt.Errorf("convert: unsupported type %s", dst.Type())
	}

	return fmt.Errorf("convert: unsupported type %s", dst.Type())
}

func textUnmarshaler(v reflect.Value) (encoding.TextUnmarshaler, bool) {
	if v.CanAddr() {
		if u, ok := v.Addr().Interface().(encoding.TextUnmarshaler); ok {
			return u, true
		}
	}
	if v.CanInterface() {
		if u, ok := v.Interface().(encoding.TextUnmarshaler); ok {
			return u, true
		}
	}

	return nil, false
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

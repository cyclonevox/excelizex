package convert

import (
	"fmt"
	"reflect"
	"strconv"
	"time"
)

// ExportFunc converts a typed field value to a cell string.
type ExportFunc func(any) (string, error)

// ExportRegistry holds named export converters for write operations.
type ExportRegistry map[string]ExportFunc

// From formats v as a cell string using optional named converter and time layout.
func From(v reflect.Value, named string, timeLayout string, reg ExportRegistry) (string, error) {
	if named != "" {
		fn, ok := reg[named]
		if !ok {
			return "", fmt.Errorf("convert: unknown export converter %q", named)
		}
		var raw any
		if v.IsValid() {
			raw = v.Interface()
		}

		return fn(raw)
	}

	return builtinExport(v, timeLayout)
}

func builtinExport(v reflect.Value, timeLayout string) (string, error) {
	if !v.IsValid() {
		return "", nil
	}
	for v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return "", nil
		}
		v = v.Elem()
	}
	switch v.Kind() {
	case reflect.String:
		return v.String(), nil
	case reflect.Bool:
		if v.Bool() {
			return "是", nil
		}

		return "否", nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(v.Int(), 10), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(v.Uint(), 10), nil
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(v.Float(), 'f', -1, v.Type().Bits()), nil
	case reflect.Struct:
		if v.Type() == reflect.TypeOf(time.Time{}) {
			t := v.Interface().(time.Time)
			if t.IsZero() {
				return "", nil
			}
			layout := timeLayout
			if layout == "" {
				layout = "2006-01-02"
			}

			return t.Format(layout), nil
		}
	}

	return "", fmt.Errorf("convert: unsupported export type %s", v.Type())
}

// ExportTo registers a typed named export converter.
func ExportTo[T any](reg ExportRegistry, name string, fn func(T) (string, error)) {
	reg[name] = func(v any) (string, error) {
		if v == nil {
			var zero T

			return fn(zero)
		}
		rv := reflect.ValueOf(v)
		for rv.Kind() == reflect.Ptr {
			if rv.IsNil() {
				var zero T

				return fn(zero)
			}
			rv = rv.Elem()
		}
		t, ok := rv.Interface().(T)
		if !ok {
			return "", fmt.Errorf("convert: export type mismatch for %q", name)
		}

		return fn(t)
	}
}

// ExportToString is a helper for converters that only need the string form of a value.
func ExportToString(reg ExportRegistry, name string, fn func(string) (string, error)) {
	reg[name] = func(v any) (string, error) {
		s := fmt.Sprint(v)
		if v == nil {
			s = ""
		}

		return fn(s)
	}
}

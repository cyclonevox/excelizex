package convert

import (
	"encoding"
	"fmt"
	"reflect"
	"strconv"
	"time"
)

// From formats v as a cell string using builtin rules and optional time layout.
func From(v reflect.Value, timeLayout string) (string, error) {
	if !v.IsValid() {
		return "", nil
	}
	for v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return "", nil
		}
		v = v.Elem()
	}
	if v.Kind() == reflect.Struct && v.Type() == reflect.TypeOf(time.Time{}) {
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
	if m, ok := textMarshaler(v); ok {
		b, err := m.MarshalText()
		if err != nil {
			return "", err
		}

		return string(b), nil
	}

	return builtinExport(v, timeLayout)
}

func textMarshaler(v reflect.Value) (encoding.TextMarshaler, bool) {
	if v.CanInterface() {
		if m, ok := v.Interface().(encoding.TextMarshaler); ok {
			return m, true
		}
	}
	if v.CanAddr() {
		if m, ok := v.Addr().Interface().(encoding.TextMarshaler); ok {
			return m, true
		}
	}

	return nil, false
}

func builtinExport(v reflect.Value, timeLayout string) (string, error) {
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
	}

	return "", fmt.Errorf("convert: unsupported export type %s", v.Type())
}

package bind

import (
	"fmt"
	"reflect"

	"github.com/cyclonevox/excelizex/convert"
	"github.com/cyclonevox/excelizex/schema"
)

// ExportRow formats one struct instance into cell strings aligned to schema column order.
// extraHeaders, if given, are appended from the extend map (excel:",extend").
func ExportRow[T any](sc schema.Schema, row T, extraHeaders ...string) ([]string, error) {
	v := reflect.ValueOf(row)
	for v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil, fmt.Errorf("bind: nil row")
		}
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return nil, fmt.Errorf("bind: expected struct row, got %s", v.Kind())
	}
	out := make([]string, len(sc.Columns)+len(extraHeaders))
	var catchAll ExcelFieldExporter
	if ptr := rowPointer(v); ptr.IsValid() {
		if ca, ok := ptr.Interface().(ExcelFieldExporter); ok {
			catchAll = ca
		}
	}
	for i, col := range sc.Columns {
		if s, handled, err := tryExcelExport(v, col.FieldPath, col.Header, catchAll); handled {
			if err != nil {
				return nil, fmt.Errorf("%s: %w", col.Header, err)
			}
			out[i] = s

			continue
		}
		field, err := fieldByPath(v, col.FieldPath)
		if err != nil {
			return nil, err
		}
		s, err := convert.From(field, col.TimeLayout)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", col.Header, err)
		}
		out[i] = s
	}
	if len(extraHeaders) == 0 {
		return out, nil
	}
	ext, err := extendMapValue(sc, v)
	if err != nil {
		return nil, err
	}
	for i, h := range extraHeaders {
		out[len(sc.Columns)+i] = ext[h]
	}

	return out, nil
}

// ExtendMap returns the row's excel:",extend" map, or nil if the schema has none.
func ExtendMap[T any](sc schema.Schema, row T) (map[string]string, error) {
	if sc.ExtendField == "" {
		return nil, nil
	}
	v := reflect.ValueOf(row)
	for v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil, nil
		}
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return nil, fmt.Errorf("bind: expected struct row, got %s", v.Kind())
	}

	return extendMapValue(sc, v)
}

func extendMapValue(sc schema.Schema, v reflect.Value) (map[string]string, error) {
	if sc.ExtendField == "" {
		return nil, nil
	}
	field, err := fieldByPath(v, sc.ExtendField)
	if err != nil {
		return nil, err
	}
	for field.Kind() == reflect.Ptr {
		if field.IsNil() {
			return nil, nil
		}
		field = field.Elem()
	}
	if field.Kind() != reflect.Map || field.IsNil() {
		return nil, nil
	}
	out, ok := field.Interface().(map[string]string)
	if !ok {
		return nil, fmt.Errorf("bind: extend field %s is not map[string]string", sc.ExtendField)
	}

	return out, nil
}

package bind

import (
	"fmt"
	"reflect"

	"github.com/cyclonevox/excelizex/v2/convert"
	"github.com/cyclonevox/excelizex/v2/schema"
)

// ExportRow formats one struct instance into cell strings aligned to schema column order.
func ExportRow[T any](sc schema.Schema, row T, reg convert.ExportRegistry) ([]string, error) {
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
	out := make([]string, len(sc.Columns))
	for i, col := range sc.Columns {
		field, err := fieldByPath(v, col.FieldPath)
		if err != nil {
			return nil, err
		}
		s, err := convert.From(field, col.Convert, col.TimeLayout, reg)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", col.Header, err)
		}
		out[i] = s
	}

	return out, nil
}

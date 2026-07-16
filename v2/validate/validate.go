package validate

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/cyclonevox/excelizex/v2/schema"
)

// Validator validates a bound row value.
type Validator interface {
	ValidateRow(row any) error
}

// Required validates schema columns tagged validate:"required".
type Required struct{}

func (Required) ValidateRow(row any) error {
	t, v, err := rowValue(row)
	if err != nil {
		return err
	}
	sc, err := schema.FromType(t)
	if err != nil {
		return err
	}
	for _, col := range sc.Columns {
		if col.Validate != "required" {
			continue
		}
		val, err := fieldByPath(v, col.FieldPath)
		if err != nil {
			return err
		}
		if isZero(val) {
			return fmt.Errorf("validate: %s is required", col.Header)
		}
	}

	return nil
}

// FieldValidator validates individual fields by header name.
type FieldValidator func(header string, value any) error

func (fn FieldValidator) ValidateRow(row any) error {
	t, v, err := rowValue(row)
	if err != nil {
		return err
	}
	sc, err := schema.FromType(t)
	if err != nil {
		return err
	}
	for _, col := range sc.Columns {
		val, err := fieldByPath(v, col.FieldPath)
		if err != nil {
			return err
		}
		if err := fn(col.Header, val.Interface()); err != nil {
			return err
		}
	}

	return nil
}

func rowValue(row any) (reflect.Type, reflect.Value, error) {
	if row == nil {
		return nil, reflect.Value{}, fmt.Errorf("validate: nil row")
	}
	v := reflect.ValueOf(row)
	t := v.Type()
	for t.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil, reflect.Value{}, fmt.Errorf("validate: nil row pointer")
		}
		v = v.Elem()
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return nil, reflect.Value{}, fmt.Errorf("validate: expected struct row")
	}

	return t, v, nil
}

func fieldByPath(v reflect.Value, path string) (reflect.Value, error) {
	parts := strings.Split(path, ".")
	cur := v
	for _, p := range parts {
		if p == "" {
			continue
		}
		if cur.Kind() == reflect.Ptr {
			if cur.IsNil() {
				cur = reflect.New(cur.Type().Elem())
			}
			cur = cur.Elem()
		}
		if cur.Kind() != reflect.Struct {
			return reflect.Value{}, fmt.Errorf("validate: invalid path %q", path)
		}
		cur = cur.FieldByName(p)
		if !cur.IsValid() {
			return reflect.Value{}, fmt.Errorf("validate: unknown field %q in path %q", p, path)
		}
	}

	return cur, nil
}

func isZero(v reflect.Value) bool {
	if !v.IsValid() {
		return true
	}
	switch v.Kind() {
	case reflect.String:
		return strings.TrimSpace(v.String()) == ""
	case reflect.Slice, reflect.Map:
		return v.IsNil() || v.Len() == 0
	case reflect.Ptr, reflect.Interface:
		return v.IsNil()
	default:
		return v.IsZero()
	}
}

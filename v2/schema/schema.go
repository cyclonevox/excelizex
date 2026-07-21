package schema

import (
	"fmt"
	"reflect"
)

// Column describes one logical Excel column derived from struct tags.
type Column struct {
	Header     string
	FieldPath  string
	Convert    string
	Validate   string
	Style      []string
	TimeLayout string
}

// Schema is the logical column model; it does not know physical header rows.
type Schema struct {
	Columns []Column
	Notice  string
}

// New builds a Schema from a struct value or pointer.
func New(v any) (Schema, error) {
	t, err := typeOf(v)
	if err != nil {
		return Schema{}, err
	}

	return FromType(t)
}

// FromType builds a Schema from a struct type.
func FromType(t reflect.Type) (Schema, error) {
	if t == nil {
		return Schema{}, fmt.Errorf("schema: nil type")
	}
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return Schema{}, fmt.Errorf("schema: expected struct, got %s", t.Kind())
	}
	var s Schema
	if err := walkStruct(t, "", &s); err != nil {
		return Schema{}, err
	}
	if err := checkDuplicateHeaders(s); err != nil {
		return Schema{}, err
	}

	return s, nil
}

// ColumnByHeader returns the column with the given header, if any.
func (s Schema) ColumnByHeader(header string) (Column, bool) {
	for _, c := range s.Columns {
		if c.Header == header {
			return c, true
		}
	}

	return Column{}, false
}

func checkDuplicateHeaders(s Schema) error {
	seen := make(map[string]string, len(s.Columns))
	for _, c := range s.Columns {
		if prev, ok := seen[c.Header]; ok {
			return fmt.Errorf("schema: duplicate header %q (%s and %s)", c.Header, prev, c.FieldPath)
		}
		seen[c.Header] = c.FieldPath
	}

	return nil
}

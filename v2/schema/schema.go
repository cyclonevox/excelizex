package schema

import "reflect"
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
	var s Schema
	if err := walkStruct(t, "", &s); err != nil {
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

// RequiredHeaders returns headers tagged with validate:"required".
func (s Schema) RequiredHeaders() []string {
	var out []string
	for _, c := range s.Columns {
		if c.Validate == "required" {
			out = append(out, c.Header)
		}
	}

	return out
}

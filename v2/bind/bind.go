package bind

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/cyclonevox/excelizex/v2/convert"
	"github.com/cyclonevox/excelizex/v2/schema"
)

// Mapping binds logical schema columns to physical column indexes.
type Mapping struct {
	Entries []Entry
	Headers []string
}

// Entry is one bound column.
type Entry struct {
	ColIndex int
	Column   schema.Column
}

// MatchColumns maps schema headers to physical column indexes.
func MatchColumns(sc schema.Schema, headers map[int]string) (Mapping, error) {
	headerToIndex := make(map[string]int, len(headers))
	seen := make(map[int]struct{})
	maxIdx := -1
	for idx, h := range headers {
		if _, dup := headerToIndex[h]; dup {
			return Mapping{}, fmt.Errorf("bind: duplicate header %q", h)
		}
		headerToIndex[h] = idx
		seen[idx] = struct{}{}
		if idx > maxIdx {
			maxIdx = idx
		}
	}

	var entries []Entry
	for _, col := range sc.Columns {
		idx, ok := headerToIndex[col.Header]
		if !ok {
			return Mapping{}, fmt.Errorf("bind: missing column %q", col.Header)
		}
		entries = append(entries, Entry{ColIndex: idx, Column: col})
	}

	return Mapping{Entries: entries, Headers: orderedHeaders(headers, maxIdx)}, nil
}

func orderedHeaders(headers map[int]string, maxIdx int) []string {
	out := make([]string, maxIdx+1)
	for i := 0; i <= maxIdx; i++ {
		out[i] = headers[i]
	}

	return out
}

// BindRow converts aligned cell values into a new instance of T.
func BindRow[T any](m Mapping, cells []string, reg convert.Registry) (T, error) {
	var zero T
	t := reflect.TypeOf(zero)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	v := reflect.New(t)
	if err := bindValue(v.Elem(), m, cells, reg); err != nil {
		return zero, err
	}
	if reflect.TypeOf(zero).Kind() == reflect.Ptr {
		return v.Interface().(T), nil
	}

	return v.Elem().Interface().(T), nil
}

func bindValue(dst reflect.Value, m Mapping, cells []string, reg convert.Registry) error {
	for _, e := range m.Entries {
		raw := cellAt(cells, e.ColIndex)
		field, err := fieldByPath(dst, e.Column.FieldPath)
		if err != nil {
			return err
		}
		if err := convert.To(raw, field, e.Column.Convert, e.Column.TimeLayout, reg); err != nil {
			return fmt.Errorf("%s: %w", e.Column.Header, err)
		}
	}

	return nil
}

func cellAt(cells []string, idx int) string {
	if idx < 0 || idx >= len(cells) {
		return ""
	}

	return cells[idx]
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
				cur.Set(reflect.New(cur.Type().Elem()))
			}
			cur = cur.Elem()
		}
		if cur.Kind() != reflect.Struct {
			return reflect.Value{}, fmt.Errorf("bind: invalid path %q", path)
		}
		cur = cur.FieldByName(p)
		if !cur.IsValid() {
			return reflect.Value{}, fmt.Errorf("bind: unknown field %q in path %q", p, path)
		}
	}

	return cur, nil
}

// ExtraHeaders returns headers present in the sheet but not in schema.
func ExtraHeaders(m Mapping, headers map[int]string, sc schema.Schema) []string {
	schemaSet := make(map[string]struct{}, len(sc.Columns))
	for _, c := range sc.Columns {
		schemaSet[c.Header] = struct{}{}
	}
	var extra []string
	for _, h := range headers {
		if h == "" {
			continue
		}
		if _, ok := schemaSet[h]; !ok {
			extra = append(extra, h)
		}
	}

	return extra
}

package bind

import (
	"fmt"
	"reflect"

	"github.com/cyclonevox/excelizex/convert"
	"github.com/cyclonevox/excelizex/schema"
)

// Mapping binds logical schema columns to physical column indexes.
type Mapping struct {
	Entries     []Entry
	Headers     []string
	Extra       []Entry
	ExtendField string
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
		delete(seen, idx)
	}
	var extra []Entry
	if sc.ExtendField != "" {
		for i := 0; i <= maxIdx; i++ {
			if _, used := seen[i]; !used {
				continue
			}
			h := headers[i]
			if h == "" {
				continue
			}
			extra = append(extra, Entry{ColIndex: i, Column: schema.Column{Header: h}})
		}
	}

	return Mapping{
		Entries:     entries,
		Headers:     orderedHeaders(headers, maxIdx),
		Extra:       extra,
		ExtendField: sc.ExtendField,
	}, nil
}

func orderedHeaders(headers map[int]string, maxIdx int) []string {
	out := make([]string, maxIdx+1)
	for i := 0; i <= maxIdx; i++ {
		out[i] = headers[i]
	}

	return out
}

// BindRow converts aligned cell values into a new instance of T.
func BindRow[T any](m Mapping, cells []string) (T, error) {
	var zero T
	t := reflect.TypeOf(zero)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	v := reflect.New(t)
	var catchAll ExcelFieldImporter
	if ca, ok := v.Interface().(ExcelFieldImporter); ok {
		catchAll = ca
	}
	if err := bindValue(v.Elem(), m, cells, catchAll); err != nil {
		return zero, err
	}
	if reflect.TypeOf(zero).Kind() == reflect.Ptr {
		return v.Interface().(T), nil
	}

	return v.Elem().Interface().(T), nil
}

func bindValue(dst reflect.Value, m Mapping, cells []string, catchAll ExcelFieldImporter) error {
	for _, e := range m.Entries {
		raw := cellAt(cells, e.ColIndex)
		handled, err := tryExcelImport(dst, e.Column.FieldPath, e.Column.Header, raw, catchAll)
		if err != nil {
			return fieldAssignError(e.Column.Header, err)
		}
		if handled {
			continue
		}
		field, err := fieldByPath(dst, e.Column.FieldPath)
		if err != nil {
			return err
		}
		if err := convert.To(raw, field, e.Column.TimeLayout); err != nil {
			return fieldAssignError(e.Column.Header, err)
		}
	}
	if m.ExtendField == "" {
		return nil
	}
	field, err := fieldByPath(dst, m.ExtendField)
	if err != nil {
		return err
	}
	ext := make(map[string]string, len(m.Extra))
	for _, e := range m.Extra {
		ext[e.Column.Header] = cellAt(cells, e.ColIndex)
	}
	if !field.CanSet() {
		return fmt.Errorf("bind: cannot set extend field %s", m.ExtendField)
	}
	field.Set(reflect.ValueOf(ext))

	return nil
}

func cellAt(cells []string, idx int) string {
	if idx < 0 || idx >= len(cells) {
		return ""
	}

	return cells[idx]
}

// ExtraHeaders returns sheet headers that were not in the schema (extend columns).
func ExtraHeaders(m Mapping) []string {
	out := make([]string, len(m.Extra))
	for i, e := range m.Extra {
		out[i] = e.Column.Header
	}

	return out
}

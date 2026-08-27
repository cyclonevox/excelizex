package schema

import (
	"fmt"
	"reflect"
	"strings"
)

func typeOf(v any) (reflect.Type, error) {
	if v == nil {
		return nil, fmt.Errorf("schema: nil value")
	}
	t := reflect.TypeOf(v)
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return nil, fmt.Errorf("schema: expected struct, got %s", t.Kind())
	}

	return t, nil
}

func walkStruct(t reflect.Type, prefix string, s *Schema) error {
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		ft := derefType(f.Type)
		if !f.IsExported() && !(f.Anonymous && ft.Kind() == reflect.Struct) {
			continue
		}
		excelTag := f.Tag.Get("excel")
		if excelTag == "-" {
			continue
		}
		if excelTag == "notice" {
			if s.Notice != "" {
				return fmt.Errorf("schema: duplicate notice tag on %s", f.Name)
			}
			s.Notice = f.Name
			continue
		}
		if shouldInline(f, excelTag) {
			if ft.Kind() != reflect.Struct {
				return fmt.Errorf("schema: inline field %s must be a struct (or *struct), got %s", f.Name, f.Type)
			}
			subPrefix := prefix
			if !f.Anonymous {
				subPrefix = prefix + f.Name + "."
			} else if prefix == "" {
				subPrefix = ""
			}
			if err := walkStruct(ft, subPrefix, s); err != nil {
				return err
			}
			continue
		}
		if excelTag == "" {
			continue
		}
		header := strings.TrimSpace(excelTag)
		if header == "" || strings.HasPrefix(header, ",") {
			return fmt.Errorf("schema: invalid excel tag on field %s", f.Name)
		}
		col := Column{
			Header:     header,
			FieldPath:  prefix + f.Name,
			Validate:   f.Tag.Get("validate"),
			TimeLayout: f.Tag.Get("time"),
		}
		if style := f.Tag.Get("style"); style != "" {
			col.Style = strings.Split(style, ",")
			for i := range col.Style {
				col.Style[i] = strings.TrimSpace(col.Style[i])
			}
		}
		s.Columns = append(s.Columns, col)
	}

	return nil
}

func derefType(t reflect.Type) reflect.Type {
	for t != nil && t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	return t
}

func shouldInline(f reflect.StructField, excelTag string) bool {
	ft := derefType(f.Type)
	if f.Anonymous && ft.Kind() == reflect.Struct {
		return excelTag == "" || excelTag == ",inline"
	}

	return excelTag == ",inline"
}

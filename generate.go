package excelizex

import (
	"errors"
	"reflect"
)

func gen(a any, name ...string) (Sheet Sheet) {
	if a == nil {
		return
	}
	typ := reflect.TypeOf(a)

	if typ.Kind() != reflect.Struct {
		panic(errors.New("generate function support using struct only"))
	}

	if len(name) != 0 {
		Sheet.Name = name[0]
	}

	for i := 0; i < typ.NumField(); i++ {
		typeField := typ.Field(i)

		headerName := typeField.Tag.Get("excel")
		if headerName == "" {
			continue
		} else {
			Sheet.Header = append(Sheet.Header, headerName)
		}
	}

	return
}

// Gen can use input slice variable generate sheet
func Gen(slice any, name ...string) (Sheet Sheet) {
	typ := reflect.TypeOf(slice)
	val := reflect.ValueOf(slice)

	if typ.Kind() != reflect.Slice {
		panic(errors.New("generate function support using struct only"))
	}

	for i := 0; i < val.Len(); i++ {
		single := val.Index(i) // Value of item
		singleTyp := single.Type()

		if i == 0 {
			Sheet = gen(single.Interface(), name...)
		}

		if single.Kind() != reflect.Struct {
			panic(errors.New("generate function support using struct payload slice only"))
		}

		list := make([]interface{}, 0)
		for j := 0; j < singleTyp.NumField(); j++ {
			field := singleTyp.Field(i)

			hasTag := field.Tag.Get("excel")
			if hasTag != "" {
				list = append(list, single.Field(j).Interface())
			}
		}

		Sheet.Data = append(Sheet.Data, list)
	}

	return
}

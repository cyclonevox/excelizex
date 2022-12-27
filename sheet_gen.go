package excelizex

import (
	"errors"
	"reflect"
)

func genSheet(a any, name ...string) (Sheet Sheet) {
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

	var headers []string
	for i := 0; i < typ.NumField(); i++ {
		typeField := typ.Field(i)

		headerName := typeField.Tag.Get("excel")
		if headerName == "" {
			continue
		} else {
			headers = append(headers, headerName)
		}
	}
	Sheet.SetHeader(headers, true)

	return
}

func genSingleData(single any) (list []any) {
	typ := reflect.TypeOf(single)
	val := reflect.ValueOf(single)

	if typ.Kind() != reflect.Struct {
		panic(errors.New("generate function support using struct payload single only"))
	}

	for j := 0; j < typ.NumField(); j++ {
		field := typ.Field(j)

		hasTag := field.Tag.Get("excel")
		if hasTag != "" {
			list = append(list, val.Field(j).Interface())
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
		if i == 0 {
			Sheet = genSheet(val.Index(i).Interface(), name...)
		}

		Sheet.Data = append(Sheet.Data, genSingleData(val.Index(i).Interface()))
	}

	return
}

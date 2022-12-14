package excelizex

import (
	"errors"
	"reflect"
)

func singleRowData(single any) (list []any) {
	typ := reflect.TypeOf(single)
	val := reflect.ValueOf(single)

	switch typ.Kind() {
	// 对于结构体会获取含有excel结构体的数值
	case reflect.Struct:
		for j := 0; j < typ.NumField(); j++ {
			field := typ.Field(j)

			hasTag := field.Tag.Get("excel")
			if hasTag != "" {
				list = append(list, val.Field(j).Interface())
			}
		}

	// 对于切片类型会直接转为[]any
	// 只支持int string float类型的切片
	case reflect.Slice:
		value := reflect.ValueOf(single)
		rsp := make([]interface{}, 0, value.Len())
		for i := 0; i < value.Len(); i++ {
			rsp = append(rsp, value.Index(i).Interface())
		}

		return rsp
	}

	return
}

// genDataSheet can use input slice to generate sheet
// This function just support simple data sheet
func genDataSheet(slice any, option ...SheetOption) (Sheet *Sheet) {
	if slice == nil {
		panic(errors.New("slice nil"))
	}

	typ := reflect.TypeOf(slice)
	val := reflect.ValueOf(slice)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
		val = val.Elem()
	}

	if reflect.ValueOf(slice).Kind() != reflect.Ptr || typ.Kind() != reflect.Slice {
		panic(errors.New("generate sheet function by Data support using Slice prt only"))
	}

	for i := 0; i < val.Len(); i++ {
		if i == 0 {
			Sheet = NewSheet(option...).SetHeaderByStruct(val.Index(i).Interface())
		}

		Sheet.Data = append(Sheet.Data, singleRowData(val.Index(i).Interface()))
	}

	return
}

package excelizex

import (
	"errors"
	"excelizex/validatorx"
	"github.com/xuri/excelize/v2"
	"io"
	"reflect"
	"strings"
)

type Func func(sliceSingle any) error

func GenFormFile(file io.Reader) *excel {
	var (
		e   excel
		err error
	)

	if e.file, err = excelize.OpenReader(file); err != nil {
		panic(err)
	}

	return &e
}

func (e *excel) GetData(sheetName string, model interface{}) (err error) {
	var rows *excelize.Rows
	if rows, err = e.getFile().Rows(sheetName); err != nil {
		return
	}

	typ := reflect.TypeOf(model)
	val := reflect.ValueOf(model)
	if typ.Kind() != reflect.Struct {
		panic(errors.New("generate function support Struct only"))
	}

	var (
		headers = make(map[string][]string)
	)
	for j := 0; j < typ.NumField(); j++ {
		field := val.Type().Field(j)

		hasTag := field.Tag.Get("excel")
		if hasTag != "" {
			validate := field.Tag.Get("validate")
			headers[hasTag] = strings.Split(validate, " ")
		}
	}

	var headerFound bool
	var validateMap = make(map[int][]string)
	for rows.Next() {
		var columns []string
		if columns, err = rows.Columns(); err != nil {
			return
		}

		if !headerFound {
			headerFound = reflect.DeepEqual(columns, headers)
			if headerFound {
				for index, col := range columns {
					validateMap[index] = headers[col]
				}
			}

			continue
		}

		for index, col := range columns {
			for _, vTag := range validateMap[index] {
				if err, _ = validatorx.New().Val(col, vTag); nil != err {
					continue
				}
			}
		}
	}

	return
}

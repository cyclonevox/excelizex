package excelizex

import (
	"errors"
	"excelizex/validatorx"
	"github.com/xuri/excelize/v2"
	"io"
	"reflect"
	"strings"
)

func ReadFormFile(file io.Reader) *excel {
	var (
		e   excel
		err error
	)

	if e._excel, err = excelize.OpenReader(file); err != nil {
		panic(err)
	}

	return &e
}

func (e *excel) GetData(sheetName string, model interface{}) (results []Result) {
	var (
		err  error
		rows *excelize.Rows
	)
	if rows, err = e.getFile().Rows(sheetName); err != nil {
		panic(err)
	}

	typ := reflect.TypeOf(model)
	val := reflect.ValueOf(model)
	if typ.Kind() != reflect.Struct {
		panic(errors.New("generate function support Struct only"))
	}

	var headers []string
	headerValidate := make(map[string][]string)
	for j := 0; j < typ.NumField(); j++ {
		field := val.Type().Field(j)

		hasTag := field.Tag.Get("excel")
		if hasTag != "" {
			headers = append(headers, hasTag)
			validate := field.Tag.Get("validate")
			headerValidate[hasTag] = strings.Split(validate, " ")
		}
	}

	var (
		row         int
		headerFound bool
		validateMap = make(map[int][]string)
	)
	for rows.Next() {
		row++
		var columns []string
		if columns, err = rows.Columns(); err != nil {
			panic(err)
		}

		if !headerFound {
			headerFound = reflect.DeepEqual(columns, headers)
			if headerFound {
				for index, col := range columns {
					validateMap[index] = headerValidate[col]
				}
			}

			continue
		}

		for index, col := range columns {
			var errInfo []string
			for _, vTag := range validateMap[index] {
				var trans string
				if err, trans = validatorx.New().Val(col, vTag); nil != err {
					errInfo = append(errInfo, headers[index]+trans)
				}

			}

			if len(errInfo) > 0 {
				results = append(results, Result{
					ErrorRow:     row,
					ErrorRowData: columns,
					ErrorInfo:    errInfo,
				})
			}
		}
	}

	return
}

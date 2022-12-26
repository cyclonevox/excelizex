package excelizex

import (
	"errors"
	"excelizex/validatorx"
	"github.com/xuri/excelize/v2"
	"io"
	"reflect"
	"strconv"
)

func ReadFormFile(reader io.Reader) *file {
	var (
		err error
		f   file
	)

	if f._excel, err = excelize.OpenReader(reader); err != nil {
		panic(err)
	}

	return &f
}

type Importable interface {
	ImportData() error
}

func (f *file) Import(sheetName string, data Importable) Result {
	var (
		results Result
		rows    *excelize.Rows
		err     error
	)
	if rows, err = f.excel().Rows(sheetName); err != nil {
		panic(err)
	}

	typ := reflect.TypeOf(data)
	val := reflect.ValueOf(data)
	if typ.Kind() != reflect.Struct {
		panic(errors.New("generate function support Struct only"))
	}

	// 获取结构体中应对的
	var (
		headers          []string
		dataConvert      = make(map[string]string)
		headerStructName = make(map[string]string)
	)

	for j := 0; j < typ.NumField(); j++ {
		field := val.Type().Field(j)

		hasTag := field.Tag.Get("file")
		if hasTag != "" {
			headers = append(headers, hasTag)

			convTag := field.Tag.Get("file-conv")
			if convTag != "" {
				headerStructName[hasTag] = convTag
			}
		}
	}

	var (
		row         int
		headerFound bool
	)
	for rows.Next() {
		row++
		var columns []string
		if columns, err = rows.Columns(); err != nil {
			panic(err)
		}

		// 寻找表头，并将行数与关联存于map作为缓存
		if !headerFound {
			headerFound = reflect.DeepEqual(columns, headers)
			results.Header = headers
			results.DataStartRow = row

			continue
		}

		// 将值加入结构体
		for index, col := range columns {
			field := val.Elem().FieldByName(headerStructName[headers[index]])

			if v, ok := dataConvert[col]; ok {
				if r := val.MethodByName(v).Call(nil); len(r) <= 1 {
					panic(errors.New("convert method call error"))
				} else {
					field.Set(r[0])
				}
			} else {
				switch field.Kind() {
				case reflect.Int64, reflect.Int32, reflect.Int8, reflect.Int16, reflect.Int:
					var i int64
					if i, err = strconv.ParseInt(col, 10, 64); nil != err {
						panic(i)
					}
					field.SetInt(i)
				case reflect.String:
					field.SetString(col)
				}
			}

		}

		if info := importData(data); len(info) > 0 {
			results.Errors = append(results.Errors, ErrorInfo{
				ErrorRow:  row,
				ErrorInfo: info,
			})

			continue
		}
	}

	results.MaxRow = row

	return results
}

func importData(data Importable) (errInfo []string) {
	// 验证结构体数据是否合法
	if err, m := validatorx.New().Struct(data); nil != err {
		for _, v := range m {
			errInfo = append(errInfo, v)
		}

		return
	}

	// 执行导入业务
	if err := data.ImportData(); err != nil {
		errInfo = append(errInfo, err.Error())

		return
	}

	return
}

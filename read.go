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

type ConvertFunc func(rawData string) (any, error)

// SetConvert 作为设置该excel 文件的转换器
// ConvertFunc 作为转换器函数 rawData 是指excel tag的对应的excel原始数据。
//
//	例如: type IdCard struct {
//			Id   string `excel:"名称" excel-convert:"nameId"`
//			...
//		 }
//
// 调用 f.SetConvert("nameId", func(rawData string) (any, error){ return raw[:3],nil}) 设置转换器以后
// 使用 excel-convert:"nameId" tag的字段都会在Read函数中调用相应的函数被转换
func (f *file) SetConvert(convertName string, convertFunc ConvertFunc) *file {
	if f.convert == nil {
		f.convert = make(map[string]ConvertFunc)
	}

	f.convert[convertName] = convertFunc

	return f
}

// SetConvertMap 可传入一个 key为convertName value为转换器函数的 map
// 以达到一次传入多个 ConvertFunc 的效果，具体使用说明可见 SetConvert 方法注释
func (f *file) SetConvertMap(convert map[string]ConvertFunc) *file {
	if f.convert == nil {
		f.convert = convert
	} else {
		for k, c := range convert {
			f.convert[k] = c
		}
	}

	return f
}

type ImportFunc func(any) error

func (f *file) Read(sheetName string, data any, fn ImportFunc) Result {
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

		hasTag := field.Tag.Get("excel")
		if hasTag != "" {
			headers = append(headers, hasTag)

			convTag := field.Tag.Get("excel-conv")
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
			results.dataStartRow = row

			continue
		}

		// 将值加入结构体
		for index, col := range columns {
			field := val.Elem().FieldByName(headerStructName[headers[index]])

			// 查看该字段是否有转换器
			if v, ok := dataConvert[col]; ok {
				var convertValue any
				if convertValue, err = f.convert[v](col); err != nil {
					results.Errors = append(results.Errors, ErrorInfo{
						ErrorRow:  row,
						ErrorInfo: []string{err.Error()},
					})

					continue
				}

				field = reflect.ValueOf(convertValue)
			}

			switch field.Kind() {
			case reflect.Float32, reflect.Float64:
				var i float64
				if i, err = strconv.ParseFloat(col, 64); nil != err {
					panic(i)
				}
				field.SetFloat(i)
			case reflect.Int64, reflect.Int32, reflect.Int8, reflect.Int16, reflect.Int:
				var i int64
				if i, err = strconv.ParseInt(col, 10, 64); nil != err {
					panic(i)
				}
				field.SetInt(i)
			case reflect.String:
				field.SetString(col)
			default:
				panic("cannot support other type besides int,float,string")
			}

		}

		if len(results.Errors) > 0 {
			continue
		}

		if info := importData(data, fn); len(info) > 0 {
			results.Errors = append(results.Errors, ErrorInfo{
				ErrorRow:  row,
				ErrorInfo: info,
			})

			continue
		}
	}

	return results
}

func importData(data any, fn ImportFunc) (errInfo []string) {
	// 验证结构体数据是否合法
	if err, m := validatorx.New().Struct(data); nil != err {
		for _, v := range m {
			errInfo = append(errInfo, v)
		}

		return
	}

	// 执行导入业务
	if err := fn(data); err != nil {
		errInfo = append(errInfo, err.Error())

		return
	}

	return
}

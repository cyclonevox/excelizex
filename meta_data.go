package excelizex

import (
	"reflect"
	"strconv"
)

type metaData struct {
	// 表名
	sheetName string
	// 存储header原始数据
	headers []string
	// k:columns index v:ExtHeader name
	colsIndexHeaderMap map[int]string
	// k:header v:Struct Field Name
	headerFieldName map[string]string
	// k:header v:convert name
	headerMappingConverter map[string]string
	// k: converter name v:converter func
	converter map[string]ConvertFunc
	// payload struct,the real val for run
	payload any
}

func newMetaData() *metaData {
	return &metaData{
		headers:                make([]string, 0),
		colsIndexHeaderMap:     make(map[int]string, 0),
		headerFieldName:        make(map[string]string),
		headerMappingConverter: make(map[string]string),
		converter:              make(map[string]ConvertFunc),
	}
}

func (e *metaData) addHeader(header string) {
	e.headers = append(e.headers, header)
}

func (e *metaData) addHeaders(headers []string) {
	e.headers = append(e.headers, headers...)
}

func (e *metaData) addHeaderFieldName(header, fieldName string) {
	e.headerFieldName[header] = fieldName
}

func (e *metaData) addHeaderConvertName(header, convertName string) {
	e.headerMappingConverter[header] = convertName
}

func (e *metaData) findHeadersMap(columns []string) (exist bool) {
	if len(columns) < len(e.headers) {
		return
	}

	for index, column := range columns {
		for _, h := range e.headers {
			if column == h {
				e.colsIndexHeaderMap[index] = column

				continue
			}
		}
	}

	if len(e.colsIndexHeaderMap) == len(e.headers) {
		exist = true
	} else {
		e.colsIndexHeaderMap = make(map[int]string, 0)
	}

	return
}

func (e *metaData) getHeader(columnIndex int) (header string) {
	return e.colsIndexHeaderMap[columnIndex]
}

func (e *metaData) findConvertByHeader(header string) (convertName string, exist bool) {
	convertName, exist = e.headerMappingConverter[header]

	return
}

func (e *metaData) getHeaderFieldName(columnIndex int) (header string) {
	return e.headerFieldName[e.getHeader(columnIndex)]
}

func (e *metaData) dataMapping(ptr any, columns []string) (err error) {
	for index, col := range columns {
		fieldName := e.getHeaderFieldName(index)
		if fieldName == "" {
			continue
		}
		field := reflect.ValueOf(ptr).Elem().FieldByName(fieldName)

		// 查看该字段是否有转换器
		if v, ok := e.findConvertByHeader(e.getHeader(index)); ok {
			var convertValue any
			if convertValue, err = e.converter[v](col); err != nil {
				return
			}

			field.Set(reflect.ValueOf(convertValue))

			continue
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

	return
}

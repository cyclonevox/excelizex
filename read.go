package excelizex

import (
	"encoding/json"
	"errors"
	"github.com/xuri/excelize/v2"
	"io"
	"reflect"
)

func ReadFormFile(reader io.Reader) *File {
	var (
		err error
		f   File
	)

	if f._excel, err = excelize.OpenReader(reader); err != nil {
		panic(err)
	}

	return &f
}

type ConvertFunc func(rawData string) (any, error)

type ImportFunc func() error

type Read struct {
	// sheet metadata
	metaData *metaData
	// sheet rows iterator
	rows *excelize.Rows
	// I don't like err to break the method call link. So I did it
	err error
}

func (f *File) Read(ptr any) (r *Read) {
	r = new(Read)

	var err error
	if r.rows, err = f.excel().Rows(f.selectSheetName); err != nil {
		r.err = err

		return
	}

	if err = r.newMetaData(ptr); err != nil {
		r.err = err

		return
	}

	// todo: Try reflect.New to generate this one's ptr payload
	r.metaData.payload = ptr

	return
}

func (r *Read) newMetaData(ptr any) (err error) {
	typ := reflect.TypeOf(ptr)
	val := reflect.ValueOf(ptr)

	if typ.Kind() != reflect.Pointer || typ.Elem().Kind() != reflect.Struct {
		err = errors.New("read function support struct type variable's Pointer type only")

		return
	}

	r.metaData = newMetaData()

	for j := 0; j < typ.Elem().NumField(); j++ {
		field := val.Elem().Type().Field(j)

		hasTag := field.Tag.Get("excel")
		if hasTag != "" {
			r.metaData.addHeader(hasTag)
			r.metaData.addHeaderFieldName(hasTag, field.Name)

			convTag := field.Tag.Get("excel-conv")
			if convTag != "" {
				r.metaData.addHeaderConvertName(hasTag, convTag)
			}
		}
	}

	return
}

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
func (r *Read) SetConvert(convertName string, convertFunc ConvertFunc) *Read {
	if r.metaData.converter == nil {
		r.metaData.converter = make(map[string]ConvertFunc)
	}

	r.metaData.converter[convertName] = convertFunc

	return r
}

// SetConvertMap 可传入一个 key为convertName value为转换器函数的 map
// 以达到一次传入多个 ConvertFunc 的效果，具体使用说明可见 SetConvert 方法注释
func (r *Read) SetConvertMap(convert map[string]ConvertFunc) *Read {
	if r.metaData.converter == nil {
		r.metaData.converter = convert
	} else {
		for k, c := range convert {
			r.metaData.converter[k] = c
		}
	}

	return r
}

func (r *Read) Run(fn ImportFunc) Result {
	if r.err != nil {
		panic(r.err.Error())
	}

	var (
		row         int
		err         error
		headerFound bool
		results     Result
	)

	for r.rows.Next() {
		row++
		var columns []string
		if columns, err = r.rows.Columns(); err != nil {
			panic(err)
		}

		// 寻找表头，并将行数与关联存于map作为缓存,并将关联的表存储进
		if !headerFound {
			headerFound = r.metaData.findHeadersMap(columns)
			if headerFound {
				results.Header = r.metaData.headers
				results.dataStartRow = row + 1
			}

			continue
		}

		// 将值映射入结构体
		if err = r.metaData.dataMapping(r.metaData.payload, columns); err != nil {
			results.addError(ErrorInfo{
				ErrorRow: row,
				RawData:  columns,
				Messages: []string{err.Error()},
			})

			continue
		}

		if info := importData(r.metaData.payload, fn); len(info) > 0 {
			results.addError(ErrorInfo{
				ErrorRow: row,
				RawData:  columns,
				Messages: info,
			})

			continue
		}
	}

	return results
}

func importData(data any, fn ImportFunc) (errInfo []string) {
	// 验证结构体数据是否合法
	if err := newValidate().Validate(data); nil != err {
		errInfo = append(errInfo, "该行有数据未正确填写")

		return
	}

	// 执行导入业务
	if err := fn(); err != nil {
		valid := json.Valid([]byte(err.Error()))

		if !valid {
			errInfo = append(errInfo, err.Error())
		} else {

			// don't ask why
			var e = struct {
				Message string `json:"message"`
			}{}

			_ = json.Unmarshal([]byte(err.Error()), &e)
			errInfo = append(errInfo, e.Message)
		}

		return
	}

	cleanData(data)

	return
}

func cleanData(ptr any) {
	ptrElemValue := reflect.ValueOf(ptr).Elem()
	num := ptrElemValue.NumField()

	for i := 0; i < num; i++ {
		ptrElemValue.Field(i).Set(reflect.New(ptrElemValue.Field(i).Type()).Elem())
	}
}

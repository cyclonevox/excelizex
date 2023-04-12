package excelizex

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/panjf2000/ants/v2"
	"github.com/xuri/excelize/v2"
	"reflect"
	"strconv"
	"sync"
)

type ConvertFunc func(rawData string) (any, error)

type ImportFunc func(any any) error

type Read struct {
	// sheet metadata
	metaData *metaData
	// sheet rows iterator
	rows *excelize.Rows
	// import func for business
	fn ImportFunc
	// Validate for mapping validation
	validates []Validate
	// results
	results *Result
	// I don't like err to break the method call link. So I did it
	err error

	// wg control to sync
	wg sync.WaitGroup
	//goroutine pool
	concPool *ants.Pool
	// reuse payload struct pool
	payloadPool *sync.Pool
}

func (f *File) Read(payload any, sheetName ...string) (r *Read) {
	r = new(Read)

	if f.selectSheetName == "" && len(sheetName) == 0 {
		panic("plz setting select sheet")
	}

	var sName string
	if len(sheetName) > 0 {
		sName = sheetName[0]
	} else {
		sName = f.selectSheetName
	}

	var err error
	if r.rows, err = f.excel().Rows(sName); err != nil {
		panic(err)
	}

	if err = r.newMetaData(payload); err != nil {
		r.err = err

		return
	}

	r.metaData.payload = payload
	r.results = new(Result)

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

func (r *Read) SetValidates(validate ...Validate) *Read {
	for _, v := range validate {
		r.validates = append(r.validates, v)
	}

	return r
}

func (r *Read) initPool(num ...int) (pool *ants.Pool, payloadPool *sync.Pool, err error) {
	// set up goroutine pool
	if len(num) > 0 {
		if num[0] < 1 {
			err = errors.New("num set less than 1")

			return
		}

		if pool, err = ants.NewPool(num[0]); err != nil {
			return
		}

	} else {
		if pool, err = ants.NewPool(1); err != nil {
			return
		}
	}

	// set up payloadPool
	payloadPool = &sync.Pool{New: func() any {
		return reflect.New(reflect.TypeOf(r.metaData.payload).Elem()).Interface()
	}}

	return
}

func (r *Read) setFunc(fn ImportFunc) {
	r.fn = fn
}

// Run is using to set excelizex.Read concurrency And execute business func.
// param fn is business func,fn's param is the struct object.
// param num for set execute business functions' goroutine num.
// attention: execute out of order temporarily.
func (r *Read) Run(fn ImportFunc, num ...int) (results *Result, err error) {
	if r.err != nil {
		err = r.err

		return
	}

	r.setFunc(fn)
	if r.concPool, r.payloadPool, err = r.initPool(num...); err != nil {
		return
	}
	defer r.concPool.Release()

	var (
		row         int
		headerFound bool
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
				r.results.Header = r.metaData.headers
				r.results.dataStartRow = row + 1
			}

			continue
		}

		r.wg.Add(1)
		// 向携程池提交任务
		if err = r.concPool.Submit(func() {
			r.exec(row, columns)
			r.wg.Done()
		}); err != nil {
			return
		}
	}

	r.wg.Wait()

	results = r.results

	return
}

func (r *Read) exec(row int, columns []string) {
	// 将值映射入结构体
	var (
		err  error
		data any
	)
	if data, err = r.dataMapping(columns); err != nil {
		r.results.addError(ErrorInfo{
			ErrorRow: row,
			RawData:  columns,
			Messages: []string{err.Error()},
		})

		return
	}

	if info := r.importData(data); len(info) > 0 {
		r.results.addError(ErrorInfo{
			ErrorRow: row,
			RawData:  columns,
			Messages: info,
		})
	}
}

func (r *Read) dataMapping(columns []string) (ptr any, err error) {
	ptr = r.payloadPool.Get()

	obj := reflect.ValueOf(ptr).Elem()

	for index, col := range columns {
		fieldName := r.metaData.getHeaderFieldName(index)
		if fieldName == "" {
			continue
		}
		field := obj.FieldByName(fieldName)

		// 查看该字段是否有转换器
		if v, ok := r.metaData.findConvertByHeader(r.metaData.getHeader(index)); ok {
			var (
				conv         ConvertFunc
				convertValue any
			)
			if conv, ok = r.metaData.converter[v]; ok {
				if convertValue, err = conv(col); err != nil {
					return
				}
			}

			if field.Type() != reflect.TypeOf(convertValue) {
				sprintf := fmt.Sprintf(
					"dataMapping error.convertor func return a wrong type.field type: %s;convertValue type: %s ",
					field.Type().String(), reflect.TypeOf(convertValue),
				)
				panic(sprintf)
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

func (r *Read) importData(data any) (errInfo []string) {
	// 如果设置了对数据结构体的验证方式 则会验证结构体数据是否合法
	for i := range r.validates {
		if err := r.validates[i].Validate(data); nil != err {
			errInfo = append(errInfo, "该行有数据未正确填写")

			return
		}
	}

	// 执行导入业务
	if err := r.fn(data); err != nil {
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

	r.cleanData(data)

	return
}

func (r *Read) cleanData(ptr any) {
	ptrElemValue := reflect.ValueOf(ptr).Elem()
	num := ptrElemValue.NumField()

	for i := 0; i < num; i++ {
		ptrElemValue.Field(i).Set(reflect.New(ptrElemValue.Field(i).Type()).Elem())
	}

	r.payloadPool.Put(ptr)
}

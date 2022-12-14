package excelizex

import (
	"errors"
	"reflect"
	"strconv"
)

type Sheet struct {
	// 表名
	Name string `json:"name"`
	// 顶栏提示
	Notice string `json:"notice"`
	// 表头
	Header []string `json:"header"`
	// 数据
	Data [][]any `json:"data"`
	// 写入到第几行,主要用于标记生成excel中的表时，需要续写的位置
	writeRow int
	// 下拉选项 暂时只支持单列
	pd *pullDown
}

func NewSheet(options ...SheetOption) *Sheet {
	sheet := new(Sheet)
	for _, option := range options {
		option(sheet)
	}

	return sheet
}

func NewDataSheet(slice any, options ...SheetOption) *Sheet {
	return genDataSheet(slice, options...)
}

func (s *Sheet) SetName(name string) *Sheet {
	s.Name = name

	return s
}

func (s *Sheet) SetNotice(notice string) *Sheet {
	s.Notice = notice

	return s
}

// SetHeader 为手动设置表的头部
func (s *Sheet) SetHeader(header []string) *Sheet {
	s.Header = header

	return s
}

// SetHeaderByStruct 方法会检测结构体中的excel标签，以获取结构体表头
func (s *Sheet) SetHeaderByStruct(a any) *Sheet {
	typ := reflect.TypeOf(a)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	if typ.Kind() != reflect.Struct {
		panic(errors.New("generate function support using struct only"))
	}

	for i := 0; i < typ.NumField(); i++ {
		typeField := typ.Field(i)

		headerName := typeField.Tag.Get("excel")
		if headerName == "" {
			continue
		} else {
			s.Header = append(s.Header, headerName)
		}
	}

	return s
}

func (s *Sheet) SetData(data [][]any) *Sheet {
	s.Data = data

	return s
}

// SetOptions 设置下拉的选项
func (s *Sheet) SetOptions(pds ...*pullDown) *Sheet {
	for _, pd := range pds {
		if s.pd == nil {
			s.pd = pd
		} else {
			s.pd.Merge(pd)
		}
	}

	return s
}

func (s *Sheet) Excel() *File {
	if s.Name == "" {
		panic("need a sheet name at least")
	}

	return New().AddSheets(s)
}

// writeRowIncr 会获取目前该写入的行
// 每次调用该方法表示行数增长 返回 A1 A2... 等名称
func (s *Sheet) writeRowIncr(num ...int) string {
	if len(num) > 0 {
		s.writeRow += num[0]
	} else {
		s.writeRow++
	}

	return "A" + strconv.FormatInt(int64(s.writeRow), 10)
}

type SheetOption = func(*Sheet)

func SetName(name string) SheetOption {
	return func(s *Sheet) {
		s.SetName(name)
	}
}

func SetHeader(header []string) SheetOption {
	return func(s *Sheet) {
		s.SetHeader(header)
	}
}

// HeaderByStruct 会根据结构体的tag来生成表头
func HeaderByStruct(a any) SheetOption {
	return func(s *Sheet) {
		s.SetHeaderByStruct(a)
	}
}

func Notice(notice string) SheetOption {
	return func(s *Sheet) {
		s.SetNotice(notice)
	}
}

// Data for 仅作为少量数据写入.如果需要写入大量数据 请使用StreamWriteIn() 以调用excelize的流式写入.
func Data(data [][]any) SheetOption {
	return func(s *Sheet) {
		s.SetData(data)
	}
}

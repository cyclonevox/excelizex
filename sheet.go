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

func (s *Sheet) SetName(name string) {
	s.Name = name
}

func (s *Sheet) SetNotice(notice string) {
	s.Notice = notice
}

// SetHeader 为手动设置表的头部
func (s *Sheet) SetHeader(header []string) {
	s.Header = header
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

func (s *Sheet) SetData(data [][]any) {
	s.Data = data
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
	return func(sheet *Sheet) {
		sheet.Name = name
	}
}

func SetHeader(header []string) SheetOption {
	return func(s *Sheet) {
		s.Header = header
	}
}

// SetHeaderByStruct 会根据结构体的tag来生成表头
func SetHeaderByStruct(a any) SheetOption {
	return func(s *Sheet) {
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
	}
}

func SetNotice(notice string) SheetOption {
	return func(s *Sheet) {
		s.Notice = notice
	}
}

// SetData for 仅作为少量数据写入.如果需要写入大量数据 请使用StreamWriteIn() 以调用excelize的流式写入.
func SetData(data [][]any) SheetOption {
	return func(s *Sheet) {
		s.Data = data
	}
}

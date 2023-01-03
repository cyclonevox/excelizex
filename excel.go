package excelizex

import (
	"bytes"
	"github.com/xuri/excelize/v2"
)

type file struct {
	_excel  *excelize.File
	convert map[string]ConvertFunc
}

func New() *file {
	return &file{_excel: excelize.NewFile()}
}

func (f *file) excel() *excelize.File {
	return f._excel
}

func (f *file) SaveAs(name string) (err error) {
	return f.excel().SaveAs(name)
}

func (f *file) Buffer() (*bytes.Buffer, error) {
	return f.excel().WriteToBuffer()
}

func (f *file) AddSheets(sheets ...*Sheet) *file {
	var err error

	for _, s := range sheets {
		if s.Name == "" {
			panic("need a sheet name at least")
		}
		f._excel.NewSheet(s.Name)

		if err = f.setDefaultFormatSheetAndStyle(s); err != nil {
			panic(err)
		}
	}

	return f
}

// AddDataSheet support use slice and their data generate a sheet with header and data
func (f *file) AddDataSheet(slice any, option ...SheetOption) *file {
	f.AddSheets(genDataSheet(slice, option...))

	return f
}

// AddSimpleSheet support use struct and their data generate a sheet with only header
func (f *file) AddSimpleSheet(a any, option ...SheetOption) *file {
	f.AddSheets(NewSheet(option...).SetHeaderByStruct(a))

	return f
}

func (f *file) setDefaultFormatSheetAndStyle(s *Sheet) (err error) {
	_excel := f.excel()

	// 设置表各列数据格式 数字默认为“文本”
	for i := range s.Header {
		var colName string
		if colName, err = excelize.ColumnNumberToName(1 + i); nil != err {
			return
		}

		if err = _excel.SetColStyle(s.Name, colName, f.StyleNumFmtText()); nil != err {
			return
		}
	}

	// 判断是否有提示并设置
	if s.Notice != "" {
		row := s.writeRowIncr()
		if err = _excel.SetCellValue(s.Name, row, s.Notice); err != nil {
			return
		}
		if err = _excel.SetCellStyle(s.Name, row, row, f.StyleRedTextLocked()); nil != err {
			return
		}
	}

	// 判断是否有提示并设置
	if len(s.Header) != 0 {
		row := s.writeRowIncr()
		if err = _excel.SetSheetRow(s.Name, row, &s.Header); err != nil {
			return
		}
		if err = _excel.SetRowStyle(s.Name, s.writeRow, s.writeRow, f.StyleLocked()); err != nil {
			return
		}
	}

	return
}

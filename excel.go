package excelizex

import (
	"bytes"
	"github.com/xuri/excelize/v2"
	"io"
)

type File struct {
	_excel  *excelize.File
	convert map[string]ConvertFunc
}

func New(reader ...io.Reader) *File {
	if len(reader) > 0 {
		if f, err := excelize.OpenReader(reader[0]); err != nil {
			panic(err)
		} else {
			return &File{_excel: f}
		}

	}

	return &File{_excel: excelize.NewFile()}
}

func (f *File) excel() *excelize.File {
	return f._excel
}

func (f *File) SaveAs(name string) (err error) {
	f.excel().DeleteSheet("Sheet1")

	return f.excel().SaveAs(name)
}

func (f *File) Buffer() (*bytes.Buffer, error) {
	f.excel().DeleteSheet("Sheet1")

	return f.excel().WriteToBuffer()
}

func (f *File) AddSheets(sheets ...*Sheet) *File {
	var err error

	for _, s := range sheets {
		if s.Name == "" || s.Name == "Sheet1" {
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
func (f *File) AddDataSheet(slicePtr any, option ...SheetOption) *File {
	f.AddSheets(genDataSheet(slicePtr, option...))

	return f
}

// AddSimpleSheet support use struct and their data generate a sheet with only header
func (f *File) AddSimpleSheet(a any, option ...SheetOption) *File {
	f.AddSheets(NewSheet(option...).SetHeaderByStruct(a))

	return f
}

func (f *File) setDefaultFormatSheetAndStyle(s *Sheet) (err error) {
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

package excelizex

import (
	"bytes"
	"fmt"
	"github.com/xuri/excelize/v2"
	"io"
	"strconv"
)

const OptionsSaveTable = "选项数据表"

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
		f.addSheet(s)

		if err = f.writeDefaultFormatSheet(s); err != nil {
			panic(err)
		}

		if err = f.setPullDown(s); err != nil {
			panic(err)
		}
	}

	return f
}

func (f *File) addSheet(sheets ...*Sheet) {
	for _, s := range sheets {
		if s.Name == "" || s.Name == "Sheet1" {
			panic("need a sheet name at least")
		}

		f._excel.NewSheet(s.Name)
	}
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

func (f *File) setPullDown(s *Sheet) (err error) {
	if s.pd == nil {
		return
	}

	dataSheet := s.pd.sheet(s.Name)
	f.addSheet(dataSheet)
	if err = f.writeData(dataSheet); err != nil {
		return
	}
	if err = f.excel().SetSheetVisible(dataSheet.Name, false); err != nil {
		return
	}

	for index, p := range s.pd.options {
		dvRange := excelize.NewDataValidation(true)
		dvRange.Sqref = p.col + strconv.FormatInt(int64(s.writeRow), 10) + ":" + p.col + "3000"

		var endColunmName string
		endColunmName, err = excelize.ColumnNumberToName(len(p.data))
		ss := fmt.Sprintf("%s!$A$%d:$%s$%d", dataSheet.Name, index+1, endColunmName, index+1)

		dvRange.SetSqrefDropList(ss)
		if err = f.excel().AddDataValidation(s.Name, dvRange); err != nil {
			return
		}
	}

	return
}

func (f *File) writeDefaultFormatSheet(s *Sheet) (err error) {
	if err = f.setColumnsText(s); err != nil {
		return
	}

	if err = f.writeNotice(s); err != nil {
		return
	}

	if err = f.writeHeader(s); err != nil {
		return
	}

	if err = f.writeData(s); err != nil {
		return
	}

	return
}

func (f *File) setColumnsText(s *Sheet) (err error) {
	// 设置表各列数据格式 数字默认为“文本”
	for i := range s.Header {
		var colName string
		if colName, err = excelize.ColumnNumberToName(1 + i); nil != err {
			return
		}

		if err = f.excel().SetColStyle(s.Name, colName, f.StyleNumFmtText()); nil != err {
			return
		}
	}

	return
}

func (f *File) writeNotice(s *Sheet) (err error) {
	// 判断是否有提示并设置
	if s.Notice != "" {
		row := s.writeRowIncr()
		if err = f.excel().SetCellValue(s.Name, row, s.Notice); err != nil {
			return
		}
		if err = f.excel().SetCellStyle(s.Name, row, row, f.StyleRedTextLocked()); nil != err {
			return
		}
	}

	return
}

func (f *File) writeHeader(s *Sheet) (err error) {
	// 判断是否有提示并设置
	if len(s.Header) != 0 {
		row := s.writeRowIncr()
		if err = f.excel().SetSheetRow(s.Name, row, &s.Header); err != nil {
			return
		}
		if err = f.excel().SetRowStyle(s.Name, s.writeRow, s.writeRow, f.StyleLocked()); err != nil {
			return
		}
	}

	return
}

func (f *File) writeData(s *Sheet) (err error) {
	// 判断是否有预置数据并设置
	if len(s.Data) != 0 {
		for _, d := range s.Data {
			var (
				row  = s.writeRowIncr()
				name string
				i    int
			)
			if name, i, err = excelize.SplitCellName(row); err != nil {
				return
			}

			for index, o := range d {
				var number int
				if number, err = excelize.ColumnNameToNumber(name); err != nil {
					return
				}

				var cellName string
				if cellName, err = excelize.CoordinatesToCellName(number+index, i); err != nil {
					return
				}
				if err = f.excel().SetCellValue(s.Name, cellName, o); err != nil {
					return
				}
			}
		}
	}

	return
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

	// 判断是否有预置数据并设置
	if len(s.Data) != 0 {
		for _, d := range s.Data {
			var (
				row  = s.writeRowIncr()
				name string
				i    int
			)
			if name, i, err = excelize.SplitCellName(row); err != nil {
				return
			}

			for index, o := range d {
				var number int
				if number, err = excelize.ColumnNameToNumber(name); err != nil {
					return
				}

				var cellName string
				if cellName, err = excelize.CoordinatesToCellName(number+index, i); err != nil {
					return
				}
				if err = _excel.SetCellValue(s.Name, cellName, o); err != nil {
					return
				}
			}
		}
	}

	return
}

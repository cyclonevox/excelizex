package excelizex

import (
	"github.com/xuri/excelize/v2"
)

type excel struct {
	_excel *excelize.File
}

func New() *excel {
	return &excel{_excel: excelize.NewFile()}
}

func (e *excel) getFile() *excelize.File {
	return e._excel
}

func (e *excel) AddSheets(sheets ...Sheet) *excel {
	var (
		_excel = e.getFile()
		err    error
	)

	for _, s := range sheets {
		if s.Name == "" {
			panic("need a sheet name at least")
		}
		e._excel.NewSheet(s.Name)

		// 设置表各列数据格式 数字默认为“文本”
		for i := range s.Header {
			var colName string
			if colName, err = excelize.ColumnNumberToName(1 + i); nil != err {
				panic(err)
			}

			if err = _excel.SetColStyle(s.Name, colName, e.StyleNumFmtText()); nil != err {
				panic(err)
			}
		}

		// 判断是否有提示并设置
		if s.Notice != "" {
			row := s.writeRowName()
			if err = _excel.SetCellValue(s.Name, row, s.Notice); err != nil {
				panic(err)
			}
			if err = _excel.SetCellStyle(s.Name, row, row, e.StyleRedTextLocked()); nil != err {
				panic(err)
			}
		}

		// 判断是否有提示并设置
		if len(s.Header) != 0 {
			row := s.writeRowName()
			if err = _excel.SetSheetRow(s.Name, row, &s.Header); err != nil {
				panic(err)
			}
			if err = _excel.SetRowStyle(s.Name, s.writeRow, s.writeRow, e.StyleLocked()); err != nil {
				panic(err)
			}
		}
	}

	return e
}

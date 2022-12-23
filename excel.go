package excelizex

import (
	"github.com/xuri/excelize/v2"
)

type excel struct {
	file *excelize.File
}

func New() *excel {
	e := &excel{file: excelize.NewFile()}

	return e
}

func (e *excel) getFile() *excelize.File {
	return e.file
}

func (e *excel) AddSheets(sheets ...Sheet) *excel {
	for _, sheet := range sheets {
		if sheet.Name == "" {
			panic("need a sheet name at least")
		}

		e.file.NewSheet(sheet.Name)
	}

	return e
}

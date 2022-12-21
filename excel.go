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

func (e *excel) AddSheets(sheet ...Sheet) *excel {

	return e
}

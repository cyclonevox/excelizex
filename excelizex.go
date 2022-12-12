package excelizex

import (
	"github.com/xuri/excelize/v2"
)

type excel struct {
	file *excelize.File

	// 全局默认样式(设置全局列表使用数字应用文本)
	publicStyle int
	// 顶栏提示默认样式
	noticeStyle int
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

// func SetPullDown() SheetOption {
// 	return func(s *Sheet) {
//
// 	}
// }

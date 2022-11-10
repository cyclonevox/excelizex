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

	var err error
	if e.noticeStyle, err = e.file.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:   true,
			Family: "微软雅黑",
			Size:   11,
			Color:  "#FF0000",
		},
	}); nil != err {
		panic(err)
	}

	if e.publicStyle, err = e.file.NewStyle(&excelize.Style{NumFmt: 49}); nil != err {
		panic(err)
	}

	return e
}

func (e *excel) getFile() *excelize.File {
	return e.file
}

func (e *excel) AddSheet(base SheetBase, opts ...SheetOption) *excel {
	s := &Sheet{
		name:   base.Name,
		notice: base.Notice,
		header: base.Header,
	}

	for _, opt := range opts {
		opt(s)
	}

	if s.maxRow == 0 {
		s.maxRow = 3000
	}

	if err := s.build(e); err != nil {
		panic(err)
	}

	if s.vSheet != nil {
		e.file.NewSheet(s.vSheetName)
		if err := e.file.SetSheetVisible(s.vSheetName, false); nil != err {
			panic(err)
		}
		// todo Add validation sheet data and validations
	}

	return e
}

// func SetPullDown() SheetOption {
// 	return func(s *Sheet) {
//
// 	}
// }

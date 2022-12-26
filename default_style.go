package excelizex

import "github.com/xuri/excelize/v2"

func (f *file) StyleNumFmtText() int {
	var (
		style int
		err   error
	)
	if style, err = f._excel.NewStyle(&excelize.Style{NumFmt: 49}); nil != err {
		panic(err)
	}

	return style
}

func (f *file) StyleRedTextLocked() int {
	var (
		style int
		err   error
	)
	if style, err = f._excel.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:   true,
			Family: "微软雅黑",
			Size:   11,
			Color:  "#FF0000",
		},
		Protection: &excelize.Protection{Locked: true},
	}); err != nil {
		panic(err)
	}

	return style
}

func (f *file) StyleLocked() int {
	var (
		style int
		err   error
	)
	if style, err = f._excel.NewStyle(&excelize.Style{Protection: &excelize.Protection{Locked: true}}); nil != err {
		panic(err)
	}

	return style
}

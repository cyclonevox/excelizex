package excelizex

import (
	"github.com/cyclonevox/excelizex/style"
)

func (f *File) styleNumFmtText() int {
	var (
		styleId int
		err     error
	)
	if styleId, err = f._excel.NewStyle(style.NumFmtText); nil != err {
		panic(err)
	}

	return styleId
}

func (f *File) defaultNoticeStyleLocked() int {
	var (
		styleId int
		err     error
	)
	if styleId, err = f._excel.NewStyle(style.DefaultNoticeStyle); err != nil {
		panic(err)
	}

	return styleId
}

func (f *File) styleLocked() int {
	var (
		styleId int
		err     error
	)
	if styleId, err = f._excel.NewStyle(style.StyleLocked); nil != err {
		panic(err)
	}

	return styleId
}

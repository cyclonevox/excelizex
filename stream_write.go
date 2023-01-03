package excelizex

import (
	"errors"
	"github.com/xuri/excelize/v2"
	"strconv"
)

type StreamWritable interface {
	Next() bool
	Data() any
	Close() error
}

// StreamWriteIn 通过调用迭代器接口为excel文件来生成表.
// 迭代器接口中的 Data() 返回返回的值的结构体来作为生成表的头.时无需用传入option手动设置表头
// option 可设定表，需要注意的是，必须设定表名称.
func (f *File) StreamWriteIn(i StreamWritable, option ...SheetOption) (err error) {
	var (
		s         *Sheet
		sw        *excelize.StreamWriter
		beginAxis int64
	)

	for j := 0; i.Next(); j++ {
		d := i.Data()

		if j == 0 {
			s = NewSheet(option...)
			// 检测到未包含header设置则会使用 SetHeaderByStruct 获取数据中结构体的header元素
			if len(s.Header) == 0 {
				s = s.SetHeaderByStruct(d)
			}
			// 未包含名称则错误
			if s.Name == "" {
				return errors.New("plz set sheet name")
			}

			f.AddSheets(s)

			if sw, err = f.excel().NewStreamWriter(s.Name); err != nil {
				return
			}

			if s.Notice == "" {
				beginAxis = 2
			} else {
				beginAxis = 3
			}
		}

		if err = sw.SetRow("A"+strconv.FormatInt(beginAxis, 10), singleRowData(d)); err != nil {
			return
		}

		beginAxis++
	}

	if err = i.Close(); err != nil {
		return
	}

	if err = sw.Flush(); err != nil {
		return
	}

	// 最后将头和notice等文件设置
	if err = f.setDefaultFormatSheetAndStyle(s); err != nil {
		return
	}

	return
}

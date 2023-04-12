package excelizex

import (
	"github.com/xuri/excelize/v2"
)

type StreamWritable interface {
	Next() bool
	DataRow() []any
	Close() error
}

// AddSheetByStream 通过调用迭代器接口为excel文件来生成表.
// 迭代器接口中的 RawData() 返回返回的值的结构体来作为生成表的头.时无需用传入option手动设置表头
// Option 可设定表，需要注意的是，必须设定表名称.
func (f *File) AddSheetByStream(i StreamWritable, sheet *sheet) (err error) {
	var sw *excelize.StreamWriter

	f.addSheet(sheet)
	if sw, err = f.excel().NewStreamWriter(sheet.name); err != nil {
		return
	}

	for j := 0; i.Next(); j++ {
		d := i.DataRow()

		if err = sw.SetRow(sheet.nextWriteRow(), getRowData(d)); err != nil {
			return
		}
	}

	if err = i.Close(); err != nil {
		return
	}

	if err = sw.Flush(); err != nil {
		return
	}

	return
}

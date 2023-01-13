package excelizex

import (
	"github.com/xuri/excelize/v2"
	"strconv"
)

const noStyle = -1

type StreamWritable interface {
	Next() bool
	DataRow() []excelize.Cell
	Close() error
}

// AddSheetByStream 通过调用迭代器接口为excel文件来生成表.
// 迭代器接口中的 Data() 返回返回的值的结构体来作为生成表的头.时无需用传入option手动设置表头
// Option 可设定表，需要注意的是，必须设定表名称.
func (f *File) AddSheetByStream(i StreamWritable, sheet *Sheet) (err error) {
	var sw *excelize.StreamWriter

	f.addSheet(sheet)
	if err = f.setPullDown(sheet); err != nil {
		return
	}
	if sw, err = f.excel().NewStreamWriter(sheet.Name); err != nil {
		return
	}

	for j := 0; i.Next(); j++ {
		d, fn := i.DataRow()
		if fn != nil {
			if err = fn(f.excel()); err != nil {
				return
			}
		}

		singleRowData(d)

		if err = sw.SetRow(sheet.nextWriteRow(), excelValues(d)); err != nil {
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

func (f *File) WriteInByStream(i StreamWritable, startLine int) (err error) {
	if f.selectSheetName == "" {
		panic("plz use *File.SelectSheet select a sheet by name first")
	}

	var sw *excelize.StreamWriter
	if sw, err = f.excel().NewStreamWriter(f.selectSheetName); err != nil {
		return
	}

	for j := 0; i.Next(); j++ {
		d, fn := i.DataRow()
		if fn != nil {
			if err = fn(f.excel()); err != nil {
				return
			}
		}

		if err = sw.SetRow("A"+strconv.FormatInt(int64(startLine), 10), singleRowData(d)); err != nil {
			return
		}

		startLine++
	}

	if err = i.Close(); err != nil {
		return
	}
	if err = sw.Flush(); err != nil {
		return
	}

	return
}

func excelValues(list []excelize.Cell) (value []any) {
	value = make([]any, len(list), len(list))

	for index, l := range list {
		value[index] = l
	}

	return
}

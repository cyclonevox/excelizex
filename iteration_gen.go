package excelizex

import (
	"errors"
	"github.com/xuri/excelize/v2"
	"strconv"
)

type Iteration interface {
	Next() bool
	Data() any
	Close() error
}

// StreamImport 通过调用迭代器接口，并且使用
func (f *file) StreamImport(i Iteration, option ...SheetOption) (err error) {
	var (
		sw        *excelize.StreamWriter
		beginAxis int64
	)

	for j := 0; i.Next(); j++ {
		result := i.Data()

		if j == 0 {
			s := genSheet(result)
			for _, o := range option {
				o(&s)
			}
			f.AddSheets(s)

			if s.Name == "" {
				return errors.New("please set sheet name")
			}

			//设置默认格式
			if err = f.setDefaultFormatSheetAndStyle(s); err != nil {
				return
			}

			if sw, err = f.excel().NewStreamWriter(s.Name); err != nil {
				return
			}

			if s.Notice != "" {
				beginAxis = 2
			} else {
				beginAxis = 3
			}
		}

		if err = sw.SetRow("A"+strconv.FormatInt(beginAxis, 10), genSingleData(result)); err != nil {
			return
		}

		beginAxis++
	}

	i.Close()

	return sw.Flush()
}

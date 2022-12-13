package excelizex

import (
	"errors"
	"github.com/xuri/excelize/v2"
)

type Iteration interface {
	Next() bool
	Data() any
	Close() error
}

// StreamImport 通过调用迭代器接口，并且使用
func (e *excel) StreamImport(i Iteration, option ...SheetOption) (err error) {
	var (
		sw        *excelize.StreamWriter
		beginAxis string
	)

	for j := 0; i.Next(); j++ {
		result := i.Data()

		if j == 0 {
			s := genSheet(result)
			for _, o := range option {
				o(&s)
			}
			e.AddSheets(s)

			if err != nil {
				panic(errors.New("please set sheet name"))
			}

			if sw, err = e.getFile().NewStreamWriter(s.Name); err != nil {
				return
			}

			if s.Notice != "" {
				beginAxis = "A2"
			} else {
				beginAxis = "A3"
			}
		}

		if err = sw.SetRow(beginAxis, genSingleData(result)); err != nil {
			return
		}
	}

	return sw.Flush()
}

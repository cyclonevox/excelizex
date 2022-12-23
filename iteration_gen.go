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
func (e *excel) StreamImport(i Iteration, option ...SheetOption) (err error) {
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
			e.AddSheets(s)

			if s.Name == "" {
				panic(errors.New("please set sheet name"))
			}

			if sw, err = e.getFile().NewStreamWriter(s.Name); err != nil {
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

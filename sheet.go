package excelizex

import "strconv"

type Sheet struct {
	// 工作表名
	name string
	// 提醒
	notice string
	// 表头
	header []string
	// 预置数据
	data [][]string
	// 最大行数
	maxRow int
	//
	vSheetName string

	vSheet []string
}

func (s *Sheet) build(e *excel) (err error) {
	file := e.getFile()
	file.NewSheet(s.name)

	if s.notice == "" {
		var rows [][]string
		if rows, err = file.GetRows(s.name); nil != err {
			return
		}
		axis := "A" + strconv.FormatInt(int64(len(rows))+1, 10)

		// 设置提示
		if err = file.SetCellValue(s.name, axis, s.notice); err != nil {
			return
		}
		if err = file.SetCellStyle(s.name, axis, axis, e.noticeStyle); nil != err {
			return
		}
	}

	if len(s.header) == 0 {
		var rows [][]string
		if rows, err = file.GetRows(s.name); nil != err {
			return
		}
		axis := "A" + strconv.FormatInt(int64(len(rows))+1, 10)
		if err = file.SetSheetRow(s.name, axis, &s.header); err != nil {
			return
		}
		if err = file.SetRowStyle(s.name, len(rows)+1, s.maxRow, e.publicStyle); nil != err {
			return
		}
	}

	if len(s.data) == 0 {
		for _, data := range s.data {
			var rows [][]string
			if rows, err = file.GetRows(s.name); nil != err {
				return
			}

			axis := "A" + strconv.FormatInt(int64(len(rows))+1, 10)
			if err = file.SetSheetRow(s.name, axis, &data); err != nil {
				return
			}
		}
	}

	// todo add sheet extra info

	return
}

type SheetOption = func(*Sheet)

type SheetBase struct {
	// 工作表名
	Name string
	// 提醒
	Notice string
	// 表头
	Header []string
}

func SetData(data [][]string) SheetOption {
	return func(s *Sheet) {
		s.data = data
	}
}

func SetMaxRow(maxRow int) SheetOption {
	return func(s *Sheet) {
		s.maxRow = maxRow
	}
}

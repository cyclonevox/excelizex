package excelizex

import (
	"bytes"
	"fmt"
	"github.com/cyclonevox/excelizex/extra"
	"github.com/cyclonevox/excelizex/style"
	"github.com/xuri/excelize/v2"
	"io"
	"strconv"
	"strings"
	"unicode/utf8"
)

const OptionsSaveTable = "选项数据表"

type File struct {
	selectSheetName string
	// k:sheetName v:sheetName
	sheets     map[string]*sheet
	_excel     *excelize.File
	styleCache map[string]*style.Payload
}

func New(reader ...io.Reader) *File {
	var (
		f   *File
		err error
	)

	if len(reader) > 0 {
		if f, err = newExcelFormIo(reader[0]); err != nil {
			panic(err)
		} else {
			return f
		}
	}

	if f == nil {
		f = &File{_excel: excelize.NewFile()}
		if f.styleCache == nil {
			f.styleCache = make(map[string]*style.Payload)
			f.AddStyles(
				style.DefaultNoticeStyle,
				style.DefaultDataStyle,
				style.DefaultHeaderStyle,
				style.DefaultHeaderRedStyle,
				style.DefaultLocked,
				style.DefaultNumFmtText,
				style.DefaultRedFont,
			)
		}
	}

	return f
}

func (f *File) AddStyles(Styles ...style.Style) *File {
	for _, _style := range Styles {
		f.addStyle(_style)
	}

	return f
}

func (f *File) addStyle(_style style.Style) {
	styleId, err := f.excel().NewStyle(_style.Style())
	if err != nil {
		panic(err)
	}

	f.styleCache[_style.Name()] = &style.Payload{
		StyleID: styleId,
		Style:   _style,
	}

	return
}

func (f *File) genNewStyle(styleNames []string) (styleID int) {
	_style, ok := f.styleCache[styleNames[0]]
	if ok {
		for _, name := range styleNames {
			_newStyle, ok := f.styleCache[name]
			if ok {
				_style.Append(_newStyle.Style)
			}
		}

		styleId, err := f.excel().NewStyle(_style.Style.Style())
		if err != nil {
			panic(err)
		}

		f.styleCache[_style.Name()] = &style.Payload{
			StyleID: styleId,
			Style:   _style.Style,
		}
	}

	return styleID
}

func (f *File) AddSheet(name string, model any) {
	var err error

	s := NewSheet(name, model)
	f.addSheet(s)

	if err = f.writeDefaultFormatSheet(s); err != nil {
		panic(err)
	}
}

func (f *File) AddFormattedSheets(sheets ...*sheet) *File {
	var err error

	for _, s := range sheets {
		f.addSheet(s)

		if err = f.writeDefaultFormatSheet(s); err != nil {
			panic(err)
		}
	}

	return f
}

func (f *File) addSheet(sheets ...*sheet) {
	for _, s := range sheets {
		if s.name == "" || s.name == "Sheet1" {
			panic("need a sheet name at least")
		}

		f._excel.NewSheet(s.name)
	}
}

func (f *File) Unlock(password string) (file *File, err error) {
	for _, n := range f.excel().GetSheetList() {
		if err = f.excel().UnprotectSheet(n, password); nil != err {
			return f, err
		}
	}

	return f, nil
}

func (f *File) SelectSheet(sheetName string) *File {
	if _, ok := f.sheets[sheetName]; ok {
		f.selectSheetName = sheetName
	} else {
		panic(fmt.Sprintf("sheetName: %s does not exist !", sheetName))
	}

	return f
}

func (f *File) excel() *excelize.File {
	return f._excel
}

func (f *File) SaveAs(name string, password ...string) (err error) {
	f.excel().DeleteSheet("Sheet1")

	if len(password) > 0 {
		protect := &excelize.FormatSheetProtection{
			Password:          password[0],
			EditObjects:       true,
			EditScenarios:     true,
			SelectLockedCells: true,
		}

		for _, n := range f.excel().GetSheetList() {
			if err = f.excel().ProtectSheet(n, protect); nil != err {
				return
			}
		}
	}

	return f.excel().SaveAs(name)
}

func (f *File) Buffer(password ...string) (*bytes.Buffer, error) {
	f.excel().DeleteSheet("Sheet1")

	if len(password) > 0 {
		protect := &excelize.FormatSheetProtection{
			Password:          password[0],
			EditObjects:       true,
			EditScenarios:     true,
			SelectLockedCells: true,
		}

		for _, n := range f.excel().GetSheetList() {
			if err := f.excel().ProtectSheet(n, protect); nil != err {
				return nil, err
			}
		}
	}

	return f.excel().WriteToBuffer()
}

func (f *File) setPullDown(s *sheet) (err error) {
	if s.pd == nil {
		return
	}

	dataSheet := s.pd.sheet(s.name)
	f.addSheet(dataSheet)
	if err = f.writeData(dataSheet); err != nil {
		return
	}
	if err = f.excel().SetSheetVisible(dataSheet.name, false); err != nil {
		return
	}

	for index, p := range s.pd.options {
		dvRange := excelize.NewDataValidation(true)
		dvRange.Sqref = p.col + strconv.FormatInt(int64(s.writeRow+1), 10) + ":" + p.col + "3000"

		var endColunmName string
		endColunmName, err = excelize.ColumnNumberToName(len(p.data))
		ss := fmt.Sprintf("%s!$A$%d:$%s$%d", dataSheet.name, index+1, endColunmName, index+1)

		dvRange.SetSqrefDropList(ss)
		if err = f.excel().AddDataValidation(s.name, dvRange); err != nil {
			return
		}
	}

	return
}

func (f *File) writeDefaultFormatSheet(s *sheet) (err error) {
	// 遗憾的是我必须先将numFmt 和 未锁定 style给每一列设置好
	if err = f.settingAllCol(s); err != nil {
		return
	}

	if err = f.writeNotice(s); err != nil {
		return
	}

	if err = f.writeHeader(s); err != nil {
		return
	}

	if err = f.setPullDown(s); err != nil {
		return
	}

	if err = f.writeData(s); err != nil {
		return
	}

	return
}

func (f *File) settingAllCol(s *sheet) (err error) {
	var colName string
	// 设置表各列数据格式 数字默认为“文本,未锁定”
	for i := range s.header {
		if colName, err = excelize.ColumnNumberToName(1 + i); nil != err {
			return
		}
		if err = f.excel().SetColStyle(s.name, colName, f.styleCache["default-all"].StyleID); nil != err {
			return
		}
	}

	return
}

func (f *File) writeNotice(s *sheet) (err error) {
	// 判断是否有提示并设置
	// 根据换行设置单元格
	if s.notice != "" {
		row := s.nextWriteRow()
		if err = f.excel().SetCellValue(s.name, row, s.notice); err != nil {
			return
		}
		if err = f.setPartStyle(s, extra.NoticePart); err != nil {
			return
		}
	}

	return
}

func (f *File) writeHeader(s *sheet) (err error) {
	row := s.nextWriteRow()
	if err = f.excel().SetSheetRow(s.name, row, &s.header); err != nil {
		return
	}
	if err = f.setPartStyle(s, extra.HeaderPart); err != nil {
		return
	}

	return
}

func (f *File) setPartStyle(s *sheet, part extra.Part) (err error) {
	if _style, ok := s.styleRef[string(part)]; ok {
		if _style[0].AutoWide {
			switch part {
			case extra.NoticePart:
				if err = f.noticeAdaptionWidth(s); err != nil {
					return
				}
			case extra.HeaderPart:
				if err = f.headerAdaptionWidth(s); err != nil {
					return
				}
			}
		}

		var styleId int
		for _, _s := range _style {
			// 当该样式为组合样式时。则需要重新生成新样式
			if len(_s.StyleNames) > 1 {
				styleId = f.genNewStyle(_s.StyleNames)
			} else if len(_s.StyleNames) == 1 {
				styleId = f.styleCache[_s.StyleNames[0]].StyleID
			}

			// 设置该样式
			if err = f.excel().SetCellStyle(
				s.name,
				_s.Cell.StartCell.Format(),
				_s.Cell.EndCell.Format(),
				styleId,
			); nil != err {
				return
			}
		}
	}

	return
}

func (f *File) headerAdaptionWidth(s *sheet) (err error) {
	var colName string
	for i := range s.header {
		if colName, err = excelize.ColumnNumberToName(1 + i); nil != err {
			return
		}

		if err = f.excel().SetColWidth(s.name, colName, colName, float64(9*utf8.RuneCount([]byte(s.header[i]))/4+1)); err != nil {
			return
		}
	}

	return
}

func (f *File) noticeAdaptionWidth(s *sheet) (err error) {
	var (
		columnNumber string
		max          int
		lines        = strings.Split(s.notice, "\n")
	)
	for _, line := range lines {
		if max < utf8.RuneCount([]byte(line)) {
			max = utf8.RuneCount([]byte(line))
		}
	}
	max = max/4 + 1

	if columnNumber, err = excelize.ColumnNumberToName(max); err != nil {
		return
	}
	if err = f.excel().MergeCell(s.name, s.getWriteRow(), columnNumber+"1"); err != nil {
		return
	}
	if err = f.excel().SetRowHeight(s.name, 1, float64(17*(len(lines)))); err != nil {
		return
	}

	return
}

func (f *File) writeData(s *sheet) (err error) {
	// 判断是否有预置数据并设置
	if len(s.data) != 0 {
		for _, d := range s.data {
			var (
				row  = s.nextWriteRow()
				name string
				i    int
			)
			if name, i, err = excelize.SplitCellName(row); err != nil {
				return
			}

			for index, o := range d {
				var number int
				if number, err = excelize.ColumnNameToNumber(name); err != nil {
					return
				}

				var cellName string
				if cellName, err = excelize.CoordinatesToCellName(number+index, i); err != nil {
					return
				}
				if err = f.excel().SetCellValue(s.name, cellName, o); err != nil {
					return
				}
			}
		}
	}

	return
}

func (f *File) ExistSheet(sheetName string) (exist bool) {
	if _, ok := f.sheets[sheetName]; ok {
		return ok
	} else {
		return false
	}
}

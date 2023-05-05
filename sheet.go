package excelizex

import (
	"reflect"
	"regexp"
	"strconv"

	"github.com/xuri/excelize/v2"
)

type Sheet struct {
	// 表名
	name string
	// 格式化后的数据
	// 顶栏提示
	notice string
	// 表头
	header []string
	// 数据
	data *[][]any

	// 下拉选项 暂时只支持单列
	pd *pullDown
	// 列和style分配语句的映射存储
	// v:Col -> k:style raw string list;col 为 -1时为notice/
	styleRef map[int][]string
	// 写入到第几行,主要用于标记生成excel中的表时，需要续写的位置
	writeRow int
}

func NewSheet(sheetName string, a any) *Sheet {
	if sheetName == "" {
		panic("Sheet name cannot be empty")
	}

	s := newMetas(a).sheet(sheetName)

	return s
}

func (s *Sheet) Excel() *File {
	if s.name == "" {
		panic("need a Sheet name at least")
	}

	return New().AddFormattedSheets(s)
}

func getRowData(row any) (list []any) {
	typ := reflect.TypeOf(row)
	val := reflect.ValueOf(row)

	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
		val = val.Elem()
	}

	switch typ.Kind() {
	case reflect.Struct:
		for j := 0; j < typ.NumField(); j++ {
			field := typ.Field(j)

			hasTag := field.Tag.Get("excel")
			if hasTag != "" && hasTag != "notice" {
				list = append(list, val.Field(j).Interface())
			}
		}
	case reflect.Slice:
		for i := 0; i < val.Len(); i++ {
			list = append(list, val.Index(i).Interface())
		}

	default:
		panic("support struct only")
	}

	return
}

// findHeaderColumnName 寻找表头名称或者是列名称
func (s *Sheet) findHeaderColumnName(headOrColName string) (columnName string, err error) {
	for i, h := range s.header {
		if h == headOrColName {
			columnName, err = excelize.ColumnNumberToName(i + 1)

			return
		}
	}

	regular := `[A-Z]+`
	reg := regexp.MustCompile(regular)
	if !reg.MatchString(headOrColName) {
		panic("plz use A-Z ColName or HeaderName for option name ")
	}

	columnName = headOrColName

	return
}

// SetOptions 设置下拉的选项
func (s *Sheet) SetOptions(headOrColName string, options any) *Sheet {
	name, err := s.findHeaderColumnName(headOrColName)
	if err != nil {
		panic(err)
	}

	pd := newPullDown().addOptions(name, options)

	if s.pd == nil {
		s.pd = pd
	} else {
		s.pd.merge(pd)
	}

	return s
}

// nextWriteRow 会获取目前该写入的行
// 每次调用该方法表示行数增长 返回 A1 A2... 等名称
func (s *Sheet) nextWriteRow(num ...int) string {
	if len(num) > 0 {
		s.writeRow += num[0]
	} else {
		s.writeRow++
	}

	return "A" + strconv.FormatInt(int64(s.writeRow), 10)
}

func (s *Sheet) getWriteRow() string {
	return "A" + strconv.FormatInt(int64(s.writeRow), 10)
}

func (s *Sheet) resetWriteRow() string {
	s.writeRow = 1

	return "A" + strconv.FormatInt(int64(s.writeRow), 10)
}

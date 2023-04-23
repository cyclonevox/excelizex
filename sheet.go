package excelizex

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/cyclonevox/excelizex/style"
	"github.com/xuri/excelize/v2"
)

type sheet struct {
	// 表名
	name string
	// 格式化后的数据
	// 顶栏提示
	notice string
	// 表头
	header []string
	// 数据
	data [][]any
	// 下拉选项 暂时只支持单列
	pd *pullDown

	// 分布和style分配语句的配置
	// v:part -> k:style string
	styleRef map[string][]style.Parsed
	// 写入到第几行,主要用于标记生成excel中的表时，需要续写的位置
	writeRow int
}

func NewSheet(sheetName string, a any) *sheet {
	if sheetName == "" {
		panic("sheet cannot be empty")
	}

	s := &sheet{
		name:     sheetName,
		styleRef: make(map[string][]style.Parsed),
		writeRow: 0,
	}
	if a != nil {
		s.initSheetData(a)
	}

	return s
}

func (s *sheet) Excel() *File {
	if s.name == "" {
		panic("need a sheet name at least")
	}

	return New().AddFormattedSheets(s)
}

func (s *sheet) initSheetData(a any) {
	typ := reflect.TypeOf(a)
	val := reflect.ValueOf(a)

	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	// 如果作为Slice类型的传入对象，则还需要注意拆分后进行处理
	switch typ.Kind() {
	case reflect.Slice:
		for i := 0; i < val.Len(); i++ {
			if i == 0 {
				s.setHeaderByStruct(val.Index(i).Interface())
			}

			s.data = append(s.data, getRowData(val.Index(i).Interface()))
		}
	case reflect.Struct:
		s.setHeaderByStruct(a)
	}

	return
}

// SetHeaderByStruct 方法会检测结构体中的excel标签，以获取结构体表头
func (s *sheet) setHeaderByStruct(a any) *sheet {
	typ := reflect.TypeOf(a)
	val := reflect.ValueOf(a)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
		val = val.Elem()
	}

	if typ.Kind() != reflect.Struct {
		panic(errors.New("generate function support using struct only"))
	}

	for i := 0; i < typ.NumField(); i++ {
		typeField := typ.Field(i)

		partTag := typeField.Tag.Get("excel")
		if partTag == "" {
			continue
		} else {

			// 判断是excel tag 是指向哪个部分
			params := strings.Split(partTag, "|")
			if len(params) > 0 {
				switch part(params[0]) {
				case noticePart:
					s.notice = val.Field(i).String()

					// 添加提示样式映射
					styleString := typeField.Tag.Get("style")
					if styleString == "" {
						continue
					}
					_noticeStyle := style.TagParse(styleString).Parse()
					_noticeStyle.Cell.StartCell = style.Cell{Col: "A", Row: 1}
					_noticeStyle.Cell.EndCell = style.Cell{Col: "A", Row: 1}
					s.styleRef[fmt.Sprintf("%s", noticePart)] = []style.Parsed{_noticeStyle}

				case headerPart:
					// todo： 现在header的style暂时不能交叉设置，原因是会被覆盖，需要在后续改动
					s.header = append(s.header, params[1])
					styleString := typeField.Tag.Get("style")
					if styleString == "" {
						continue
					}

					colName, err := excelize.ColumnNumberToName(len(s.header))
					if err != nil {
						panic(err)
					}
					headerStyle := style.TagParse(styleString).Parse()

					// todo: 待优化
					var sp []style.Parsed

					var okk bool
					if pp, ok := s.styleRef[fmt.Sprintf("%s", headerPart)]; ok {
						for _, p := range pp {
							if reflect.DeepEqual(p.StyleNames, headerStyle.StyleNames) {
								p.Cell.EndCell = style.Cell{Col: colName, Row: 2}
								okk = true
							}
							sp = append(sp, p)
						}

						if !okk {
							headerStyle.Cell.StartCell = style.Cell{Col: colName, Row: 2}
							headerStyle.Cell.EndCell = style.Cell{Col: colName, Row: 2}

							sp = append(sp, headerStyle)
						}
					} else {
						headerStyle.Cell.StartCell = style.Cell{Col: colName, Row: 2}
						headerStyle.Cell.EndCell = style.Cell{Col: colName, Row: 2}

						sp = append(sp, headerStyle)
					}

					s.styleRef[fmt.Sprintf("%s", headerPart)] = sp

					styleString = typeField.Tag.Get("data-style")
					// todo :暂不支持 太累了抱歉
					//dataStyle := style.TagParse(styleString).Parse(extra.dataPart)
					//s.styleRef[fmt.Sprintf("%s-%s", extra.dataPart, params[1])] = dataStyle
				}
			}

		}
	}

	return s
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
func (s *sheet) findHeaderColumnName(headOrColName string) (columnName string, err error) {
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
func (s *sheet) SetOptions(headOrColName string, options any) *sheet {
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
func (s *sheet) nextWriteRow(num ...int) string {
	if len(num) > 0 {
		s.writeRow += num[0]
	} else {
		s.writeRow++
	}

	return "A" + strconv.FormatInt(int64(s.writeRow), 10)
}

func (s *sheet) getWriteRow() string {
	return "A" + strconv.FormatInt(int64(s.writeRow), 10)
}

func (s *sheet) resetWriteRow() string {
	s.writeRow = 1

	return "A" + strconv.FormatInt(int64(s.writeRow), 10)
}

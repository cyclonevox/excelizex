package excelizex

import (
	"errors"
	"fmt"
	"github.com/cyclonevox/excelizex/extra"
	"github.com/cyclonevox/excelizex/style"
	"github.com/xuri/excelize/v2"
	"reflect"
	"regexp"
	"strconv"
	"strings"
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

func NewSheet(sheetName string) *sheet {

	s := &sheet{
		name:     sheetName,
		styleRef: make(map[string][]style.Parsed),
		writeRow: 0,
	}

	return s
}

func (s *sheet) Excel() *File {
	if s.name == "" {
		panic("need a sheet name at least")
	}

	return New().AddSheets(s)
}

func (s *sheet) initSheet(a any) *sheet {
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

	return s
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
				switch extra.Part(params[0]) {
				case extra.NoticePart:
					s.notice = val.Field(i).String()

					// 添加提示样式映射
					styleString := typeField.Tag.Get("style")
					_noticeStyle := style.TagParse(styleString).Parse()
					_noticeStyle[0].Cell.StartCell = style.Cell{Col: "A", Row: 1}
					_noticeStyle[0].Cell.EndCell = style.Cell{Col: "A", Row: 1}
					s.styleRef[fmt.Sprintf("%s", extra.NoticePart)] = _noticeStyle

				case extra.HeaderPart:
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
					for _, _style := range headerStyle {
						if pp, ok := s.styleRef[fmt.Sprintf("%s", extra.HeaderPart)]; ok {
							for _, p := range pp {
								if reflect.DeepEqual(p.StyleNames, _style.StyleNames) {
									p.Cell.EndCell = style.Cell{Col: colName, Row: 1}
								}
							}
						} else {
							_style.Cell.StartCell = style.Cell{Col: colName, Row: 1}
							_style.Cell.EndCell = style.Cell{Col: colName, Row: 1}
						}

						sp = append(sp, _style)
					}

					s.styleRef[fmt.Sprintf("%s", extra.HeaderPart)] = sp

					styleString = typeField.Tag.Get("data-style")
					// todo :暂不支持 太累了抱歉
					//dataStyle := style.TagParse(styleString).Parse(extra.DataPart)
					//s.styleRef[fmt.Sprintf("%s-%s", extra.DataPart, params[1])] = dataStyle
				}
			}

		}
	}

	return s
}

func getRowData(row any) (list []any) {
	typ := reflect.TypeOf(row)
	val := reflect.ValueOf(row)

	if typ.Kind() == reflect.Struct {
		for j := 0; j < typ.NumField(); j++ {
			field := typ.Field(j)

			hasTag := field.Tag.Get("excel")
			if hasTag != "" {
				list = append(list, val.Field(j).Interface())
			}
		}
	} else {
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

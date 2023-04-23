package excelizex

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/cyclonevox/excelizex/style"
	"github.com/xuri/excelize/v2"
)

const (
	headerPart part = "header"
	noticePart part = "notice"
	dataPart   part = "data"
)

type part string

type field string

type Meta interface {
	Part() part
	ColIndex() int
	CellValue() string
	ValidateTag() string
	StyleTag() string
}

// 用于存储表元信息，包含
type metaRaw struct {
	// 部分
	Part part
	// 样式 用于存储样式Tag内容
	StyleTag string
	// 验证字段 用于存储validate tag中的数据
	validateTag string
	// 列索引
	ColIndex int
	// Notice的值/表头的值
	CellValue string
}

func (m *meta)

type meta struct {

}

type metaRefs []*meta

func (m metaRefs) metaList() []*meta {
	return m
}

func (m metaRefs) binarySearch(index int) *meta {
	l := m.metaList()
	i, low, high := 0, 0, len(l)-1
	if high == 0 {
		return l[high]
	}

	//循环的终止条件
	for low <= high {
		i++
		// 初始化枢轴
		mid := (low + high) / 2

		if l[mid].ColIndex > index {
			high = mid - 1
		} else if l[mid].ColIndex < index {
			low = mid + 1
		} else {
			return l[mid]
		}
	}

	panic("cannot find index data")
}

type metaCache struct {
	typ   reflect.Type
	val   reflect.Value
	data  []*meta
	cache map[field]metaRefs
}

func (m *metaCache) findMetaByIndex(index int) *meta {
	return m.data[index]
}

func (m *metaCache) findMetaByFieldName(field field, index int) *meta {
	return m.cache[field].binarySearch(index)
}

func (m *metaCache) setMeta(index int, f field, mt *meta) {
	m.data[index] = mt
	m.cache[f] = append(m.cache[f], mt)
}

func newMetaCache(a any) *metaCache {
	typ := reflect.TypeOf(a)
	val := reflect.ValueOf(a)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
		val = val.Elem()
	}

	if typ.Kind() != reflect.Struct {
		panic(errors.New("generate function support using struct only"))
	}

	numField := typ.NumField()
	m := &metaCache{
		typ:   typ,
		val:   val,
		data:  make([]*meta, numField, numField),
		cache: make(map[field]metaRefs, numField),
	}

	for i := 0; i < numField; i++ {
		m.newMetaParse(typ, val, i)
	}

	return m
}

func (m *metaCache) newMetaParse(typ reflect.Type, val reflect.Value, i int) {
	tf := typ.Field(i)
	tv := val.Field(i)

	var mt *meta
	// 1.检查该字段是否实现了动态扩展头接口。
	if extHeader, ok := tv.Interface().(ExtHeader); ok {
		extHeader.HeaderName()
		extHeader.ValidateTag()
		extHeader.StyleTag()
		extHeader.Data()

		mt = &meta{
			Part:     headerPart,
			ColIndex: i,
		}

		m.setMeta(i, field(tf.Name), mt)
	}

	// 2.是否包含了excel标签
	t := tf.Tag.Get("excel")
	if t == "" {
		m = new(meta)

		params := strings.Split(t, "|")
		switch part(params[0]) {
		case noticePart:
			m.Part = noticePart
			m.CellValue = fieldVal.String()

			// 添加提示样式映射
			styleString := f.Tag.Get("style")
			if styleString == "" {
				return
			}
			_noticeStyle := style.TagParse(styleString).Parse()
			_noticeStyle.Cell = style.Cell{Col: "A", Row: 1}
			_noticeStyle.Cell = style.Cell{Col: "A", Row: 1}
			s.styleRef[fmt.Sprintf("%s", noticePart)] = []style.Parsed{_noticeStyle}
		case headerPart:

		}
	}

	return
}

// 解析tag 将notice解析到 cache 中便于查询
func (m *metaCache) parseNotice(tag string) {

}

// 解析header 将header解析到 cache 中
func (m *metaCache) parseHeader(tag string) {
	// todo： 现在header的style暂时不能交叉设置，原因是会被覆盖，需要在后续改动
	// todo： 现在header的style暂时不能交叉设置，原因是会被覆盖，需要在后续改动
	m.header = append(s.header, params[1])
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

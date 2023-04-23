package excelizex

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/cyclonevox/excelizex/style"
)

const (
	headerPart part = "header"
	noticePart part = "notice"
	dataPart   part = "data"
)

type part string

type field string

type meta struct {
	style.Style

	Part      part
	ColIndex  int
	CellValue string
	ConvFunc  *ConvertFunc
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
	// 1.检查该字段是否实现了动态扩展头方法。
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

func parseHeader(s string) {

}

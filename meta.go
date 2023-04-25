package excelizex

import (
	"errors"
	"reflect"
	"strings"
)

const (
	headerPart Part = "header"
	noticePart Part = "notice"
	dataPart   Part = "data"
)

type Part string

type Meta interface {
	Part() Part
	ColIndex() int
	CellValue() string
	ValidateTag() string
	StyleTag() string
}

// MetaRaw 用于存储表元信息，包含各种格式的元信息
type metaRaw struct {
	// 部分
	part Part
	// 样式 用于存储样式Tag内容
	styleTag string
	// 验证字段 用于存储validate tag中的数据
	validateTag string
	// 列索引
	colIndex int
	// Notice的值/表头的值
	cellValue string
}

type metaRefs []*metaRaw

func (m metaRefs) metaList() []*metaRaw {
	return m
}

func (m metaRefs) binarySearch(index int) *metaRaw {
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

		if l[mid].colIndex > index {
			high = mid - 1
		} else if l[mid].colIndex < index {
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
	data  []*metaRaw
	cache map[string]metaRefs
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
		data:  make([]*metaRaw, 0, numField),
		cache: make(map[string]metaRefs, numField),
	}

	m.newMetaParse(a)

	return m
}

func (m *metaCache) findMetaByIndex(index int) *metaRaw {
	return m.data[index]
}

func (m *metaCache) findMetaByFieldName(field string, index int) *metaRaw {
	return m.cache[field].binarySearch(index)
}

func (m *metaCache) setMeta(field string, mt *metaRaw) {
	m.data = append(m.data, mt)
	m.cache[field] = append(m.cache[field], mt)
}

func (m *metaCache) newMetaParse(a any) {
	typ := reflect.TypeOf(a)
	val := reflect.ValueOf(a)

	for i := 0; i < typ.NumField(); i++ {
		tf := typ.Field(i)
		tv := val.Field(i)

		var mr *metaRaw
		// 1.检查该字段是否实现了动态扩展头接口。
		if extHeader, ok := tv.Interface().(Meta); ok {
			mr = &metaRaw{
				part:        extHeader.Part(),
				styleTag:    extHeader.StyleTag(),
				validateTag: extHeader.ValidateTag(),
				colIndex:    i,
				cellValue:   extHeader.CellValue(),
			}
		}

		// 2.是否包含了excel标签
		t := tf.Tag.Get("excel")
		if t == "" {
			params := strings.Split(t, "|")
			styleTag := tf.Tag.Get("styleTag")
			validateTag := tf.Tag.Get("validateTag")
			switch Part(params[0]) {
			case noticePart:
				// 原始数据存入metaCache
				mr = m.parseTagOrExt(noticePart, styleTag, validateTag, tv.String())
			case headerPart:
				// 原始数据存入metaCache
				mr = m.parseTagOrExt(headerPart, styleTag, validateTag, params[1])
			}
		}

		m.setMeta(tf.Name, mr)
	}

	return
}

// 解析tag或者是扩展字段解析到 cache 中便于查询
func (m *metaCache) parseTagOrExt(
	part Part,
	style string, validate string, cellValue string,
) (mt *metaRaw) {
	i := len(m.data)

	mt = &metaRaw{
		part:        part,
		styleTag:    style,
		validateTag: validate,
		colIndex:    i + 1,
		cellValue:   cellValue,
	}

	return
}

package excelizex

import (
	"reflect"
)

const (
	headerPart Part = "header"
	noticePart Part = "notice"
	dataPart   Part = "data"
)

type Part string

type HeaderMeta interface {
	ExtData() any
	ExtHeader() string
	ExtValidateTag() string
	ExtStyleTag() string
	ExtConvertor() string
}

// MetaRaw 用于存储表元原始信息
type metaRaw struct {
	// 部分
	part Part
	// 样式 用于存储样式Tag内容
	styleTag string
	// 验证字段 用于存储validate tag中的数据
	validateTag string
	// 转换器字段 用于存储convert tag中的数据
	convertTag string
	// 列索引
	colIndex int
	// Notice的值/表头的值/cell的值
	cellValue string
	// fieldNames
	fieldNames []string
}

type metaRaws struct {
	cursor int
	parsed bool
	raws   []*metaRaw

	hasData bool
	data    [][]any
}

func newMetas(a any) *metaRaws {
	r := &metaRaws{
		cursor: 1,
		raws:   make([]*metaRaw, 0),
	}

	val := reflect.ValueOf(a)
	// Slice字段则进入下层循环，可直接导入数据
	if val.Kind() == reflect.Slice {
		r.data = make([][]any, val.Len())
		for j := 0; j < val.Len(); j++ {
			r.hasData = true
			if j > 0 {
				r.parsed = true
			}

			r.parseMeta(val.Index(j).Interface(), j)
		}
	} else {
		r.hasData = false
		r.parseMeta(a, 0)
	}

	return r
}

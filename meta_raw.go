package excelizex

import (
	"reflect"
	"strings"
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
}

// MetaRaw 用于存储表元原始信息
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

type metaRaws struct {
	cursor int
	parsed bool
	raws   []*metaRaw
	set    map[int]struct{}

	hasData bool
	data    [][]any
}

func (mr *metaRaws) sheet(sheetName string) *Sheet {
	s := &Sheet{
		name:     sheetName,
		styleRef: make(map[int][]string),
	}

	if mr.hasData {
		s.data = &mr.data
	}

	for _, raw := range mr.raws {
		if raw.part == noticePart {
			s.notice = raw.cellValue

			styles := strings.Split(raw.styleTag, "+")
			s.styleRef[-1] = styles
		}

		if raw.part == headerPart {
			s.header = append(s.header, raw.cellValue)
			styles := strings.Split(raw.styleTag, "+")
			s.styleRef[raw.colIndex] = styles
		}
	}

	return s
}

func newMetas(a any) *metaRaws {
	r := &metaRaws{
		cursor: 1,
		set:    map[int]struct{}{},
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

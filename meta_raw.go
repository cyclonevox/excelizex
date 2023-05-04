package excelizex

import (
	"errors"
	"fmt"
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
	raws   []*metaRaw
}

func newMetas(a any) *metaRaws {
	r := &metaRaws{
		cursor: 1,
		raws:   make([]*metaRaw, 0),
	}

	r.parseMeta(a)

	return r
}

func (mr *metaRaws) sheet(sheetName string) *Sheet {
	s := &Sheet{
		name:     sheetName,
		styleRef: make(map[int][]string),
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

func (mr *metaRaws) append(raw *metaRaw, moveCursor ...int) {
	mr.raws = append(mr.raws, raw)

	if len(moveCursor) != 0 {
		mr.cursor += moveCursor[0]
	}
}

func (mr *metaRaws) parseMeta(a any) {
	val := reflect.ValueOf(a)
	typ := reflect.TypeOf(a)

	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
		val = val.Elem()
	}

	if typ.Kind() != reflect.Struct {
		panic(errors.New("generate function support using struct only"))
	}

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)

		// 若字段为拓展字段则进行处理，存储为Meta原始数据
		if meta, ok := field.Interface().(HeaderMeta); ok {
			mr.append(&metaRaw{
				part:        headerPart,
				styleTag:    meta.ExtStyleTag(),
				validateTag: meta.ExtValidateTag(),
				colIndex:    mr.cursor,
				cellValue:   meta.ExtHeader(),
			}, 1)

			continue
		}

		// 2.是否包含了excel标签
		t := typ.Field(i).Tag.Get("excel")
		params := strings.Split(t, "|")
		if t != "" {
			styleTag := typ.Field(i).Tag.Get("style")
			validateTag := typ.Field(i).Tag.Get("validate")
			switch Part(params[0]) {
			case noticePart:
				// 原始数据存入metaCache
				mr.append(&metaRaw{
					part:        noticePart,
					styleTag:    styleTag,
					validateTag: validateTag,
					colIndex:    -1,
					cellValue:   field.String(),
				}, 2)
			case headerPart:
				if len(params) < 2 {
					panic(
						fmt.Sprintf(
							"%s header tag format error : %s .Using header|xxx format",
							typ.Field(i).Name, t,
						),
					)
				}

				// 原始数据存入metaCache
				mr.append(&metaRaw{
					part:        headerPart,
					styleTag:    styleTag,
					validateTag: validateTag,
					colIndex:    mr.cursor,
					cellValue:   params[1],
				}, 1)
			}
		}

		// 结构体字段做递归处理
		if field.Kind() == reflect.Struct && (t != "" && params[0] == "extend") {
			mr.parseMeta(field.Interface())
		}

		// Slice字段则进入下层循环
		if field.Kind() == reflect.Slice && (t != "" && params[0] == "extend") {
			for j := 0; j < field.Len(); j++ {
				if _, ok := field.Index(j).Interface().(HeaderMeta); ok {
					mr.parseExtMetaList(field.Index(j))

					continue
				}

				mr.parseMeta(field.Index(j).Interface())
			}
		}

	}
}

func (mr *metaRaws) parseExtMetaList(field reflect.Value) {
	// 若字段为拓展字段则进行处理，存储为Meta原始数据
	if meta, ok := field.Interface().(HeaderMeta); ok {
		mr.append(&metaRaw{
			part:        headerPart,
			styleTag:    meta.ExtStyleTag(),
			validateTag: meta.ExtValidateTag(),
			colIndex:    mr.cursor,
			cellValue:   meta.ExtHeader(),
		}, 1)
	}
}

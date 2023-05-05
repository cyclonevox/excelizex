package excelizex

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

func (mr *metaRaws) append(raw *metaRaw, sliceIndex int, data ...any) {
	if !mr.parsed {
		mr.raws = append(mr.raws, raw)
		mr.set[raw.colIndex] = struct{}{}
		mr.cursor += 1
	}

	if mr.hasData && len(data) > 0 && raw.part != noticePart {
		mr.data[sliceIndex] = append(mr.data[sliceIndex], data)
	}

}

func (mr *metaRaws) parseMeta(a any, sliceIndex int) {
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
			}, sliceIndex, meta.ExtData())

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
				}, sliceIndex)
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
				}, sliceIndex, field.Interface())
			}
		}

		// 结构体字段做递归处理
		if field.Kind() == reflect.Struct && (t != "" && params[0] == "extend") {
			mr.parseMeta(field.Interface(), sliceIndex)
		}

		// Slice字段则进入下层循环
		if field.Kind() == reflect.Slice && (t != "" && params[0] == "extend") {
			for j := 0; j < field.Len(); j++ {
				if _, ok := field.Index(j).Interface().(HeaderMeta); ok {
					mr.parseExtMetaList(field.Index(j), sliceIndex)

					continue
				}

				mr.parseMeta(field.Index(j).Interface(), sliceIndex)
			}
		}

	}

	return
}

func (mr *metaRaws) parseExtMetaList(field reflect.Value, sliceIndex int) {
	// 若字段为拓展字段则进行处理，存储为Meta原始数据
	if meta, ok := field.Interface().(HeaderMeta); ok {
		mr.append(&metaRaw{
			part:        headerPart,
			styleTag:    meta.ExtStyleTag(),
			validateTag: meta.ExtValidateTag(),
			colIndex:    mr.cursor,
			cellValue:   meta.ExtHeader(),
		}, sliceIndex, meta.ExtData())
	}
}

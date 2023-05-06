package excelizex

import "strings"

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
			if raw.styleTag != "" {
				styles := strings.Split(raw.styleTag, "+")
				s.styleRef[-1] = styles
			}
		}

		if raw.part == headerPart {
			s.header = append(s.header, raw.cellValue)
			if raw.styleTag != "" {
				styles := strings.Split(raw.styleTag, "+")
				s.styleRef[raw.colIndex] = styles
			}
		}
	}

	return s
}

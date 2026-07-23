package excelizex

import (
	"github.com/cyclonevox/excelizex/v2/layout"
	"github.com/cyclonevox/excelizex/v2/schema"
)

// Sheet is a fluent handle for one worksheet.
type Sheet struct {
	wb     *Workbook
	name   string
	layout layout.Layout
	schema *schema.Schema
	notice string
}

// Sheet selects a worksheet by name.
func (wb *Workbook) Sheet(name string) *Sheet {
	return &Sheet{
		wb:     wb,
		name:   name,
		layout: layout.NoticeHeaderData{},
	}
}

// WithLayout sets the physical layout for this sheet.
func (s *Sheet) WithLayout(l layout.Layout) *Sheet {
	if l != nil {
		s.layout = l
	}

	return s
}

// WithSchema overrides schema parsed from the row type.
func (s *Sheet) WithSchema(sc schema.Schema) *Sheet {
	cp := sc
	s.schema = &cp

	return s
}

// WithNotice sets the expected notice text for validation (optional).
func (s *Sheet) WithNotice(text string) *Sheet {
	s.notice = text

	return s
}

// Read begins a generic read pipeline for row type T.
// Use as: wb.Sheet("导入").WithLayout(...).Read[StudentRow]()
func Read[T any](s *Sheet) *ReadBuilder[T] {
	return &ReadBuilder[T]{
		sheet:       s,
		validators:  nil,
		concurrency: 1,
	}
}

// Write begins a generic write pipeline for row type T.
// Use as: wb.Sheet("导入").WithLayout(...).Write[StudentRow]().Rows(...).Apply()
func Write[T any](s *Sheet) *WriteBuilder[T] {
	return &WriteBuilder[T]{
		sheet: s,
	}
}

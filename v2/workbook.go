package excelizex

import (
	"bytes"
	"io"

	"github.com/cyclonevox/excelizex/v2/style"
	"github.com/xuri/excelize/v2"
)

// Workbook wraps an excelize file for fluent read/write operations.
type Workbook struct {
	f      *excelize.File
	styles *style.Registry
}

// New creates an empty workbook.
func New() *Workbook {
	f := excelize.NewFile()
	wb := &Workbook{f: f, styles: style.NewRegistry(f)}
	_ = wb.styles.RegisterDefaults()

	return wb
}

// Open reads a workbook from r.
func Open(r io.Reader) (*Workbook, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	f, err := excelize.OpenReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	wb := &Workbook{f: f, styles: style.NewRegistry(f)}
	_ = wb.styles.RegisterDefaults()

	return wb, nil
}

// RegisterStyle adds a custom named style for write operations.
func (wb *Workbook) RegisterStyle(s style.Style) error {
	if wb.styles == nil {
		wb.styles = style.NewRegistry(wb.f)
	}

	return wb.styles.Register(s)
}

// File returns the underlying excelize file.
func (wb *Workbook) File() *excelize.File {
	return wb.f
}

// Save writes the workbook to w.
func (wb *Workbook) Save(w io.Writer) error {
	return wb.f.Write(w)
}

// Close closes the underlying excelize file.
func (wb *Workbook) Close() error {
	return wb.f.Close()
}

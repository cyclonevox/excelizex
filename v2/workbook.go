package excelizex

import (
	"bytes"
	"fmt"
	"io"
	"sync"

	"github.com/cyclonevox/excelizex/v2/style"
	"github.com/xuri/excelize/v2"
)

// Workbook wraps an excelize file for fluent read/write operations.
// Library methods serialize access to the underlying excelize file via mu.
// File() returns the raw *excelize.File without locking; callers must synchronize
// their own use of that handle.
type Workbook struct {
	mu     sync.Mutex
	f      *excelize.File
	styles *style.Registry
}

// New creates an empty workbook. It always returns a non-nil Workbook; if built-in
// style registration fails, the workbook is still usable but named default styles
// may be incomplete.
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
	if err := wb.styles.RegisterDefaults(); err != nil {
		f.Close()

		return nil, err
	}

	return wb, nil
}

// RegisterStyle adds a custom named style for write operations.
func (wb *Workbook) RegisterStyle(s style.Style) error {
	wb.mu.Lock()
	defer wb.mu.Unlock()
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
	wb.mu.Lock()
	defer wb.mu.Unlock()
	if wb.f == nil {
		return fmt.Errorf("workbook: closed")
	}

	return wb.f.Write(w)
}

func (wb *Workbook) getRowsLocked(sheet string) ([][]string, error) {
	wb.mu.Lock()
	defer wb.mu.Unlock()
	if wb.f == nil {
		return nil, fmt.Errorf("workbook: closed")
	}

	return wb.f.GetRows(sheet)
}

// Close closes the underlying excelize file.
func (wb *Workbook) Close() error {
	if wb == nil {
		return nil
	}
	wb.mu.Lock()
	defer wb.mu.Unlock()
	if wb.f == nil {
		return nil
	}
	err := wb.f.Close()
	wb.f = nil

	return err
}

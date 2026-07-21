package excelizex

import (
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
	mu         sync.Mutex
	f          *excelize.File
	styles     *style.Registry
	createdNew bool // true only for New(); Open() workbooks never auto-delete sheets
}

// OpenOption configures Open.
type OpenOption func(*excelize.Options)

// WithUnzipSizeLimit sets the maximum unzipped size accepted when opening a workbook.
func WithUnzipSizeLimit(n int64) OpenOption {
	return func(o *excelize.Options) {
		o.UnzipSizeLimit = n
	}
}

// WithUnzipXMLSizeLimit sets the memory limit for unzipping worksheet XML.
func WithUnzipXMLSizeLimit(n int64) OpenOption {
	return func(o *excelize.Options) {
		o.UnzipXMLSizeLimit = n
	}
}

// New creates an empty workbook. It always returns a non-nil Workbook; if built-in
// style registration fails, the workbook is still usable but named default styles
// may be incomplete.
func New() *Workbook {
	f := excelize.NewFile()
	wb := &Workbook{f: f, styles: style.NewRegistry(f), createdNew: true}
	_ = wb.styles.RegisterDefaults()

	return wb
}

// Open reads a workbook from r. It does not buffer the whole stream before
// handing it to excelize; pass size-limit options for untrusted uploads.
func Open(r io.Reader, opts ...OpenOption) (*Workbook, error) {
	var o excelize.Options
	for _, opt := range opts {
		if opt != nil {
			opt(&o)
		}
	}
	f, err := excelize.OpenReader(r, o)
	if err != nil {
		return nil, err
	}
	wb := &Workbook{f: f, styles: style.NewRegistry(f), createdNew: false}
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

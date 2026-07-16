package excelizex

import (
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/cyclonevox/excelizex/v2/layout"
	"github.com/xuri/excelize/v2"
)

const errorColumnHeader = "错误原因"

// RowError describes one failed row.
type RowError struct {
	Row      int
	RawCells []string
	Messages []string
}

// Result collects row-level errors during read.
type Result struct {
	SheetName    string
	Notice       string
	Headers      []string
	layout       layout.Layout
	dataStartRow int
	headerRow    int

	mu     sync.Mutex
	errors []RowError
}

func newResult(sheet string, lyt layout.Layout) *Result {
	start, _ := lyt.HeaderRows()

	return &Result{
		SheetName:    sheet,
		layout:       lyt,
		dataStartRow: lyt.DataStartRow(),
		headerRow:    start,
	}
}

func (r *Result) setHeaders(headers []string) {
	r.Headers = append([]string(nil), headers...)
}

func (r *Result) setNotice(text string) {
	r.Notice = text
}

func (r *Result) addError(row int, cells []string, msgs ...string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.errors = append(r.errors, RowError{
		Row:      row,
		RawCells: append([]string(nil), cells...),
		Messages: append([]string(nil), msgs...),
	})
}

// Errors returns a copy of collected row errors.
func (r *Result) Errors() []RowError {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]RowError, len(r.errors))
	copy(out, r.errors)

	return out
}

// HasErrors reports whether any row errors were collected.
func (r *Result) HasErrors() bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	return len(r.errors) > 0
}

// WriteErrors rewrites the sheet keeping only failed rows and appending an error column.
func (wb *Workbook) WriteErrors(res *Result) error {
	if res == nil || !res.HasErrors() {
		return nil
	}
	sheet := res.SheetName
	rows, err := wb.f.GetRows(sheet)
	if err != nil {
		return fmt.Errorf("write errors: get rows: %w", err)
	}

	if err := wb.removeDataRows(sheet, res.dataStartRow, len(rows)); err != nil {
		return err
	}

	headers := append([]string(nil), res.Headers...)
	if len(headers) == 0 {
		return fmt.Errorf("write errors: missing headers")
	}
	if !hasErrorColumn(headers) {
		headers = append(headers, errorColumnHeader)
	}
	headerAddr, err := excelize.JoinCellName("A", res.headerRow)
	if err != nil {
		return err
	}
	if err := wb.f.SetSheetRow(sheet, headerAddr, &headers); err != nil {
		return fmt.Errorf("write errors: set header: %w", err)
	}

	errCol, err := excelize.ColumnNumberToName(len(headers))
	if err != nil {
		return err
	}

	writeRow := res.dataStartRow
	for _, e := range res.Errors() {
		row := padRow(e.RawCells, len(headers)-1)
		addr, err := excelize.JoinCellName("A", writeRow)
		if err != nil {
			return err
		}
		if err := wb.f.SetSheetRow(sheet, addr, &row); err != nil {
			return fmt.Errorf("write errors: set row: %w", err)
		}
		msgAddr, err := excelize.JoinCellName(errCol, writeRow)
		if err != nil {
			return err
		}
		if err := wb.f.SetCellStr(sheet, msgAddr, strings.Join(e.Messages, "; ")); err != nil {
			return fmt.Errorf("write errors: set error cell: %w", err)
		}
		writeRow++
	}

	return nil
}

func (wb *Workbook) removeDataRows(sheet string, startRow, totalRows int) error {
	for row := totalRows; row >= startRow; row-- {
		if err := wb.f.RemoveRow(sheet, row); err != nil {
			return fmt.Errorf("write errors: remove row %d: %w", row, err)
		}
	}

	return nil
}

func hasErrorColumn(headers []string) bool {
	if len(headers) == 0 {
		return false
	}

	return headers[len(headers)-1] == errorColumnHeader
}

func padRow(cells []string, width int) []string {
	if len(cells) >= width {
		out := make([]string, width)
		copy(out, cells[:width])

		return out
	}
	out := make([]string, width)
	copy(out, cells)

	return out
}

// DataStartRow returns the 1-based data start row captured from layout.
func (r *Result) DataStartRow() int {
	return r.dataStartRow
}

// HeaderRow returns the 1-based header start row.
func (r *Result) HeaderRow() int {
	return r.headerRow
}

// StringRow converts row number to excelize row reference helper.
func StringRow(n int) string {
	return strconv.Itoa(n)
}

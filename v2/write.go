package excelizex

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/cyclonevox/excelizex/v2/bind"
	"github.com/cyclonevox/excelizex/v2/convert"
	"github.com/cyclonevox/excelizex/v2/layout"
	"github.com/cyclonevox/excelizex/v2/schema"
	"github.com/cyclonevox/excelizex/v2/style"
	"github.com/xuri/excelize/v2"
)

const optionsSheetSuffix = "选项数据表"

type dropdownSpec struct {
	header  string
	options []string
}

// WriteBuilder configures and executes a typed write pipeline.
type WriteBuilder[T any] struct {
	sheet           *Sheet
	converters      convert.ExportRegistry
	rows            []T
	templateOnly    bool
	dropdowns       []dropdownSpec
	protectPassword string
}

// Convert registers a named export converter for this write.
func (w *WriteBuilder[T]) Convert(name string, fn convert.ExportFunc) *WriteBuilder[T] {
	if w.converters == nil {
		w.converters = make(convert.ExportRegistry)
	}
	w.converters[name] = fn

	return w
}

// Rows sets data rows to write under the layout.
func (w *WriteBuilder[T]) Rows(rows ...T) *WriteBuilder[T] {
	w.rows = append(w.rows, rows...)

	return w
}

// Template marks the write as header/notice only (no data rows).
func (w *WriteBuilder[T]) Template() *WriteBuilder[T] {
	w.templateOnly = true

	return w
}

// Dropdown sets single-level validation options for a logical header name.
func (w *WriteBuilder[T]) Dropdown(header string, options []string) *WriteBuilder[T] {
	if header == "" {
		return w
	}
	opts := append([]string(nil), options...)
	w.dropdowns = append(w.dropdowns, dropdownSpec{header: header, options: opts})

	return w
}

// Protect sets an optional sheet password applied on Apply.
func (w *WriteBuilder[T]) Protect(password string) *WriteBuilder[T] {
	w.protectPassword = password

	return w
}

// Apply writes the sheet content (template and/or rows).
func (w *WriteBuilder[T]) Apply() error {
	w.sheet.wb.mu.Lock()
	defer w.sheet.wb.mu.Unlock()
	if w.sheet.wb.f == nil {
		return fmt.Errorf("workbook: closed")
	}

	sc, err := w.schemaForWrite()
	if err != nil {
		return err
	}
	lyt := w.sheet.layout
	if lyt == nil {
		lyt = layout.NoticeHeaderData{}
	}
	if err := ensureSheet(w.sheet.wb, w.sheet.name); err != nil {
		return err
	}
	if err := cleanupDefaultSheet(w.sheet.wb, w.sheet.name); err != nil {
		return err
	}

	noticeText := w.sheet.notice
	if noticeText == "" && len(w.rows) > 0 {
		var err error
		noticeText, err = noticeFromRow(sc, w.rows[0])
		if err != nil {
			return fmt.Errorf("write: notice: %w", err)
		}
	}

	if err := writeNotice(w.sheet.wb, w.sheet.name, lyt, noticeText); err != nil {
		return err
	}
	if err := writeHeaders(w.sheet.wb, w.sheet.name, lyt, sc); err != nil {
		return err
	}

	dataCount := 0
	if !w.templateOnly && len(w.rows) > 0 {
		dataCount, err = writeDataRows(w.sheet.wb, w.sheet.name, lyt, sc, w.rows, w.converters)
		if err != nil {
			return err
		}
	}

	if err := applyStyles(w.sheet.wb, w.sheet.name, lyt, sc, dataCount); err != nil {
		return err
	}
	if err := applyDropdowns(w.sheet.wb, w.sheet.name, lyt, sc, w.dropdowns); err != nil {
		return err
	}
	if w.protectPassword != "" {
		if err := protectSheet(w.sheet.wb.f, w.sheet.name, w.protectPassword); err != nil {
			return fmt.Errorf("write: protect sheet: %w", err)
		}
	}

	return nil
}

// ExportTo registers a typed named export converter on this write builder.
func ExportTo[T any, V any](w *WriteBuilder[T], name string, fn func(V) (string, error)) *WriteBuilder[T] {
	convert.ExportTo(w.converters, name, fn)

	return w
}

func (w *WriteBuilder[T]) schemaForWrite() (schema.Schema, error) {
	if w.sheet.schema != nil {
		return *w.sheet.schema, nil
	}
	var zero T

	return schema.New(zero)
}

func ensureSheet(wb *Workbook, name string) error {
	idx, err := wb.f.GetSheetIndex(name)
	if err != nil {
		return fmt.Errorf("write: sheet index: %w", err)
	}
	if idx != -1 {
		return nil
	}
	if _, err := wb.f.NewSheet(name); err != nil {
		return fmt.Errorf("write: new sheet: %w", err)
	}

	return nil
}

func cleanupDefaultSheet(wb *Workbook, name string) error {
	if name == "Sheet1" {
		return nil
	}
	idx, err := wb.f.GetSheetIndex("Sheet1")
	if err != nil {
		return err
	}
	if idx == -1 {
		return nil
	}

	return wb.f.DeleteSheet("Sheet1")
}

func noticeFromRow[T any](sc schema.Schema, row T) (string, error) {
	if sc.Notice == "" {
		return "", nil
	}
	v := reflect.ValueOf(row)
	for v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return "", nil
		}
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return "", nil
	}
	f := v.FieldByName(sc.Notice)
	if !f.IsValid() {
		return "", nil
	}

	return convert.From(f, "", "", nil)
}

func writeNotice(wb *Workbook, sheet string, lyt layout.Layout, text string) error {
	row, ok := lyt.NoticeRow()
	if !ok || text == "" {
		return nil
	}
	addr, err := excelize.JoinCellName("A", row)
	if err != nil {
		return err
	}
	if err := wb.f.SetCellStr(sheet, addr, text); err != nil {
		return fmt.Errorf("write: notice: %w", err)
	}
	if wb.styles != nil {
		if id, err := wb.styles.Resolve("notice"); err == nil {
			if err := wb.f.SetCellStyle(sheet, addr, addr, id); err != nil {
				return fmt.Errorf("write: notice style: %w", err)
			}
		}
	}

	return nil
}

func writeHeaders(wb *Workbook, sheet string, lyt layout.Layout, sc schema.Schema) error {
	start, _ := lyt.HeaderRows()
	headers := make([]string, len(sc.Columns))
	for i, c := range sc.Columns {
		headers[i] = c.Header
	}
	addr, err := excelize.JoinCellName("A", start)
	if err != nil {
		return err
	}
	if err := wb.f.SetSheetRow(sheet, addr, &headers); err != nil {
		return fmt.Errorf("write: header: %w", err)
	}

	return nil
}

func writeDataRows[T any](wb *Workbook, sheet string, lyt layout.Layout, sc schema.Schema, rows []T, reg convert.ExportRegistry) (int, error) {
	dataStart := lyt.DataStartRow()
	for i, row := range rows {
		cells, err := bind.ExportRow(sc, row, reg)
		if err != nil {
			return i, fmt.Errorf("write: row %d: %w", i+1, err)
		}
		addr, err := excelize.JoinCellName("A", dataStart+i)
		if err != nil {
			return i, err
		}
		if err := wb.f.SetSheetRow(sheet, addr, &cells); err != nil {
			return i, fmt.Errorf("write: set row: %w", err)
		}
	}

	return len(rows), nil
}

func applyStyles(wb *Workbook, sheet string, lyt layout.Layout, sc schema.Schema, dataRows int) error {
	if wb.styles == nil {
		return nil
	}
	headerRow, _ := lyt.HeaderRows()
	dataStart := lyt.DataStartRow()
	endDataRow := dataStart + dataRows - 1
	if dataRows == 0 {
		endDataRow = dataStart + 999
	}
	for i, col := range sc.Columns {
		colName, err := excelize.ColumnNumberToName(i + 1)
		if err != nil {
			return err
		}
		headerCell := colName + strconv.Itoa(headerRow)
		if parts := style.SplitRole(col.Style, 0); len(parts) > 0 {
			id, err := wb.styles.Resolve(parts...)
			if err != nil {
				return fmt.Errorf("write: header style: %w", err)
			}
			if err := wb.f.SetCellStyle(sheet, headerCell, headerCell, id); err != nil {
				return fmt.Errorf("write: set header style: %w", err)
			}
		}
		if parts := style.SplitRole(col.Style, 1); len(parts) > 0 {
			id, err := wb.styles.Resolve(parts...)
			if err != nil {
				return fmt.Errorf("write: body style: %w", err)
			}
			top := colName + strconv.Itoa(dataStart)
			bottom := colName + strconv.Itoa(endDataRow)
			if err := wb.f.SetCellStyle(sheet, top, bottom, id); err != nil {
				return fmt.Errorf("write: set body style: %w", err)
			}
		}
	}

	return nil
}

func applyDropdowns(wb *Workbook, sheet string, lyt layout.Layout, sc schema.Schema, specs []dropdownSpec) error {
	if len(specs) == 0 {
		return nil
	}
	optionsSheet := sheet + optionsSheetSuffix
	idx, err := wb.f.GetSheetIndex(optionsSheet)
	if err != nil {
		return fmt.Errorf("write: options sheet index: %w", err)
	}
	if idx == -1 {
		if _, err := wb.f.NewSheet(optionsSheet); err != nil {
			return fmt.Errorf("write: options sheet: %w", err)
		}
	}
	dataStart := lyt.DataStartRow()
	for i, spec := range specs {
		colIdx := columnIndex(sc, spec.header)
		if colIdx < 0 {
			return fmt.Errorf("write: dropdown: unknown header %q", spec.header)
		}
		colName, err := excelize.ColumnNumberToName(colIdx + 1)
		if err != nil {
			return err
		}
		rowVals := make([]any, len(spec.options))
		for j, o := range spec.options {
			rowVals[j] = o
		}
		optAddr, err := excelize.JoinCellName("A", i+1)
		if err != nil {
			return err
		}
		if err := wb.f.SetSheetRow(optionsSheet, optAddr, &rowVals); err != nil {
			return fmt.Errorf("write: dropdown options: %w", err)
		}
		endCol, err := excelize.ColumnNumberToName(len(spec.options))
		if err != nil {
			return err
		}
		formula := fmt.Sprintf("%s!$A$%d:$%s$%d", optionsSheet, i+1, endCol, i+1)
		dv := excelize.NewDataValidation(true)
		dv.Sqref = fmt.Sprintf("%s%d:%s1048576", colName, dataStart, colName)
		dv.SetSqrefDropList(formula)
		dv.ShowInputMessage = true
		dv.ShowErrorMessage = true
		msg := "请按下拉框中的文本进行正确填写"
		dv.Error = &msg
		if err := wb.f.AddDataValidation(sheet, dv); err != nil {
			return fmt.Errorf("write: add validation: %w", err)
		}
	}
	if err := wb.f.SetSheetVisible(optionsSheet, false); err != nil {
		return fmt.Errorf("write: hide options sheet: %w", err)
	}

	return nil
}

func columnIndex(sc schema.Schema, header string) int {
	for i, c := range sc.Columns {
		if c.Header == header {
			return i
		}
	}

	return -1
}

func protectSheet(f *excelize.File, sheet, password string) error {
	return f.ProtectSheet(sheet, &excelize.SheetProtectionOptions{
		Password:            password,
		EditObjects:         true,
		EditScenarios:       true,
		SelectLockedCells:   true,
		SelectUnlockedCells: true,
	})
}
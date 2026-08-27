package excelizex

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/cyclonevox/excelizex/bind"
	"github.com/cyclonevox/excelizex/convert"
	"github.com/cyclonevox/excelizex/layout"
	"github.com/cyclonevox/excelizex/schema"
	"github.com/cyclonevox/excelizex/style"
	"github.com/xuri/excelize/v2"
)

const (
	optionsSheetSuffix = "_opts"
	treeSheetSuffix    = "_tree"
	maxSheetNameLen    = 31
)

type dropdownSpec struct {
	header  string
	options []string
}

type cascadeSpec struct {
	child  string
	parent string
	tree   map[string][]string
}

// WriteBuilder configures and executes a typed write pipeline.
type WriteBuilder[T any] struct {
	sheet           *Sheet
	rows            []T
	templateOnly    bool
	dropdowns       []dropdownSpec
	cascades        []cascadeSpec
	extraHeaders    []string
	extraSet        bool
	streamFn        func(yield func(T) error) error
	protectPassword string
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

// Extend sets extra column headers for a map[string]string field tagged excel:",extend".
// Column order follows this list. Stream writes require Extend when extra columns are needed
// (the callback cannot be scanned ahead of time). Without Extend, Rows() uses a sorted union
// of keys found on the data rows.
func (w *WriteBuilder[T]) Extend(headers ...string) *WriteBuilder[T] {
	w.extraHeaders = append(w.extraHeaders, headers...)
	w.extraSet = true

	return w
}

// Stream writes rows from src via excelize StreamWriter. Do not call other methods on the
// same Workbook from yield (Apply holds the workbook lock). Cannot be combined with Rows or Template.
func (w *WriteBuilder[T]) Stream(src func(yield func(T) error) error) *WriteBuilder[T] {
	w.streamFn = src

	return w
}

// Cascade adds a dependent dropdown: child options depend on the parent cell via INDIRECT
// named ranges. Parent values must be valid Excel defined names (no spaces). If the parent
// column has no Dropdown yet, one is added from the map keys (sorted).
func (w *WriteBuilder[T]) Cascade(child, parent string, tree map[string][]string) *WriteBuilder[T] {
	if child == "" || parent == "" {
		return w
	}
	cp := make(map[string][]string, len(tree))
	for k, v := range tree {
		cp[k] = append([]string(nil), v...)
	}
	w.cascades = append(w.cascades, cascadeSpec{child: child, parent: parent, tree: cp})

	return w
}

// Protect sets an optional sheet password applied on Apply.
func (w *WriteBuilder[T]) Protect(password string) *WriteBuilder[T] {
	w.protectPassword = password

	return w
}

// Apply rebuilds the sheet content (template and/or rows).
// It replaces the target sheet: previous data rows, notice text, data validations,
// and companion options sheets from earlier Apply calls are cleared first.
func (w *WriteBuilder[T]) Apply() error {
	w.sheet.wb.mu.Lock()
	defer w.sheet.wb.mu.Unlock()
	if w.sheet.wb.f == nil {
		return fmt.Errorf("workbook: closed")
	}
	if w.streamFn != nil && len(w.rows) > 0 {
		return fmt.Errorf("write: Rows and Stream cannot be used together")
	}
	if w.streamFn != nil && w.templateOnly {
		return fmt.Errorf("write: Template and Stream cannot be used together")
	}

	sc, err := w.schemaForWrite()
	if err != nil {
		return err
	}
	lyt := w.sheet.layout
	if lyt == nil {
		lyt = layout.NoticeHeaderData{}
	}
	extra, err := w.resolveExtraHeaders(sc)
	if err != nil {
		return err
	}
	if err := w.validateCascades(); err != nil {
		return err
	}
	headers := writeHeadersList(sc, extra)

	if err := ensureSheet(w.sheet.wb, w.sheet.name); err != nil {
		return err
	}
	if err := cleanupDefaultSheet(w.sheet.wb, w.sheet.name); err != nil {
		return err
	}
	keepOptions := len(w.dropdowns) > 0 || len(w.cascades) > 0
	if err := resetSheetForApply(w.sheet.wb, w.sheet.name, keepOptions); err != nil {
		return err
	}
	if err := resetTreeSheet(w.sheet.wb, w.sheet.name, len(w.cascades) > 0); err != nil {
		return err
	}

	if w.streamFn != nil {
		return w.applyStream(sc, lyt, extra, headers)
	}

	return w.applyCells(sc, lyt, extra, headers)
}

func (w *WriteBuilder[T]) applyCells(sc schema.Schema, lyt layout.Layout, extra, headers []string) error {
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
	if err := writeHeaders(w.sheet.wb, w.sheet.name, lyt, headers); err != nil {
		return err
	}
	headerStart, _ := lyt.HeaderRows()
	if err := applyHeaderMerges(w.sheet.wb, w.sheet.name, lyt, headerStart, len(headers), nil); err != nil {
		return err
	}

	dataCount := 0
	if !w.templateOnly && len(w.rows) > 0 {
		var err error
		dataCount, err = writeDataRows(w.sheet.wb, w.sheet.name, lyt, sc, w.rows, extra)
		if err != nil {
			return err
		}
	}

	if err := applyStyles(w.sheet.wb, w.sheet.name, lyt, sc, dataCount); err != nil {
		return err
	}
	if err := applyDropdowns(w.sheet.wb, w.sheet.name, lyt, headers, w.dropdownsForApply()); err != nil {
		return err
	}
	if err := applyCascades(w.sheet.wb, w.sheet.name, lyt, headers, w.cascades); err != nil {
		return err
	}
	if w.protectPassword != "" {
		if err := protectSheet(w.sheet.wb.f, w.sheet.name, w.protectPassword); err != nil {
			return fmt.Errorf("write: protect sheet: %w", err)
		}
	}

	return nil
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

// cleanupDefaultSheet removes excelize's blank Sheet1 only for workbooks created
// by New(), and only when Sheet1 is still empty. Open() workbooks never auto-delete.
func cleanupDefaultSheet(wb *Workbook, name string) error {
	if !wb.createdNew || name == "Sheet1" {
		return nil
	}
	idx, err := wb.f.GetSheetIndex("Sheet1")
	if err != nil {
		return err
	}
	if idx == -1 {
		return nil
	}
	rows, err := wb.f.GetRows("Sheet1")
	if err != nil {
		return err
	}
	if !sheetRowsEmpty(rows) {
		return nil
	}

	return wb.f.DeleteSheet("Sheet1")
}

func resetSheetForApply(wb *Workbook, sheet string, keepOptionsSheet bool) error {
	_ = wb.f.UnprotectSheet(sheet)
	if err := wb.f.DeleteDataValidation(sheet); err != nil {
		return fmt.Errorf("write: clear validations: %w", err)
	}
	if err := clearSheetRows(wb, sheet); err != nil {
		return err
	}

	return resetOptionsSheet(wb, sheet, keepOptionsSheet)
}

func resetTreeSheet(wb *Workbook, sheet string, keep bool) error {
	name := treeSheetName(sheet)
	if err := clearTreeDefinedNames(wb, sheet, name); err != nil {
		return err
	}
	idx, err := wb.f.GetSheetIndex(name)
	if err != nil {
		return fmt.Errorf("write: tree sheet index: %w", err)
	}
	if idx == -1 {
		return nil
	}
	if !keep {
		return wb.f.DeleteSheet(name)
	}

	return clearSheetRows(wb, name)
}

func clearTreeDefinedNames(wb *Workbook, dataSheet, treeSheet string) error {
	needle := quoteSheetForFormula(treeSheet)
	for _, dn := range wb.f.GetDefinedName() {
		if !strings.Contains(dn.RefersTo, needle) {
			continue
		}
		if dn.Scope != dataSheet && dn.Scope != "Workbook" {
			continue
		}
		if err := wb.f.DeleteDefinedName(&excelize.DefinedName{Name: dn.Name, Scope: dn.Scope}); err != nil {
			return fmt.Errorf("write: delete defined name %q: %w", dn.Name, err)
		}
	}

	return nil
}

func clearSheetRows(wb *Workbook, sheet string) error {
	rows, err := wb.f.GetRows(sheet)
	if err != nil {
		return fmt.Errorf("write: get rows: %w", err)
	}
	for row := len(rows); row >= 1; row-- {
		if err := wb.f.RemoveRow(sheet, row); err != nil {
			return fmt.Errorf("write: clear row %d: %w", row, err)
		}
	}

	return nil
}

func resetOptionsSheet(wb *Workbook, sheet string, keep bool) error {
	name := optionsSheetName(sheet)
	idx, err := wb.f.GetSheetIndex(name)
	if err != nil {
		return fmt.Errorf("write: options sheet index: %w", err)
	}
	if idx == -1 {
		return nil
	}
	if !keep {
		return wb.f.DeleteSheet(name)
	}

	return clearSheetRows(wb, name)
}

func sheetRowsEmpty(rows [][]string) bool {
	for _, row := range rows {
		for _, c := range row {
			if strings.TrimSpace(c) != "" {
				return false
			}
		}
	}

	return true
}

func optionsSheetName(sheet string) string {
	return companionSheetName(sheet, optionsSheetSuffix)
}

func treeSheetName(sheet string) string {
	return companionSheetName(sheet, treeSheetSuffix)
}

func companionSheetName(sheet, suffix string) string {
	suffixRunes := utf8.RuneCountInString(suffix)
	maxBase := maxSheetNameLen - suffixRunes
	if maxBase < 1 {
		maxBase = 1
	}
	runes := []rune(sheet)
	if len(runes) > maxBase {
		runes = runes[:maxBase]
	}

	return string(runes) + suffix
}

func quoteSheetForFormula(name string) string {
	return "'" + strings.ReplaceAll(name, "'", "''") + "'"
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

	return convert.From(f, "")
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

func writeHeaders(wb *Workbook, sheet string, lyt layout.Layout, headers []string) error {
	start, end := lyt.HeaderRows()
	headerRows, err := lyt.RenderHeaders(headers)
	if err != nil {
		return fmt.Errorf("write: render headers: %w", err)
	}
	wantRows := end - start + 1
	if len(headerRows) != wantRows {
		return fmt.Errorf("write: layout rendered %d header rows, HeaderRows expects %d", len(headerRows), wantRows)
	}
	for i, row := range headerRows {
		addr, err := excelize.JoinCellName("A", start+i)
		if err != nil {
			return err
		}
		cells := append([]any(nil), row...)
		if err := wb.f.SetSheetRow(sheet, addr, &cells); err != nil {
			return fmt.Errorf("write: header: %w", err)
		}
	}

	return nil
}

func writeDataRows[T any](wb *Workbook, sheet string, lyt layout.Layout, sc schema.Schema, rows []T, extra []string) (int, error) {
	dataStart := lyt.DataStartRow()
	for i, row := range rows {
		cells, err := bind.ExportRow(sc, row, extra...)
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
	headerStart, headerEnd := lyt.HeaderRows()
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
		if parts := style.SplitRole(col.Style, 0); len(parts) > 0 {
			id, err := wb.styles.Resolve(parts...)
			if err != nil {
				return fmt.Errorf("write: header style: %w", err)
			}
			top := colName + strconv.Itoa(headerStart)
			bottom := colName + strconv.Itoa(headerEnd)
			if err := wb.f.SetCellStyle(sheet, top, bottom, id); err != nil {
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

func applyDropdowns(wb *Workbook, sheet string, lyt layout.Layout, headers []string, specs []dropdownSpec) error {
	if len(specs) == 0 {
		return nil
	}
	optionsSheet := optionsSheetName(sheet)
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
		if len(spec.options) == 0 {
			return fmt.Errorf("write: dropdown %q: empty options", spec.header)
		}
		colIdx := columnIndex(headers, spec.header)
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
		formula := fmt.Sprintf("%s!$A$%d:$%s$%d", quoteSheetForFormula(optionsSheet), i+1, endCol, i+1)
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

func columnIndex(headers []string, header string) int {
	for i, h := range headers {
		if h == header {
			return i
		}
	}

	return -1
}

func writeHeadersList(sc schema.Schema, extra []string) []string {
	headers := make([]string, 0, len(sc.Columns)+len(extra))
	for _, c := range sc.Columns {
		headers = append(headers, c.Header)
	}

	return append(headers, extra...)
}

func (w *WriteBuilder[T]) resolveExtraHeaders(sc schema.Schema) ([]string, error) {
	if w.extraSet {
		if sc.ExtendField == "" && len(w.extraHeaders) > 0 {
			return nil, fmt.Errorf("write: Extend used but type has no excel:\",extend\" map")
		}
		if err := validateExtraHeaders(sc, w.extraHeaders); err != nil {
			return nil, err
		}

		return append([]string(nil), w.extraHeaders...), nil
	}
	if sc.ExtendField == "" || w.streamFn != nil || w.templateOnly {
		return nil, nil
	}

	return collectExtendHeaders(sc, w.rows)
}

func (w *WriteBuilder[T]) validateCascades() error {
	if len(w.cascades) == 0 {
		return nil
	}
	child := make(map[string]struct{}, len(w.cascades))
	for _, c := range w.cascades {
		if c.child == "" || c.parent == "" {
			return fmt.Errorf("write: cascade: empty header")
		}
		if c.child == c.parent {
			return fmt.Errorf("write: cascade: child and parent are both %q", c.child)
		}
		if _, dup := child[c.child]; dup {
			return fmt.Errorf("write: cascade: duplicate child header %q", c.child)
		}
		child[c.child] = struct{}{}
		if len(c.tree) == 0 {
			return fmt.Errorf("write: cascade %q: empty tree", c.child)
		}
	}
	for _, d := range w.dropdowns {
		if _, ok := child[d.header]; ok {
			return fmt.Errorf("write: cascade child %q cannot also have Dropdown", d.header)
		}
	}

	return nil
}

func validateExtraHeaders(sc schema.Schema, extra []string) error {
	seen := make(map[string]struct{}, len(extra))
	for _, h := range extra {
		if strings.TrimSpace(h) == "" {
			return fmt.Errorf("write: empty extend header")
		}
		if _, ok := sc.ColumnByHeader(h); ok {
			return fmt.Errorf("write: extend header %q collides with schema column", h)
		}
		if _, dup := seen[h]; dup {
			return fmt.Errorf("write: duplicate extend header %q", h)
		}
		seen[h] = struct{}{}
	}

	return nil
}

func collectExtendHeaders[T any](sc schema.Schema, rows []T) ([]string, error) {
	seen := make(map[string]struct{})
	var out []string
	for _, row := range rows {
		m, err := bind.ExtendMap(sc, row)
		if err != nil {
			return nil, err
		}
		keys := make([]string, 0, len(m))
		for k := range m {
			if strings.TrimSpace(k) == "" {
				continue
			}
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			if _, ok := seen[k]; ok {
				continue
			}
			if _, inSchema := sc.ColumnByHeader(k); inSchema {
				return nil, fmt.Errorf("write: extend header %q collides with schema column", k)
			}
			seen[k] = struct{}{}
			out = append(out, k)
		}
	}

	return out, nil
}

func (w *WriteBuilder[T]) dropdownsForApply() []dropdownSpec {
	out := append([]dropdownSpec(nil), w.dropdowns...)
	have := make(map[string]struct{}, len(out))
	for _, d := range out {
		have[d.header] = struct{}{}
	}
	for _, c := range w.cascades {
		if _, ok := have[c.parent]; ok {
			continue
		}
		keys := make([]string, 0, len(c.tree))
		for k := range c.tree {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		out = append(out, dropdownSpec{header: c.parent, options: keys})
		have[c.parent] = struct{}{}
	}

	return out
}

func applyHeaderMerges(wb *Workbook, sheet string, lyt layout.Layout, headerStart, nCols int, sw *excelize.StreamWriter) error {
	m, ok := lyt.(layout.Merger)
	if !ok {
		return nil
	}
	for _, mg := range m.HeaderMerges(headerStart, nCols) {
		from, err := excelize.CoordinatesToCellName(mg.From+1, mg.Row)
		if err != nil {
			return err
		}
		to, err := excelize.CoordinatesToCellName(mg.To+1, mg.Row)
		if err != nil {
			return err
		}
		if sw != nil {
			if err := sw.MergeCell(from, to); err != nil {
				return fmt.Errorf("write: merge: %w", err)
			}

			continue
		}
		if err := wb.f.MergeCell(sheet, from, to); err != nil {
			return fmt.Errorf("write: merge: %w", err)
		}
	}

	return nil
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

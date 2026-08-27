package excelizex

import (
	"fmt"

	"github.com/cyclonevox/excelizex/bind"
	"github.com/cyclonevox/excelizex/layout"
	"github.com/cyclonevox/excelizex/schema"
	"github.com/cyclonevox/excelizex/style"
	"github.com/xuri/excelize/v2"
)

func (w *WriteBuilder[T]) applyStream(sc schema.Schema, lyt layout.Layout, extra, headers []string) error {
	sheet := w.sheet.name
	sw, err := w.sheet.wb.f.NewStreamWriter(sheet)
	if err != nil {
		return fmt.Errorf("write: stream writer: %w", err)
	}
	flushed := false
	defer func() {
		if !flushed {
			_ = sw.Flush()
		}
	}()

	headerIDs, err := columnStyleIDs(w.sheet.wb, sc, len(extra), 0)
	if err != nil {
		return err
	}
	bodyIDs, err := columnStyleIDs(w.sheet.wb, sc, len(extra), 1)
	if err != nil {
		return err
	}
	noticeID := 0
	if w.sheet.wb.styles != nil {
		if id, err := w.sheet.wb.styles.Resolve("notice"); err == nil {
			noticeID = id
		}
	}

	if row, ok := lyt.NoticeRow(); ok && w.sheet.notice != "" {
		if err := streamSetRow(sw, row, []string{w.sheet.notice}, []int{noticeID}); err != nil {
			return fmt.Errorf("write: notice: %w", err)
		}
	}

	start, end := lyt.HeaderRows()
	headerRows, err := lyt.RenderHeaders(headers)
	if err != nil {
		return fmt.Errorf("write: render headers: %w", err)
	}
	if len(headerRows) != end-start+1 {
		return fmt.Errorf("write: layout rendered %d header rows, HeaderRows expects %d", len(headerRows), end-start+1)
	}
	for i, row := range headerRows {
		if err := streamSetRow(sw, start+i, anyRowToStrings(row), headerIDs); err != nil {
			return fmt.Errorf("write: header: %w", err)
		}
	}
	if err := applyHeaderMerges(w.sheet.wb, sheet, lyt, start, len(headers), sw); err != nil {
		return err
	}

	dataStart := lyt.DataStartRow()
	n := 0
	if w.streamFn != nil {
		err := w.streamFn(func(row T) error {
			cells, err := bind.ExportRow(sc, row, extra...)
			if err != nil {
				return err
			}
			if err := streamSetRow(sw, dataStart+n, cells, bodyIDs); err != nil {
				return err
			}
			n++

			return nil
		})
		if err != nil {
			return fmt.Errorf("write: stream: %w", err)
		}
	}

	if err := applyDropdowns(w.sheet.wb, sheet, lyt, headers, w.dropdownsForApply()); err != nil {
		return err
	}
	if err := applyCascades(w.sheet.wb, sheet, lyt, headers, w.cascades); err != nil {
		return err
	}
	if w.protectPassword != "" {
		if err := protectSheet(w.sheet.wb.f, sheet, w.protectPassword); err != nil {
			return fmt.Errorf("write: protect sheet: %w", err)
		}
	}
	if err := sw.Flush(); err != nil {
		return fmt.Errorf("write: stream flush: %w", err)
	}
	flushed = true

	return nil
}

func columnStyleIDs(wb *Workbook, sc schema.Schema, extra, role int) ([]int, error) {
	ids := make([]int, len(sc.Columns)+extra)
	if wb.styles == nil {
		return ids, nil
	}
	for i, col := range sc.Columns {
		parts := style.SplitRole(col.Style, role)
		if len(parts) == 0 {
			continue
		}
		id, err := wb.styles.Resolve(parts...)
		if err != nil {
			return nil, fmt.Errorf("write: style: %w", err)
		}
		ids[i] = id
	}

	return ids, nil
}

func streamSetRow(sw *excelize.StreamWriter, row int, values []string, styleIDs []int) error {
	addr, err := excelize.JoinCellName("A", row)
	if err != nil {
		return err
	}
	cells := make([]interface{}, len(values))
	for i, v := range values {
		if i < len(styleIDs) && styleIDs[i] != 0 {
			cells[i] = excelize.Cell{StyleID: styleIDs[i], Value: v}
		} else {
			cells[i] = v
		}
	}

	return sw.SetRow(addr, cells)
}

func anyRowToStrings(row []any) []string {
	out := make([]string, len(row))
	for i, v := range row {
		if v == nil {
			continue
		}
		if s, ok := v.(string); ok {
			out[i] = s

			continue
		}
		out[i] = fmt.Sprint(v)
	}

	return out
}

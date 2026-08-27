package excelizex

import (
	"fmt"
	"sort"

	"github.com/cyclonevox/excelizex/layout"
	"github.com/xuri/excelize/v2"
)

func applyCascades(wb *Workbook, sheet string, lyt layout.Layout, headers []string, specs []cascadeSpec) error {
	if len(specs) == 0 {
		return nil
	}
	treeSheet := treeSheetName(sheet)
	idx, err := wb.f.GetSheetIndex(treeSheet)
	if err != nil {
		return fmt.Errorf("write: tree sheet index: %w", err)
	}
	if idx == -1 {
		if _, err := wb.f.NewSheet(treeSheet); err != nil {
			return fmt.Errorf("write: tree sheet: %w", err)
		}
	}
	dataStart := lyt.DataStartRow()
	treeCol := 1
	childSeen := make(map[string]struct{})
	for _, spec := range specs {
		if spec.child == spec.parent {
			return fmt.Errorf("write: cascade: child and parent are both %q", spec.child)
		}
		if _, dup := childSeen[spec.child]; dup {
			return fmt.Errorf("write: cascade: duplicate child header %q", spec.child)
		}
		childSeen[spec.child] = struct{}{}
		parentIdx := columnIndex(headers, spec.parent)
		if parentIdx < 0 {
			return fmt.Errorf("write: cascade: unknown parent header %q", spec.parent)
		}
		childIdx := columnIndex(headers, spec.child)
		if childIdx < 0 {
			return fmt.Errorf("write: cascade: unknown child header %q", spec.child)
		}
		parentCol, err := excelize.ColumnNumberToName(parentIdx + 1)
		if err != nil {
			return err
		}
		childCol, err := excelize.ColumnNumberToName(childIdx + 1)
		if err != nil {
			return err
		}
		keys := make([]string, 0, len(spec.tree))
		for k := range spec.tree {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, parentVal := range keys {
			children := spec.tree[parentVal]
			if parentVal == "" {
				return fmt.Errorf("write: cascade %q: empty parent value", spec.parent)
			}
			if len(children) == 0 {
				return fmt.Errorf("write: cascade %q: no options for %q", spec.child, parentVal)
			}
			treeColName, err := excelize.ColumnNumberToName(treeCol)
			if err != nil {
				return err
			}
			for i, child := range children {
				cell, err := excelize.JoinCellName(treeColName, i+1)
				if err != nil {
					return err
				}
				if err := wb.f.SetCellStr(treeSheet, cell, child); err != nil {
					return fmt.Errorf("write: cascade options: %w", err)
				}
			}
			refers := fmt.Sprintf("%s!$%s$1:$%s$%d", quoteSheetForFormula(treeSheet), treeColName, treeColName, len(children))
			if err := wb.f.SetDefinedName(&excelize.DefinedName{
				Name:     parentVal,
				RefersTo: refers,
				Scope:    sheet,
			}); err != nil {
				return fmt.Errorf("write: cascade defined name %q: %w", parentVal, err)
			}
			treeCol++
		}
		formula := fmt.Sprintf("INDIRECT($%s%d)", parentCol, dataStart)
		dv := excelize.NewDataValidation(true)
		dv.Sqref = fmt.Sprintf("%s%d:%s1048576", childCol, dataStart, childCol)
		dv.SetSqrefDropList(formula)
		dv.ShowInputMessage = true
		dv.ShowErrorMessage = true
		msg := "请先选择上级下拉，再选择本列"
		dv.Error = &msg
		if err := wb.f.AddDataValidation(sheet, dv); err != nil {
			return fmt.Errorf("write: cascade validation: %w", err)
		}
	}
	if err := wb.f.SetSheetVisible(treeSheet, false); err != nil {
		return fmt.Errorf("write: hide tree sheet: %w", err)
	}

	return nil
}

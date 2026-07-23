package excelizex_test

import (
	"bytes"
	"fmt"
	"testing"

	excelizex "github.com/cyclonevox/excelizex/v2"
	"github.com/xuri/excelize/v2"
)

// StudentRow is a typical import DTO；年级用 ExcelGrade 钩子转 A/B。
type StudentRow struct {
	Name  string `excel:"姓名" validate:"required"`
	Age   int    `excel:"年龄"`
	Grade int    `excel:"等级"`
}

func (r *StudentRow) ExcelGrade(raw string) error {
	switch raw {
	case "A":
		r.Grade = 1
		return nil
	case "B":
		r.Grade = 2
		return nil
	case "":
		r.Grade = 0
		return nil
	default:
		return fmt.Errorf("unknown grade %q", raw)
	}
}

func (r *StudentRow) ExcelExportGrade() (string, error) {
	switch r.Grade {
	case 1:
		return "A", nil
	case 2:
		return "B", nil
	case 0:
		return "", nil
	default:
		return "", fmt.Errorf("unknown grade %d", r.Grade)
	}
}

type simpleRow struct {
	Name string `excel:"姓名"`
	Age  int    `excel:"年龄"`
}

func buildNoticeGradeSheet(t *testing.T, dataRows [][]string) *bytes.Buffer {
	t.Helper()
	f := excelize.NewFile()
	sheet := "导入"
	_, _ = f.NewSheet(sheet)
	_ = f.DeleteSheet("Sheet1")
	_ = f.SetCellStr(sheet, "A1", "请按模板填写")
	_ = f.SetSheetRow(sheet, "A2", &[]string{"姓名", "年龄", "等级"})
	for i, row := range dataRows {
		addr, _ := excelize.JoinCellName("A", 3+i)
		_ = f.SetSheetRow(sheet, addr, &row)
	}
	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		t.Fatal(err)
	}

	return &buf
}

func openWorkbook(t *testing.T, buf *bytes.Buffer) *excelizex.Workbook {
	t.Helper()
	wb, err := excelizex.Open(buf)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = wb.Close() })

	return wb
}

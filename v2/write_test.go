package excelizex_test

import (
	"fmt"
	"testing"

	excelizex "github.com/cyclonevox/excelizex/v2"
	"github.com/cyclonevox/excelizex/v2/layout"
	"github.com/cyclonevox/excelizex/v2/style"
	"github.com/xuri/excelize/v2"
)

type gradeLabel int

func (g gradeLabel) MarshalText() ([]byte, error) {
	return []byte(fmt.Sprintf("G%d", g)), nil
}

func TestRegisterStyle(t *testing.T) {
	wb := excelizex.New()
	t.Cleanup(func() { _ = wb.Close() })

	accent := style.New("accent", &excelize.Style{
		Font: &excelize.Font{Bold: true, Color: "#00AA00"},
	})
	if err := wb.RegisterStyle(accent); err != nil {
		t.Fatal(err)
	}

	type styledRow struct {
		Name string `excel:"姓名" style:"accent,body"`
		Age  int    `excel:"年龄" style:"header,body"`
	}

	if err := excelizex.Write[styledRow](wb.Sheet("样式").WithLayout(layout.HeaderData{})).
		Rows(styledRow{Name: "测试", Age: 30}).
		Apply(); err != nil {
		t.Fatal(err)
	}
}

func TestTextMarshalerWriteBuilder(t *testing.T) {
	type gradeRow struct {
		Grade gradeLabel `excel:"等级"`
	}

	wb := excelizex.New()
	t.Cleanup(func() { _ = wb.Close() })

	if err := excelizex.Write[gradeRow](wb.Sheet("导出").WithLayout(layout.HeaderData{})).
		Rows(gradeRow{Grade: 2}).Apply(); err != nil {
		t.Fatal(err)
	}

	val, err := wb.File().GetCellValue("导出", "A2")
	if err != nil {
		t.Fatal(err)
	}
	if val != "G2" {
		t.Fatalf("cell: %q want G2", val)
	}
}

func TestNoticeFromRowField(t *testing.T) {
	type noticeRow struct {
		Notice string `excel:"notice"`
		Name   string `excel:"姓名"`
	}

	wb := excelizex.New()
	t.Cleanup(func() { _ = wb.Close() })

	if err := excelizex.Write[noticeRow](wb.Sheet("导入").WithLayout(layout.NoticeHeaderData{})).
		Rows(noticeRow{Notice: "从行字段生成提示", Name: "张三"}).
		Apply(); err != nil {
		t.Fatal(err)
	}

	val, err := wb.File().GetCellValue("导入", "A1")
	if err != nil {
		t.Fatal(err)
	}
	if val != "从行字段生成提示" {
		t.Fatalf("notice cell: %q", val)
	}
}

func TestApplyReplacesPreviousDataRows(t *testing.T) {
	wb := excelizex.New()
	t.Cleanup(func() { _ = wb.Close() })
	sheet := wb.Sheet("导出").WithLayout(layout.HeaderData{})

	many := make([]simpleRow, 100)
	for i := range many {
		many[i] = simpleRow{Name: fmt.Sprintf("旧%d", i), Age: i}
	}
	if err := excelizex.Write[simpleRow](sheet).Rows(many...).Apply(); err != nil {
		t.Fatal(err)
	}
	few := []simpleRow{{Name: "新0", Age: 1}, {Name: "新1", Age: 2}}
	for i := 0; i < 8; i++ {
		few = append(few, simpleRow{Name: fmt.Sprintf("新%d", i+2), Age: i + 3})
	}
	if err := excelizex.Write[simpleRow](sheet).Rows(few...).Apply(); err != nil {
		t.Fatal(err)
	}

	rows, err := wb.File().GetRows("导出")
	if err != nil {
		t.Fatal(err)
	}
	// header + 10 data rows
	if len(rows) != 11 {
		t.Fatalf("rows after rewrite = %d, want 11 (no leftover from 100-row write)", len(rows))
	}
	if rows[1][0] != "新0" {
		t.Fatalf("first data row = %v", rows[1])
	}
}

func TestTemplateApplyClearsOldData(t *testing.T) {
	wb := excelizex.New()
	t.Cleanup(func() { _ = wb.Close() })
	sheet := wb.Sheet("模板").WithLayout(layout.HeaderData{})
	if err := excelizex.Write[simpleRow](sheet).
		Rows(simpleRow{Name: "残留", Age: 9}, simpleRow{Name: "也残留", Age: 8}).
		Apply(); err != nil {
		t.Fatal(err)
	}
	if err := excelizex.Write[simpleRow](sheet).Template().Apply(); err != nil {
		t.Fatal(err)
	}
	rows, err := wb.File().GetRows("模板")
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 {
		t.Fatalf("template rewrite left %d rows, want header only", len(rows))
	}
	if rows[0][0] != "姓名" {
		t.Fatalf("header = %v", rows[0])
	}
}

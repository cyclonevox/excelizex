package excelizex_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	excelizex "github.com/cyclonevox/excelizex"
	"github.com/cyclonevox/excelizex/layout"
	"github.com/cyclonevox/excelizex/style"
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

type scoreExtendRow struct {
	Name  string            `excel:"姓名"`
	Extra map[string]string `excel:",extend"`
}

func TestWriteReadExtendColumns(t *testing.T) {
	wb := excelizex.New()
	t.Cleanup(func() { _ = wb.Close() })
	if err := excelizex.Write[scoreExtendRow](wb.Sheet("成绩").WithLayout(layout.HeaderData{})).
		Extend("语文", "数学").
		Rows(
			scoreExtendRow{Name: "张三", Extra: map[string]string{"语文": "90", "数学": "80"}},
			scoreExtendRow{Name: "李四", Extra: map[string]string{"语文": "70", "数学": "88"}},
		).
		Apply(); err != nil {
		t.Fatal(err)
	}
	rows, err := wb.File().GetRows("成绩")
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 3 || rows[0][1] != "语文" || rows[0][2] != "数学" {
		t.Fatalf("headers: %v", rows)
	}
	got, res, err := excelizex.Read[scoreExtendRow](wb.Sheet("成绩").WithLayout(layout.HeaderData{})).
		Collect(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if res.HasErrors() {
		t.Fatalf("errors: %v", res.Errors())
	}
	if len(got) != 2 || got[0].Extra["语文"] != "90" || got[1].Extra["数学"] != "88" {
		t.Fatalf("rows: %+v", got)
	}
}

func TestWriteStreamRoundTrip(t *testing.T) {
	wb := excelizex.New()
	t.Cleanup(func() { _ = wb.Close() })
	const n = 200
	if err := excelizex.Write[simpleRow](wb.Sheet("导出").WithLayout(layout.HeaderData{})).
		Stream(func(yield func(simpleRow) error) error {
			for i := 0; i < n; i++ {
				if err := yield(simpleRow{Name: fmt.Sprintf("n%d", i), Age: i}); err != nil {
					return err
				}
			}

			return nil
		}).
		Apply(); err != nil {
		t.Fatal(err)
	}
	got, res, err := excelizex.Read[simpleRow](wb.Sheet("导出").WithLayout(layout.HeaderData{})).
		Collect(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if res.HasErrors() {
		t.Fatalf("errors: %v", res.Errors())
	}
	if len(got) != n || got[0].Name != "n0" || got[n-1].Age != n-1 {
		t.Fatalf("rows: len=%d first=%+v last=%+v", len(got), got[0], got[n-1])
	}
}

func TestWriteStreamDropdown(t *testing.T) {
	wb := excelizex.New()
	t.Cleanup(func() { _ = wb.Close() })
	if err := excelizex.Write[simpleRow](wb.Sheet("导出").WithLayout(layout.HeaderData{})).
		Dropdown("姓名", []string{"张三", "李四"}).
		Stream(func(yield func(simpleRow) error) error {
			return yield(simpleRow{Name: "张三", Age: 18})
		}).
		Apply(); err != nil {
		t.Fatal(err)
	}
	dvs, err := wb.File().GetDataValidations("导出")
	if err != nil {
		t.Fatal(err)
	}
	if len(dvs) == 0 {
		t.Fatal("expected dropdown after stream write")
	}
}

func TestWriteGroupedHeaders(t *testing.T) {
	type row struct {
		Name    string `excel:"姓名"`
		Age     int    `excel:"年龄"`
		Math    int    `excel:"数学"`
		Chinese int    `excel:"语文"`
	}
	lyt := layout.Grouped{
		Groups: []layout.Group{
			{Title: "基本信息", Span: 2},
			{Title: "成绩", Span: 2},
		},
	}
	wb := excelizex.New()
	t.Cleanup(func() { _ = wb.Close() })
	if err := excelizex.Write[row](wb.Sheet("分组").WithLayout(lyt)).
		Rows(row{Name: "张三", Age: 18, Math: 90, Chinese: 88}).
		Apply(); err != nil {
		t.Fatal(err)
	}
	merged, err := wb.File().GetMergeCells("分组")
	if err != nil {
		t.Fatal(err)
	}
	if len(merged) < 2 {
		t.Fatalf("expected group merges, got %v", merged)
	}
	got, res, err := excelizex.Read[row](wb.Sheet("分组").WithLayout(lyt)).
		Collect(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if res.HasErrors() || len(got) != 1 || got[0].Chinese != 88 {
		t.Fatalf("rows=%+v errors=%v", got, res.Errors())
	}
}

func TestWriteCascadeDropdown(t *testing.T) {
	type row struct {
		Name     string `excel:"姓名"`
		Province string `excel:"省"`
		City     string `excel:"市"`
	}
	wb := excelizex.New()
	t.Cleanup(func() { _ = wb.Close() })
	if err := excelizex.Write[row](wb.Sheet("地区").WithLayout(layout.HeaderData{})).
		Cascade("市", "省", map[string][]string{
			"浙江": {"杭州", "宁波"},
			"江苏": {"南京", "苏州"},
		}).
		Template().
		Apply(); err != nil {
		t.Fatal(err)
	}
	dvs, err := wb.File().GetDataValidations("地区")
	if err != nil {
		t.Fatal(err)
	}
	if len(dvs) < 2 {
		t.Fatalf("expected parent and child validations, got %d", len(dvs))
	}
	foundIndirect := false
	for _, dv := range dvs {
		if strings.Contains(dv.Formula1, "INDIRECT") {
			foundIndirect = true
			break
		}
	}
	if !foundIndirect {
		t.Fatalf("expected INDIRECT validation, got %+v", dvs)
	}
	names := wb.File().GetDefinedName()
	var gotNames []string
	for _, n := range names {
		gotNames = append(gotNames, n.Name)
	}
	if !containsName(gotNames, "浙江") || !containsName(gotNames, "江苏") {
		t.Fatalf("defined names: %v", gotNames)
	}
}

func TestWriteStreamRejectsRows(t *testing.T) {
	wb := excelizex.New()
	t.Cleanup(func() { _ = wb.Close() })
	err := excelizex.Write[simpleRow](wb.Sheet("导出").WithLayout(layout.HeaderData{})).
		Rows(simpleRow{Name: "x", Age: 1}).
		Stream(func(yield func(simpleRow) error) error { return yield(simpleRow{Name: "y", Age: 2}) }).
		Apply()
	if err == nil {
		t.Fatal("expected Rows+Stream error")
	}
}

func TestWriteExtendRequiresMap(t *testing.T) {
	wb := excelizex.New()
	t.Cleanup(func() { _ = wb.Close() })
	err := excelizex.Write[simpleRow](wb.Sheet("导出").WithLayout(layout.HeaderData{})).
		Extend("语文").
		Rows(simpleRow{Name: "张三", Age: 18}).
		Apply()
	if err == nil {
		t.Fatal("expected Extend without extend field to fail")
	}
}

func containsName(names []string, want string) bool {
	for _, n := range names {
		if n == want {
			return true
		}
	}

	return false
}

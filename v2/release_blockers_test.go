package excelizex_test

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	excelizex "github.com/cyclonevox/excelizex/v2"
	"github.com/cyclonevox/excelizex/v2/layout"
	"github.com/cyclonevox/excelizex/v2/schema"
	"github.com/cyclonevox/excelizex/v2/style"
	"github.com/xuri/excelize/v2"
)

type blockerRow struct {
	Name string `excel:"姓名"`
	Age  int    `excel:"年龄"`
}

func TestEachDefaultConcurrencyIsOne(t *testing.T) {
	wb := excelizex.New()
	t.Cleanup(func() { _ = wb.Close() })
	rows := make([]blockerRow, 8)
	for i := range rows {
		rows[i] = blockerRow{Name: fmt.Sprintf("u%d", i), Age: i}
	}
	if err := excelizex.Write[blockerRow](wb.Sheet("导入").WithLayout(layout.HeaderData{})).
		Rows(rows...).Apply(); err != nil {
		t.Fatal(err)
	}

	var current, maxConcurrent atomic.Int32
	_, err := excelizex.Read[blockerRow](wb.Sheet("导入").WithLayout(layout.HeaderData{})).
		Each(context.Background(), func(_ excelizex.Context, _ blockerRow) error {
			c := current.Add(1)
			for {
				old := maxConcurrent.Load()
				if c <= old || maxConcurrent.CompareAndSwap(old, c) {
					break
				}
			}
			time.Sleep(25 * time.Millisecond)
			current.Add(-1)

			return nil
		})
	if err != nil {
		t.Fatal(err)
	}
	if maxConcurrent.Load() != 1 {
		t.Fatalf("max concurrent workers = %d, want 1", maxConcurrent.Load())
	}
}

func TestConcurrencyOptionOverridesSetConcurrency(t *testing.T) {
	wb := excelizex.New()
	t.Cleanup(func() { _ = wb.Close() })
	rows := make([]blockerRow, 8)
	for i := range rows {
		rows[i] = blockerRow{Name: fmt.Sprintf("u%d", i), Age: i}
	}
	if err := excelizex.Write[blockerRow](wb.Sheet("导入").WithLayout(layout.HeaderData{})).
		Rows(rows...).Apply(); err != nil {
		t.Fatal(err)
	}

	var current, maxConcurrent atomic.Int32
	_, err := excelizex.Read[blockerRow](wb.Sheet("导入").WithLayout(layout.HeaderData{})).
		SetConcurrency(8).
		Each(context.Background(), func(_ excelizex.Context, _ blockerRow) error {
			c := current.Add(1)
			for {
				old := maxConcurrent.Load()
				if c <= old || maxConcurrent.CompareAndSwap(old, c) {
					break
				}
			}
			time.Sleep(25 * time.Millisecond)
			current.Add(-1)

			return nil
		}, excelizex.Concurrency(1))
	if err != nil {
		t.Fatal(err)
	}
	if maxConcurrent.Load() != 1 {
		t.Fatalf("max concurrent workers = %d, want 1 after Concurrency(1)", maxConcurrent.Load())
	}
}

func TestOpenWriteNewSheetKeepsBusinessSheet1(t *testing.T) {
	src := excelize.NewFile()
	t.Cleanup(func() { _ = src.Close() })
	if err := src.SetCellStr("Sheet1", "A1", "业务数据不可删"); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := src.Write(&buf); err != nil {
		t.Fatal(err)
	}

	wb, err := excelizex.Open(&buf)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = wb.Close() })

	if err := excelizex.Write[blockerRow](wb.Sheet("新表").WithLayout(layout.HeaderData{})).
		Rows(blockerRow{Name: "张三", Age: 18}).Apply(); err != nil {
		t.Fatal(err)
	}

	idx, err := wb.File().GetSheetIndex("Sheet1")
	if err != nil {
		t.Fatal(err)
	}
	if idx == -1 {
		t.Fatal("Sheet1 was deleted after writing another sheet on Open() workbook")
	}
	got, err := wb.File().GetCellValue("Sheet1", "A1")
	if err != nil {
		t.Fatal(err)
	}
	if got != "业务数据不可删" {
		t.Fatalf("Sheet1 content = %q", got)
	}
}

func TestApplyReplacesPreviousDataRows(t *testing.T) {
	wb := excelizex.New()
	t.Cleanup(func() { _ = wb.Close() })
	sheet := wb.Sheet("导出").WithLayout(layout.HeaderData{})

	many := make([]blockerRow, 100)
	for i := range many {
		many[i] = blockerRow{Name: fmt.Sprintf("旧%d", i), Age: i}
	}
	if err := excelizex.Write[blockerRow](sheet).Rows(many...).Apply(); err != nil {
		t.Fatal(err)
	}
	few := []blockerRow{{Name: "新0", Age: 1}, {Name: "新1", Age: 2}}
	for i := 0; i < 8; i++ {
		few = append(few, blockerRow{Name: fmt.Sprintf("新%d", i+2), Age: i + 3})
	}
	if err := excelizex.Write[blockerRow](sheet).Rows(few...).Apply(); err != nil {
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
	if err := excelizex.Write[blockerRow](sheet).
		Rows(blockerRow{Name: "残留", Age: 9}, blockerRow{Name: "也残留", Age: 8}).
		Apply(); err != nil {
		t.Fatal(err)
	}
	if err := excelizex.Write[blockerRow](sheet).Template().Apply(); err != nil {
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

func TestSchemaPointerInlineNoPanic(t *testing.T) {
	type Address struct {
		City   string `excel:"城市"`
		Street string `excel:"街道"`
	}
	type Row struct {
		Name string   `excel:"姓名"`
		Addr *Address `excel:",inline"`
	}
	sc, err := schema.New(Row{})
	if err != nil {
		t.Fatal(err)
	}
	if len(sc.Columns) != 3 {
		t.Fatalf("columns = %d want 3", len(sc.Columns))
	}
	if _, err := schema.FromType(nil); err == nil {
		t.Fatal("expected error for nil type")
	}
	if _, err := schema.New(42); err == nil {
		t.Fatal("expected error for non-struct")
	}
}

func TestStyleBodyBluePlusLockedKeepsFillAndProtection(t *testing.T) {
	f := excelize.NewFile()
	t.Cleanup(func() { _ = f.Close() })
	reg := style.NewRegistry(f)
	if err := reg.RegisterDefaults(); err != nil {
		t.Fatal(err)
	}
	id, err := reg.Resolve("body-blue", "locked")
	if err != nil {
		t.Fatal(err)
	}
	if err := f.SetCellStyle("Sheet1", "A1", "A1", id); err != nil {
		t.Fatal(err)
	}
	styleID, err := f.GetCellStyle("Sheet1", "A1")
	if err != nil {
		t.Fatal(err)
	}
	st, err := f.GetStyle(styleID)
	if err != nil {
		t.Fatal(err)
	}
	if st.Fill.Type == "" || len(st.Fill.Color) == 0 || st.Fill.Color[0] == "" {
		t.Fatalf("expected body-blue fill retained, got %+v", st.Fill)
	}
	if st.NumFmt != 49 {
		t.Fatalf("expected NumFmt 49, got %d", st.NumFmt)
	}
	if st.Protection == nil || !st.Protection.Locked {
		t.Fatalf("expected locked protection, got %+v", st.Protection)
	}
}

func TestWriteErrorsOrdersConcurrentFailuresByRow(t *testing.T) {
	wb := excelizex.New()
	t.Cleanup(func() { _ = wb.Close() })
	rows := make([]blockerRow, 20)
	for i := range rows {
		rows[i] = blockerRow{Name: fmt.Sprintf("u%d", i), Age: i}
	}
	if err := excelizex.Write[blockerRow](wb.Sheet("导入").WithLayout(layout.HeaderData{})).
		Rows(rows...).Apply(); err != nil {
		t.Fatal(err)
	}

	res, err := excelizex.Read[blockerRow](wb.Sheet("导入").WithLayout(layout.HeaderData{})).
		Each(context.Background(), func(ctx excelizex.Context, row blockerRow) error {
			if row.Age%2 == 0 {
				return fmt.Errorf("fail age %d", row.Age)
			}

			return nil
		}, excelizex.Concurrency(8))
	if err != nil {
		t.Fatal(err)
	}
	errs := res.Errors()
	if len(errs) == 0 {
		t.Fatal("expected row errors")
	}
	for i := 1; i < len(errs); i++ {
		if errs[i].Row < errs[i-1].Row {
			t.Fatalf("errors not sorted by row: %d then %d", errs[i-1].Row, errs[i].Row)
		}
	}
	if err := wb.WriteErrors(res); err != nil {
		t.Fatal(err)
	}
	outRows, err := wb.File().GetRows("导入")
	if err != nil {
		t.Fatal(err)
	}
	prevAge := -1
	for i := 1; i < len(outRows); i++ {
		age := mustAtoi(t, outRows[i][1])
		if age < prevAge {
			t.Fatalf("WriteErrors rows out of order: age %d after %d", age, prevAge)
		}
		prevAge = age
		if !strings.Contains(outRows[i][len(outRows[i])-1], "fail age") {
			t.Fatalf("missing error message on row %v", outRows[i])
		}
	}
}

func mustAtoi(t *testing.T, s string) int {
	t.Helper()
	var n int
	if _, err := fmt.Sscanf(s, "%d", &n); err != nil {
		t.Fatalf("atoi %q: %v", s, err)
	}

	return n
}

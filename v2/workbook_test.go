package excelizex_test

import (
	"bytes"
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	excelizex "github.com/cyclonevox/excelizex/v2"
	"github.com/cyclonevox/excelizex/v2/layout"
	"github.com/xuri/excelize/v2"
)

type regStressImportRow struct {
	Name  string `excel:"姓名"`
	Age   int    `excel:"年龄"`
	Grade int    `excel:"等级"`
}

func (r *regStressImportRow) ExcelGrade(raw string) error {
	switch raw {
	case "A":
		r.Grade = 1
		return nil
	case "B":
		r.Grade = 2
		return nil
	default:
		return fmt.Errorf("unknown grade %q", raw)
	}
}

type regStressScoreRow struct {
	Name  string `excel:"姓名"`
	Score int    `excel:"分数"`
}

func buildStressImportXLSX(t *testing.T, n int) *bytes.Buffer {
	t.Helper()
	f := excelize.NewFile()
	sheet := "导入"
	_, _ = f.NewSheet(sheet)
	_ = f.DeleteSheet("Sheet1")
	_ = f.SetCellStr(sheet, "A1", "压测")
	_ = f.SetSheetRow(sheet, "A2", &[]string{"姓名", "年龄", "等级"})
	for i := 0; i < n; i++ {
		addr, _ := excelize.JoinCellName("A", 3+i)
		_ = f.SetSheetRow(sheet, addr, &[]string{fmt.Sprintf("用户%d", i+1), "25", "A"})
	}
	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		t.Fatal(err)
	}

	return &buf
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

	if err := excelizex.Write[simpleRow](wb.Sheet("新表").WithLayout(layout.HeaderData{})).
		Rows(simpleRow{Name: "张三", Age: 18}).Apply(); err != nil {
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

// Each 与 Save/Apply 并发压测：同一 Workbook 上 mutex 串行化 excelize 访问。
func TestConcurrentEachAndSave(t *testing.T) {
	const n = 200
	buf := buildStressImportXLSX(t, n)
	wb, err := excelizex.Open(buf)
	if err != nil {
		t.Fatal(err)
	}

	done := make(chan struct{})
	var stressErr atomic.Value
	go func() {
		defer close(done)
		var out bytes.Buffer
		for i := 0; i < 30; i++ {
			out.Reset()
			if err := wb.Save(&out); err != nil {
				stressErr.Store(err)
				return
			}
			if err := excelizex.Write[regStressScoreRow](wb.Sheet("导出").WithLayout(layout.HeaderData{})).
				Rows(regStressScoreRow{Name: fmt.Sprintf("行%d", i), Score: i}).
				Apply(); err != nil {
				stressErr.Store(err)
				return
			}
			time.Sleep(time.Millisecond)
		}
	}()

	_, err = excelizex.Read[regStressImportRow](wb.Sheet("导入")).
		Each(context.Background(), func(ctx excelizex.Context, row regStressImportRow) error {
			time.Sleep(time.Microsecond)

			return nil
		}, excelizex.Concurrency(8))
	if err != nil {
		t.Fatal(err)
	}
	<-done
	if v := stressErr.Load(); v != nil {
		t.Fatalf("stress: %v", v)
	}
	_ = wb.Close()
}

func TestWithRowNilContext(t *testing.T) {
	c := excelizex.WithRow(nil, 5)
	if c.Row != 5 || c.Context == nil {
		t.Fatalf("context: row=%d ctx=%v", c.Row, c.Context)
	}
}

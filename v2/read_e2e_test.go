package excelizex_test

import (
	"bytes"
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	excelizex "github.com/cyclonevox/excelizex/v2"
	"github.com/cyclonevox/excelizex/v2/convert"
	"github.com/cyclonevox/excelizex/v2/layout"
	"github.com/cyclonevox/excelizex/v2/validate"
	"github.com/xuri/excelize/v2"
)

type studentRow struct {
	Name   string    `excel:"姓名" validate:"required"`
	Age    int       `excel:"年龄"`
	Active bool      `excel:"启用"`
	Joined time.Time `excel:"入学日期" time:"2006-01-02"`
	Grade  int       `excel:"等级" conv:"grade"`
}

func buildStudentSheet(t *testing.T, layoutKind string) *bytes.Buffer {
	t.Helper()
	f := excelize.NewFile()
	sheet := "Sheet1"
	if layoutKind == "notice" {
		_ = f.SetCellStr(sheet, "A1", "导入说明")
		_ = f.SetSheetRow(sheet, "A2", &[]string{"姓名", "年龄", "启用", "入学日期", "等级"})
		_ = f.SetSheetRow(sheet, "A3", &[]string{"张三", "18", "是", "2024-09-01", "A"})
		_ = f.SetSheetRow(sheet, "A4", &[]string{"李四", "bad", "否", "2024-09-02", "B"})
	} else {
		_ = f.SetSheetRow(sheet, "A1", &[]string{"姓名", "年龄", "启用", "入学日期", "等级"})
		_ = f.SetSheetRow(sheet, "A2", &[]string{"王五", "20", "1", "2024-10-01", "A"})
	}
	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		t.Fatal(err)
	}

	return &buf
}

func TestReadCollectE2E(t *testing.T) {
	buf := buildStudentSheet(t, "notice")
	wb, err := excelizex.Open(buf)
	if err != nil {
		t.Fatal(err)
	}
	defer wb.Close()

	rb := excelizex.Read[studentRow](wb.Sheet("Sheet1").WithLayout(layout.NoticeHeaderData{})).
		Convert("grade", func(raw string) (any, error) {
			switch raw {
			case "A":
				return 1, nil
			case "B":
				return 2, nil
			default:
				return 0, fmt.Errorf("unknown grade %q", raw)
			}
		}).
		Validate(validate.Required{})

	rows, res, err := rb.Collect(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 {
		t.Fatalf("ok rows: got %d want 1; errors=%v", len(rows), res.Errors())
	}
	if rows[0].Name != "张三" || rows[0].Age != 18 || !rows[0].Active || rows[0].Grade != 1 {
		t.Fatalf("row: %+v", rows[0])
	}
	if len(res.Errors()) != 1 {
		t.Fatalf("errors: %v", res.Errors())
	}
}

func TestReadHeaderDataLayout(t *testing.T) {
	buf := buildStudentSheet(t, "header")
	wb, err := excelizex.Open(buf)
	if err != nil {
		t.Fatal(err)
	}
	defer wb.Close()

	rows, _, err := excelizex.Read[studentRow](wb.Sheet("Sheet1").WithLayout(layout.HeaderData{})).
		Convert("grade", func(raw string) (any, error) { return 1, nil }).
		Collect(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].Name != "王五" {
		t.Fatalf("rows: %+v", rows)
	}
}

func TestEachConcurrentRace(t *testing.T) {
	buf := buildStudentSheet(t, "notice")
	wb, err := excelizex.Open(buf)
	if err != nil {
		t.Fatal(err)
	}
	defer wb.Close()

	var count atomic.Int32
	_, err = excelizex.Read[studentRow](wb.Sheet("Sheet1")).
		Convert("grade", func(raw string) (any, error) { return 1, nil }).
		SetConcurrency(4).
		Each(context.Background(), func(ctx excelizex.Context, row studentRow) error {
			if row.Name == "张三" {
				count.Add(1)
			}
			time.Sleep(5 * time.Millisecond)

			return nil
		}, excelizex.Concurrency(4))
	if err != nil {
		t.Fatal(err)
	}
	if count.Load() != 1 {
		t.Fatalf("processed: %d", count.Load())
	}
}

func TestWriteErrorsRoundtrip(t *testing.T) {
	f := excelize.NewFile()
	sheet := "Sheet1"
	_ = f.SetCellStr(sheet, "A1", "导入说明")
	_ = f.SetSheetRow(sheet, "A2", &[]string{"姓名", "年龄", "启用", "入学日期", "等级"})
	_ = f.SetSheetRow(sheet, "A3", &[]string{"张三", "18", "是", "2024-09-01", "A"})
	_ = f.SetSheetRow(sheet, "A4", &[]string{"李四", "19", "否", "2024-09-02", "B"})
	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		t.Fatal(err)
	}

	wb, err := excelizex.Open(&buf)
	if err != nil {
		t.Fatal(err)
	}

	rb := excelizex.Read[studentRow](wb.Sheet("Sheet1")).
		Convert("grade", func(raw string) (any, error) {
			if raw == "A" {
				return 1, nil
			}

			return 0, fmt.Errorf("bad grade")
		})
	_, res, err := rb.Collect(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !res.HasErrors() {
		t.Fatal("expected errors")
	}
	if err := wb.WriteErrors(res); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err := wb.Save(&out); err != nil {
		t.Fatal(err)
	}
	wb2, err := excelizex.Open(&out)
	if err != nil {
		t.Fatal(err)
	}
	defer wb2.Close()
	rows, res2, err := excelizex.Read[studentRow](wb2.Sheet("Sheet1")).
		Convert("grade", func(raw string) (any, error) { return 1, nil }).
		Collect(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 {
		t.Fatalf("after write errors ok rows: %d errs=%v", len(rows), res2.Errors())
	}
	if len(res2.Errors()) != 0 {
		t.Fatalf("unexpected errors: %v", res2.Errors())
	}
}

func TestConvertToHelper(t *testing.T) {
	reg := convert.Registry{}
	convert.ConvertTo(reg, "x", func(s string) (int, error) { return len(s), nil })
	if len(reg) != 1 {
		t.Fatal("registry empty")
	}
}

func TestMissingColumnUpfront(t *testing.T) {
	f := excelize.NewFile()
	_ = f.SetCellStr("Sheet1", "A1", "notice")
	_ = f.SetSheetRow("Sheet1", "A2", &[]string{"姓名"})
	var buf bytes.Buffer
	_ = f.Write(&buf)
	wb, _ := excelizex.Open(&buf)
	_, _, err := excelizex.Read[studentRow](wb.Sheet("Sheet1")).Collect(context.Background())
	if err == nil {
		t.Fatal("expected schema match error")
	}
}

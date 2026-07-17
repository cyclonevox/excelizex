package excelizex_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"

	excelizex "github.com/cyclonevox/excelizex/v2"
	"github.com/cyclonevox/excelizex/v2/layout"
	"github.com/cyclonevox/excelizex/v2/schema"
	"github.com/cyclonevox/excelizex/v2/style"
	"github.com/xuri/excelize/v2"
)

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

func TestEachMap(t *testing.T) {
	t.Run("maps rows to callback type", func(t *testing.T) {
		buf := buildNoticeGradeSheet(t, [][]string{
			{"张三", "18", "A"},
			{"李四", "20", "B"},
		})
		wb := openWorkbook(t, buf)

		type cmd struct {
			Label string
		}
		var (
			seen []string
			mu   sync.Mutex
		)
		_, err := excelizex.EachMap(
			excelizex.Read[StudentRow](wb.Sheet("导入").WithLayout(layout.NoticeHeaderData{})).
				Convert("grade", exampleGradeImport),
			context.Background(),
			func(row StudentRow) (cmd, error) {
				return cmd{Label: row.Name + "-ok"}, nil
			},
			func(_ excelizex.Context, c cmd) error {
				mu.Lock()
				seen = append(seen, c.Label)
				mu.Unlock()

				return nil
			},
		)
		if err != nil {
			t.Fatal(err)
		}
		if len(seen) != 2 {
			t.Fatalf("seen: %v", seen)
		}
		want := map[string]bool{"张三-ok": true, "李四-ok": true}
		for _, s := range seen {
			if !want[s] {
				t.Fatalf("unexpected label %q in %v", s, seen)
			}
		}
	})

	t.Run("mapFn error recorded", func(t *testing.T) {
		buf := buildNoticeGradeSheet(t, [][]string{{"bad", "18", "A"}})
		wb := openWorkbook(t, buf)

		type cmd struct{}
		res, err := excelizex.EachMap(
			excelizex.Read[StudentRow](wb.Sheet("导入").WithLayout(layout.NoticeHeaderData{})).
				Convert("grade", exampleGradeImport),
			context.Background(),
			func(row StudentRow) (cmd, error) {
				if row.Name == "bad" {
					return cmd{}, errors.New("map rejected")
				}

				return cmd{}, nil
			},
			func(_ excelizex.Context, _ cmd) error { return nil },
		)
		if err != nil {
			t.Fatal(err)
		}
		if !res.HasErrors() {
			t.Fatal("expected map error in result")
		}
		if res.Errors()[0].Messages[0] != "map rejected" {
			t.Fatalf("errors: %v", res.Errors())
		}
	})
}

func TestConvertToRead(t *testing.T) {
	type twiceRow struct {
		N int `excel:"数值" conv:"twice"`
	}

	f := excelize.NewFile()
	sheet := "数据"
	_, _ = f.NewSheet(sheet)
	_ = f.DeleteSheet("Sheet1")
	_ = f.SetSheetRow(sheet, "A1", &[]string{"数值"})
	_ = f.SetSheetRow(sheet, "A2", &[]string{"21"})
	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		t.Fatal(err)
	}
	wb := openWorkbook(t, &buf)

	rows, _, err := excelizex.ConvertTo(
		excelizex.Read[twiceRow](wb.Sheet(sheet).WithLayout(layout.HeaderData{})),
		"twice",
		func(raw string) (int, error) {
			n, err := strconv.Atoi(raw)
			if err != nil {
				return 0, err
			}

			return n * 2, nil
		},
	).Collect(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].N != 42 {
		t.Fatalf("rows: %+v", rows)
	}
}

func TestSetConcurrency(t *testing.T) {
	const n = 6
	data := make([][]string, n)
	for i := 0; i < n; i++ {
		data[i] = []string{fmt.Sprintf("用户%d", i), "20", "A"}
	}
	buf := buildNoticeGradeSheet(t, data)
	wb := openWorkbook(t, buf)

	var count atomic.Int32
	_, err := excelizex.Read[StudentRow](wb.Sheet("导入").WithLayout(layout.NoticeHeaderData{})).
		Convert("grade", exampleGradeImport).
		SetConcurrency(2).
		Each(context.Background(), func(_ excelizex.Context, _ StudentRow) error {
			count.Add(1)

			return nil
		})
	if err != nil {
		t.Fatal(err)
	}
	if count.Load() != n {
		t.Fatalf("processed %d want %d", count.Load(), n)
	}
}

func TestFailFastOption(t *testing.T) {
	buf := buildNoticeGradeSheet(t, [][]string{
		{"张三", "18", "A"},
		{"李四", "bad-age", "B"},
	})
	wb := openWorkbook(t, buf)

	_, err := excelizex.Read[StudentRow](wb.Sheet("导入").WithLayout(layout.NoticeHeaderData{})).
		Convert("grade", exampleGradeImport).
		Each(context.Background(), func(_ excelizex.Context, _ StudentRow) error {
			return nil
		}, excelizex.FailFast())
	if err == nil {
		t.Fatal("expected fail-fast bind error")
	}
}

func TestWithSchema(t *testing.T) {
	sc, err := schema.New(StudentRow{})
	if err != nil {
		t.Fatal(err)
	}

	wb := excelizex.New()
	t.Cleanup(func() { _ = wb.Close() })

	if err := excelizex.Write[StudentRow](wb.Sheet("导入").
		WithLayout(layout.NoticeHeaderData{}).
		WithSchema(sc).
		WithNotice("请按模板填写")).
		Convert("grade", exampleGradeExport).
		Rows(StudentRow{Name: "王五", Age: 22, Grade: 1}).
		Apply(); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	if err := wb.Save(&buf); err != nil {
		t.Fatal(err)
	}
	wb2 := openWorkbook(t, &buf)

	rows, res, err := excelizex.Read[StudentRow](wb2.Sheet("导入").
		WithLayout(layout.NoticeHeaderData{}).
		WithSchema(sc)).
		Convert("grade", exampleGradeImport).
		Collect(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if res.HasErrors() {
		t.Fatalf("errors: %v", res.Errors())
	}
	if len(rows) != 1 || rows[0].Name != "王五" || rows[0].Grade != 1 {
		t.Fatalf("rows: %+v", rows)
	}
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

func TestExportToWriteBuilder(t *testing.T) {
	type gradeRow struct {
		Grade int `excel:"等级" conv:"grade"`
	}

	wb := excelizex.New()
	t.Cleanup(func() { _ = wb.Close() })

	if err := excelizex.ExportTo(
		excelizex.Write[gradeRow](wb.Sheet("导出").WithLayout(layout.HeaderData{})),
		"grade",
		func(g int) (string, error) {
			return fmt.Sprintf("G%d", g), nil
		},
	).Rows(gradeRow{Grade: 2}).Apply(); err != nil {
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

func TestResultHelpers(t *testing.T) {
	buf := buildNoticeGradeSheet(t, [][]string{{"张三", "18", "A"}})
	wb := openWorkbook(t, buf)

	_, res, err := excelizex.Read[StudentRow](wb.Sheet("导入").WithLayout(layout.NoticeHeaderData{})).
		Convert("grade", exampleGradeImport).
		Collect(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if res.DataStartRow() != 3 {
		t.Fatalf("DataStartRow: %d want 3", res.DataStartRow())
	}
	if res.HeaderRow() != 2 {
		t.Fatalf("HeaderRow: %d want 2", res.HeaderRow())
	}
	if excelizex.StringRow(3) != "3" {
		t.Fatalf("StringRow: %q", excelizex.StringRow(3))
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

func TestWriteErrorsPadRow(t *testing.T) {
	wb := excelizex.New()
	t.Cleanup(func() { _ = wb.Close() })

	if err := excelizex.Write[StudentRow](wb.Sheet("导入").
		WithLayout(layout.NoticeHeaderData{}).
		WithNotice("提示")).
		Convert("grade", exampleGradeExport).
		Rows(
			StudentRow{Name: "张三", Age: 18, Grade: 1},
			StudentRow{Name: "李四", Age: 20, Grade: 2},
		).
		Apply(); err != nil {
		t.Fatal(err)
	}
	_ = wb.File().SetSheetRow("导入", "A4", &[]string{"王五", "bad", "A"})

	_, res, err := excelizex.Read[StudentRow](wb.Sheet("导入").WithLayout(layout.NoticeHeaderData{})).
		Convert("grade", exampleGradeImport).
		Collect(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !res.HasErrors() {
		t.Fatal("expected bind error")
	}
	if err := wb.WriteErrors(res); err != nil {
		t.Fatal(err)
	}

	raw, err := wb.File().GetRows("导入")
	if err != nil {
		t.Fatal(err)
	}
	if len(raw) < 2 {
		t.Fatal("missing header row")
	}
	headers := raw[1]
	if headers[len(headers)-1] != "错误原因" {
		t.Fatalf("headers: %v", headers)
	}
}

func TestWithRowNilContext(t *testing.T) {
	c := excelizex.WithRow(nil, 5)
	if c.Row != 5 || c.Context == nil {
		t.Fatalf("context: row=%d ctx=%v", c.Row, c.Context)
	}
}

func TestConcurrencyOptionClamp(t *testing.T) {
	buf := buildNoticeGradeSheet(t, [][]string{{"张三", "18", "A"}})
	wb := openWorkbook(t, buf)

	var count atomic.Int32
	_, err := excelizex.Read[StudentRow](wb.Sheet("导入").WithLayout(layout.NoticeHeaderData{})).
		Convert("grade", exampleGradeImport).
		Each(context.Background(), func(_ excelizex.Context, _ StudentRow) error {
			count.Add(1)

			return nil
		}, excelizex.Concurrency(0))
	if err != nil {
		t.Fatal(err)
	}
	if count.Load() != 1 {
		t.Fatalf("processed %d want 1", count.Load())
	}
}

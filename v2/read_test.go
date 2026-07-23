package excelizex_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	excelizex "github.com/cyclonevox/excelizex/v2"
	"github.com/cyclonevox/excelizex/v2/layout"
	"github.com/cyclonevox/excelizex/v2/schema"
	"github.com/xuri/excelize/v2"
)

type twiceInt int

func (t *twiceInt) UnmarshalText(text []byte) error {
	n, err := strconv.Atoi(string(text))
	if err != nil {
		return err
	}
	*t = twiceInt(n * 2)

	return nil
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
			excelizex.Read[StudentRow](wb.Sheet("导入").WithLayout(layout.NoticeHeaderData{})),
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
			excelizex.Read[StudentRow](wb.Sheet("导入").WithLayout(layout.NoticeHeaderData{})),
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

func TestTextUnmarshalerRead(t *testing.T) {
	type twiceRow struct {
		N twiceInt `excel:"数值"`
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

	rows, _, err := excelizex.Read[twiceRow](wb.Sheet(sheet).WithLayout(layout.HeaderData{})).
		Collect(context.Background())
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

func TestResultHelpers(t *testing.T) {
	buf := buildNoticeGradeSheet(t, [][]string{{"张三", "18", "A"}})
	wb := openWorkbook(t, buf)

	_, res, err := excelizex.Read[StudentRow](wb.Sheet("导入").WithLayout(layout.NoticeHeaderData{})).
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

func TestWriteErrorsPadRow(t *testing.T) {
	wb := excelizex.New()
	t.Cleanup(func() { _ = wb.Close() })

	if err := excelizex.Write[StudentRow](wb.Sheet("导入").
		WithLayout(layout.NoticeHeaderData{}).
		WithNotice("提示")).
		Rows(
			StudentRow{Name: "张三", Age: 18, Grade: 1},
			StudentRow{Name: "李四", Age: 20, Grade: 2},
		).
		Apply(); err != nil {
		t.Fatal(err)
	}
	_ = wb.File().SetSheetRow("导入", "A4", &[]string{"王五", "bad", "A"})

	_, res, err := excelizex.Read[StudentRow](wb.Sheet("导入").WithLayout(layout.NoticeHeaderData{})).
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

func TestConcurrencyOptionClamp(t *testing.T) {
	buf := buildNoticeGradeSheet(t, [][]string{{"张三", "18", "A"}})
	wb := openWorkbook(t, buf)

	var count atomic.Int32
	_, err := excelizex.Read[StudentRow](wb.Sheet("导入").WithLayout(layout.NoticeHeaderData{})).
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

func TestEachDefaultConcurrencyIsOne(t *testing.T) {
	wb := excelizex.New()
	t.Cleanup(func() { _ = wb.Close() })
	rows := make([]simpleRow, 8)
	for i := range rows {
		rows[i] = simpleRow{Name: fmt.Sprintf("u%d", i), Age: i}
	}
	if err := excelizex.Write[simpleRow](wb.Sheet("导入").WithLayout(layout.HeaderData{})).
		Rows(rows...).Apply(); err != nil {
		t.Fatal(err)
	}

	var current, maxConcurrent atomic.Int32
	_, err := excelizex.Read[simpleRow](wb.Sheet("导入").WithLayout(layout.HeaderData{})).
		Each(context.Background(), func(_ excelizex.Context, _ simpleRow) error {
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
	rows := make([]simpleRow, 8)
	for i := range rows {
		rows[i] = simpleRow{Name: fmt.Sprintf("u%d", i), Age: i}
	}
	if err := excelizex.Write[simpleRow](wb.Sheet("导入").WithLayout(layout.HeaderData{})).
		Rows(rows...).Apply(); err != nil {
		t.Fatal(err)
	}

	var current, maxConcurrent atomic.Int32
	_, err := excelizex.Read[simpleRow](wb.Sheet("导入").WithLayout(layout.HeaderData{})).
		SetConcurrency(8).
		Each(context.Background(), func(_ excelizex.Context, _ simpleRow) error {
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

func TestWriteErrorsOrdersConcurrentFailuresByRow(t *testing.T) {
	wb := excelizex.New()
	t.Cleanup(func() { _ = wb.Close() })
	rows := make([]simpleRow, 20)
	for i := range rows {
		rows[i] = simpleRow{Name: fmt.Sprintf("u%d", i), Age: i}
	}
	if err := excelizex.Write[simpleRow](wb.Sheet("导入").WithLayout(layout.HeaderData{})).
		Rows(rows...).Apply(); err != nil {
		t.Fatal(err)
	}

	res, err := excelizex.Read[simpleRow](wb.Sheet("导入").WithLayout(layout.HeaderData{})).
		Each(context.Background(), func(ctx excelizex.Context, row simpleRow) error {
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

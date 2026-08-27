package main

import (
	"context"
	"io"
	"path/filepath"
	"testing"

	excelizex "github.com/cyclonevox/excelizex"
	"github.com/cyclonevox/excelizex/layout"
)

func TestEnrollmentPipeline(t *testing.T) {
	dir := t.TempDir()
	svc, err := run(dir, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	if svc.Count() != 6 {
		t.Fatalf("enrolled %d, want 6", svc.Count())
	}
	want := map[string]string{
		"110101199001011234": "张三",
		"110101199002021234": "李四",
		"110101199003031234": "周九",
		"110101199004041234": "赵六",
		"110101199005051234": "钱七",
		"110101199006061234": "王五",
	}
	got := map[string]string{}
	for _, row := range svc.List() {
		got[row.IDCard] = row.Name
	}
	for id, name := range want {
		if got[id] != name {
			t.Errorf("id %s: got %q want %q", id, got[id], name)
		}
	}

	wb, closer, err := openWorkbook(filepath.Join(dir, "5-已录取名单.xlsx"))
	if err != nil {
		t.Fatal(err)
	}
	defer closer()
	defer wb.Close()
	rows, res, err := excelizex.Read[student](wb.Sheet(sheetImport).
		WithLayout(layout.NoticeHeaderData{})).
		Collect(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if res.HasErrors() {
		t.Fatalf("roster errors: %v", res.Errors())
	}
	if len(rows) != 6 {
		t.Fatalf("roster rows %d, want 6", len(rows))
	}
}

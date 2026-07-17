package layout_test

import (
	"testing"

	"github.com/cyclonevox/excelizex/v2/layout"
)

func TestNoticeHeaderDataRows(t *testing.T) {
	l := layout.NoticeHeaderData{}
	row, ok := l.NoticeRow()
	if !ok || row != 1 {
		t.Fatalf("notice row: %d ok=%v", row, ok)
	}
	start, end := l.HeaderRows()
	if start != 2 || end != 2 {
		t.Fatalf("header rows: %d-%d", start, end)
	}
	if l.DataStartRow() != 3 {
		t.Fatalf("data start: %d", l.DataStartRow())
	}
}

func TestHeaderDataRows(t *testing.T) {
	l := layout.HeaderData{}
	if _, ok := l.NoticeRow(); ok {
		t.Fatal("expected no notice row")
	}
	start, end := l.HeaderRows()
	if start != 1 || end != 1 {
		t.Fatalf("header rows: %d-%d", start, end)
	}
	if l.DataStartRow() != 2 {
		t.Fatalf("data start: %d", l.DataStartRow())
	}
}

func TestResolveHeaders(t *testing.T) {
	l := layout.HeaderData{}
	m, err := l.ResolveHeaders([][]string{{"姓名", "年龄", ""}})
	if err != nil {
		t.Fatal(err)
	}
	if m[0] != "姓名" || m[1] != "年龄" {
		t.Fatalf("headers: %v", m)
	}
	if _, ok := m[2]; ok {
		t.Fatal("empty header should be skipped")
	}
}

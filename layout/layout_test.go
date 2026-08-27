package layout_test

import (
	"testing"

	"github.com/cyclonevox/excelizex/layout"
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

func TestStackedResolveLastNonEmpty(t *testing.T) {
	l := layout.Stacked{HasNotice: true}
	start, end := l.HeaderRows()
	if start != 2 || end != 3 {
		t.Fatalf("header rows: %d-%d", start, end)
	}
	if l.DataStartRow() != 4 {
		t.Fatalf("data start: %d", l.DataStartRow())
	}
	m, err := l.ResolveHeaders([][]string{
		{"基本信息", "", "成绩"},
		{"姓名", "年龄", "语文"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if m[0] != "姓名" || m[1] != "年龄" || m[2] != "语文" {
		t.Fatalf("headers: %v", m)
	}
}

func TestGroupedRenderAndMerge(t *testing.T) {
	l := layout.Grouped{
		HasNotice: true,
		Groups: []layout.Group{
			{Title: "基本信息", Span: 2},
			{Title: "成绩", Span: 2},
		},
	}
	rows, err := l.RenderHeaders([]string{"姓名", "年龄", "语文", "数学"})
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 {
		t.Fatalf("rows: %d", len(rows))
	}
	if rows[0][0] != "基本信息" || rows[0][1] != nil || rows[0][2] != "成绩" {
		t.Fatalf("group row: %v", rows[0])
	}
	if rows[1][0] != "姓名" || rows[1][3] != "数学" {
		t.Fatalf("leaf row: %v", rows[1])
	}
	merges := l.HeaderMerges(2, 4)
	if len(merges) != 2 || merges[0].From != 0 || merges[0].To != 1 || merges[1].From != 2 || merges[1].To != 3 {
		t.Fatalf("merges: %+v", merges)
	}
}

func TestRenderHeaders(t *testing.T) {
	rows, err := layout.HeaderData{}.RenderHeaders([]string{"姓名", "年龄"})
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0][0] != "姓名" || rows[0][1] != "年龄" {
		t.Fatalf("render: %v", rows)
	}
	rows, err = layout.NoticeHeaderData{}.RenderHeaders([]string{"A"})
	if err != nil || len(rows) != 1 {
		t.Fatalf("notice render: %v %v", rows, err)
	}
}

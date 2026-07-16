package schema_test

import (
	"testing"

	"github.com/cyclonevox/excelizex/v2/schema"
)

type inlineRow struct {
	Name string `excel:"姓名"`
	Addr struct {
		City   string `excel:"城市"`
		Street string `excel:"街道"`
	} `excel:",inline"`
	Skipped string `excel:"-"`
	Grade   int    `excel:"等级" conv:"grade"`
}

type noticeRow struct {
	Notice string `excel:"notice"`
	Name   string `excel:"姓名" validate:"required"`
}

func TestParseBasicAndInline(t *testing.T) {
	sc, err := schema.New(inlineRow{})
	if err != nil {
		t.Fatal(err)
	}
	if len(sc.Columns) != 4 {
		t.Fatalf("columns: got %d want 4", len(sc.Columns))
	}
	want := map[string]string{
		"姓名": "Name",
		"城市": "Addr.City",
		"街道": "Addr.Street",
		"等级": "Grade",
	}
	for _, c := range sc.Columns {
		if c.FieldPath != want[c.Header] {
			t.Fatalf("header %q field path %q want %q", c.Header, c.FieldPath, want[c.Header])
		}
	}
	col, ok := sc.ColumnByHeader("等级")
	if !ok || col.Convert != "grade" {
		t.Fatalf("conv tag: %+v", col)
	}
}

func TestParseIgnoreAndNotice(t *testing.T) {
	sc, err := schema.New(noticeRow{})
	if err != nil {
		t.Fatal(err)
	}
	if sc.Notice != "Notice" {
		t.Fatalf("notice field name: %q", sc.Notice)
	}
	if got := sc.RequiredHeaders(); len(got) != 1 || got[0] != "姓名" {
		t.Fatalf("required headers: %v", got)
	}
}

func TestParseStyleAndTime(t *testing.T) {
	type row struct {
		When string `excel:"日期" time:"2006-01-02" style:"header-red,body-blue"`
	}
	sc, err := schema.New(row{})
	if err != nil {
		t.Fatal(err)
	}
	c := sc.Columns[0]
	if c.TimeLayout != "2006-01-02" {
		t.Fatalf("time layout: %q", c.TimeLayout)
	}
	if len(c.Style) != 2 || c.Style[0] != "header-red" {
		t.Fatalf("style: %v", c.Style)
	}
}

func TestParseAnonymousEmbed(t *testing.T) {
	type inner struct {
		City string `excel:"城市"`
	}
	type row struct {
		inner
		Name string `excel:"姓名"`
	}
	sc, err := schema.New(row{})
	if err != nil {
		t.Fatal(err)
	}
	if len(sc.Columns) != 2 {
		t.Fatalf("columns: %d", len(sc.Columns))
	}
}

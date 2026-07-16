package bind_test

import (
	"testing"

	"github.com/cyclonevox/excelizex/v2/bind"
	"github.com/cyclonevox/excelizex/v2/convert"
	"github.com/cyclonevox/excelizex/v2/schema"
)

type studentRow struct {
	Name string `excel:"姓名" validate:"required"`
	Age  int    `excel:"年龄"`
}

func TestMatchColumnsMissing(t *testing.T) {
	sc, err := schema.New(studentRow{})
	if err != nil {
		t.Fatal(err)
	}
	_, err = bind.MatchColumns(sc, map[int]string{0: "姓名"})
	if err == nil {
		t.Fatal("expected missing column error")
	}
}

func TestMatchColumnsDuplicateHeader(t *testing.T) {
	sc, err := schema.New(studentRow{})
	if err != nil {
		t.Fatal(err)
	}
	_, err = bind.MatchColumns(sc, map[int]string{0: "姓名", 1: "姓名"})
	if err == nil {
		t.Fatal("expected duplicate header error")
	}
}

func TestBindRow(t *testing.T) {
	sc, err := schema.New(studentRow{})
	if err != nil {
		t.Fatal(err)
	}
	m, err := bind.MatchColumns(sc, map[int]string{0: "姓名", 1: "年龄"})
	if err != nil {
		t.Fatal(err)
	}
	row, err := bind.BindRow[studentRow](m, []string{"张三", "18"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if row.Name != "张三" || row.Age != 18 {
		t.Fatalf("row: %+v", row)
	}
}

func TestBindRowBadNumber(t *testing.T) {
	sc, _ := schema.New(studentRow{})
	m, _ := bind.MatchColumns(sc, map[int]string{0: "姓名", 1: "年龄"})
	_, err := bind.BindRow[studentRow](m, []string{"张三", "abc"}, nil)
	if err == nil {
		t.Fatal("expected convert error")
	}
}

func TestExtraHeaders(t *testing.T) {
	sc, _ := schema.New(studentRow{})
	m, _ := bind.MatchColumns(sc, map[int]string{0: "姓名", 1: "年龄", 2: "备注"})
	extra := bind.ExtraHeaders(m, map[int]string{0: "姓名", 1: "年龄", 2: "备注"}, sc)
	if len(extra) != 1 || extra[0] != "备注" {
		t.Fatalf("extra: %v", extra)
	}
	_ = convert.Registry{}
}

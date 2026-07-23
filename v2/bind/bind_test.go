package bind_test

import (
	"testing"

	"github.com/cyclonevox/excelizex/v2/bind"
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
	row, err := bind.BindRow[studentRow](m, []string{"张三", "18"})
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
	_, err := bind.BindRow[studentRow](m, []string{"张三", "abc"})
	if err == nil {
		t.Fatal("expected convert error")
	}
}

func TestBindInlineNestedPath(t *testing.T) {
	type address struct {
		City string `excel:"城市"`
	}
	type row struct {
		Name string  `excel:"姓名"`
		Addr address `excel:",inline"`
	}
	sc, err := schema.New(row{})
	if err != nil {
		t.Fatal(err)
	}
	m, err := bind.MatchColumns(sc, map[int]string{0: "姓名", 1: "城市"})
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 3; i++ {
		got, err := bind.BindRow[row](m, []string{"张三", "北京"})
		if err != nil {
			t.Fatal(err)
		}
		if got.Name != "张三" || got.Addr.City != "北京" {
			t.Fatalf("row: %+v", got)
		}
	}
}

func TestBindPointerInlineAlloc(t *testing.T) {
	type address struct {
		City string `excel:"城市"`
	}
	type row struct {
		Name string   `excel:"姓名"`
		Addr *address `excel:",inline"`
	}
	sc, err := schema.New(row{})
	if err != nil {
		t.Fatal(err)
	}
	m, err := bind.MatchColumns(sc, map[int]string{0: "姓名", 1: "城市"})
	if err != nil {
		t.Fatal(err)
	}
	got, err := bind.BindRow[row](m, []string{"李四", "上海"})
	if err != nil {
		t.Fatal(err)
	}
	if got.Addr == nil || got.Addr.City != "上海" {
		t.Fatalf("row: %+v", got)
	}
}


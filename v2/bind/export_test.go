package bind_test

import (
	"testing"
	"time"

	"github.com/cyclonevox/excelizex/v2/bind"
	"github.com/cyclonevox/excelizex/v2/convert"
	"github.com/cyclonevox/excelizex/v2/schema"
)

type exportRow struct {
	Name string    `excel:"姓名"`
	Age  int       `excel:"年龄"`
	When time.Time `excel:"日期" time:"2006-01-02"`
}

func TestExportRow(t *testing.T) {
	sc, err := schema.New(exportRow{})
	if err != nil {
		t.Fatal(err)
	}
	when, _ := time.Parse("2006-01-02", "2024-01-02")
	cells, err := bind.ExportRow(sc, exportRow{Name: "张三", Age: 18, When: when}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if cells[0] != "张三" || cells[1] != "18" || cells[2] != "2024-01-02" {
		t.Fatalf("cells: %v", cells)
	}
	reg := convert.ExportRegistry{}
	convert.ExportTo(reg, "noop", func(s string) (string, error) { return s, nil })
	_ = reg
}

package bind_test

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/cyclonevox/excelizex/v2/bind"
	"github.com/cyclonevox/excelizex/v2/schema"
)

type exportRow struct {
	Name string    `excel:"姓名"`
	Age  int       `excel:"年龄"`
	When time.Time `excel:"日期" time:"2006-01-02"`
}

type gradeChar int

func (g gradeChar) MarshalText() ([]byte, error) {
	switch g {
	case 1:
		return []byte("A"), nil
	case 2:
		return []byte("B"), nil
	default:
		return nil, fmt.Errorf("unknown grade %d", g)
	}
}

type exportGradeRow struct {
	Grade gradeChar `excel:"等级"`
}

func TestExportRow(t *testing.T) {
	sc, err := schema.New(exportRow{})
	if err != nil {
		t.Fatal(err)
	}
	when, _ := time.Parse("2006-01-02", "2024-01-02")
	cells, err := bind.ExportRow(sc, exportRow{Name: "张三", Age: 18, When: when})
	if err != nil {
		t.Fatal(err)
	}
	if cells[0] != "张三" || cells[1] != "18" || cells[2] != "2024-01-02" {
		t.Fatalf("cells: %v", cells)
	}
}

func TestExportRowTextMarshaler(t *testing.T) {
	sc, err := schema.New(exportGradeRow{})
	if err != nil {
		t.Fatal(err)
	}
	cells, err := bind.ExportRow(sc, exportGradeRow{Grade: 1})
	if err != nil {
		t.Fatal(err)
	}
	if cells[0] != "A" {
		t.Fatalf("cells: %v", cells)
	}
	_ = reflect.ValueOf(gradeChar(1))
}

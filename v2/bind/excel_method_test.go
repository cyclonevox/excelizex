package bind_test

import (
	"fmt"
	"testing"

	"github.com/cyclonevox/excelizex/v2/bind"
	"github.com/cyclonevox/excelizex/v2/schema"
)

type excelGradeRow struct {
	Name  string `excel:"姓名"`
	Grade int    `excel:"年级"`
}

func (r *excelGradeRow) ExcelGrade(raw string) error {
	switch raw {
	case "A":
		r.Grade = 1
		return nil
	case "B":
		r.Grade = 2
		return nil
	case "":
		r.Grade = 0
		return nil
	default:
		return fmt.Errorf("unknown grade %q", raw)
	}
}

func (r *excelGradeRow) ExcelExportGrade() (string, error) {
	switch r.Grade {
	case 1:
		return "A", nil
	case 2:
		return "B", nil
	case 0:
		return "", nil
	default:
		return "", fmt.Errorf("unknown grade %d", r.Grade)
	}
}

type excelCatchAllRow struct {
	Name  string `excel:"姓名"`
	Level int    `excel:"等级"`
	Note  string `excel:"备注"`
}

func (r *excelCatchAllRow) ExcelField(header, raw string) (bool, error) {
	switch header {
	case "等级":
		switch raw {
		case "高":
			r.Level = 3
			return true, nil
		case "低":
			r.Level = 1
			return true, nil
		default:
			return true, fmt.Errorf("bad level %q", raw)
		}
	default:
		return false, nil
	}
}

func (r *excelCatchAllRow) ExcelExportField(header string) (string, bool, error) {
	if header != "等级" {
		return "", false, nil
	}
	switch r.Level {
	case 3:
		return "高", true, nil
	case 1:
		return "低", true, nil
	default:
		return "", true, fmt.Errorf("bad level %d", r.Level)
	}
}

func TestExcelImportPlanCached(t *testing.T) {
	sc, err := schema.New(excelGradeRow{})
	if err != nil {
		t.Fatal(err)
	}
	m, err := bind.MatchColumns(sc, map[int]string{0: "姓名", 1: "年级"})
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 3; i++ {
		if _, err := bind.BindRow[excelGradeRow](m, []string{"张三", "A"}); err != nil {
			t.Fatal(err)
		}
	}
}

func TestBindExcelFieldMethod(t *testing.T) {
	sc, err := schema.New(excelGradeRow{})
	if err != nil {
		t.Fatal(err)
	}
	m, err := bind.MatchColumns(sc, map[int]string{0: "姓名", 1: "年级"})
	if err != nil {
		t.Fatal(err)
	}
	row, err := bind.BindRow[excelGradeRow](m, []string{"张三", "A"})
	if err != nil {
		t.Fatal(err)
	}
	if row.Name != "张三" || row.Grade != 1 {
		t.Fatalf("row: %+v", row)
	}
}

func TestBindExcelFieldCatchAll(t *testing.T) {
	sc, err := schema.New(excelCatchAllRow{})
	if err != nil {
		t.Fatal(err)
	}
	m, err := bind.MatchColumns(sc, map[int]string{0: "姓名", 1: "等级", 2: "备注"})
	if err != nil {
		t.Fatal(err)
	}
	row, err := bind.BindRow[excelCatchAllRow](m, []string{"李四", "高", "ok"})
	if err != nil {
		t.Fatal(err)
	}
	if row.Name != "李四" || row.Level != 3 || row.Note != "ok" {
		t.Fatalf("row: %+v", row)
	}
}

func TestExportExcelMethod(t *testing.T) {
	sc, err := schema.New(excelGradeRow{})
	if err != nil {
		t.Fatal(err)
	}
	cells, err := bind.ExportRow(sc, excelGradeRow{Name: "张三", Grade: 2})
	if err != nil {
		t.Fatal(err)
	}
	if cells[0] != "张三" || cells[1] != "B" {
		t.Fatalf("cells: %v", cells)
	}
}

func TestExportExcelCatchAll(t *testing.T) {
	sc, err := schema.New(excelCatchAllRow{})
	if err != nil {
		t.Fatal(err)
	}
	cells, err := bind.ExportRow(sc, excelCatchAllRow{Name: "王五", Level: 1, Note: "x"})
	if err != nil {
		t.Fatal(err)
	}
	if cells[0] != "王五" || cells[1] != "低" || cells[2] != "x" {
		t.Fatalf("cells: %v", cells)
	}
}

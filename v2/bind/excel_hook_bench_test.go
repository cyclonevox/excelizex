package bind_test

import (
	"fmt"
	"testing"

	"github.com/cyclonevox/excelizex/v2/bind"
	"github.com/cyclonevox/excelizex/v2/schema"
)

// --- reflect path: Excel{Field} per column ---

type benchReflectRow struct {
	Name   string `excel:"姓名"`
	Grade1 int    `excel:"等级1"`
	Grade2 int    `excel:"等级2"`
	Grade3 int    `excel:"等级3"`
	Grade4 int    `excel:"等级4"`
	Grade5 int    `excel:"等级5"`
}

func parseBenchGrade(raw string) (int, error) {
	switch raw {
	case "A":
		return 1, nil
	case "B":
		return 2, nil
	case "":
		return 0, nil
	default:
		return 0, fmt.Errorf("bad grade %q", raw)
	}
}

func formatBenchGrade(n int) (string, error) {
	switch n {
	case 1:
		return "A", nil
	case 2:
		return "B", nil
	case 0:
		return "", nil
	default:
		return "", fmt.Errorf("bad grade %d", n)
	}
}

func (r *benchReflectRow) ExcelGrade1(raw string) error {
	n, err := parseBenchGrade(raw)
	r.Grade1 = n
	return err
}
func (r *benchReflectRow) ExcelGrade2(raw string) error {
	n, err := parseBenchGrade(raw)
	r.Grade2 = n
	return err
}
func (r *benchReflectRow) ExcelGrade3(raw string) error {
	n, err := parseBenchGrade(raw)
	r.Grade3 = n
	return err
}
func (r *benchReflectRow) ExcelGrade4(raw string) error {
	n, err := parseBenchGrade(raw)
	r.Grade4 = n
	return err
}
func (r *benchReflectRow) ExcelGrade5(raw string) error {
	n, err := parseBenchGrade(raw)
	r.Grade5 = n
	return err
}

func (r *benchReflectRow) ExcelExportGrade1() (string, error) { return formatBenchGrade(r.Grade1) }
func (r *benchReflectRow) ExcelExportGrade2() (string, error) { return formatBenchGrade(r.Grade2) }
func (r *benchReflectRow) ExcelExportGrade3() (string, error) { return formatBenchGrade(r.Grade3) }
func (r *benchReflectRow) ExcelExportGrade4() (string, error) { return formatBenchGrade(r.Grade4) }
func (r *benchReflectRow) ExcelExportGrade5() (string, error) { return formatBenchGrade(r.Grade5) }

// --- interface path: ExcelField / ExcelExportField ---

type benchIfaceRow struct {
	Name   string `excel:"姓名"`
	Grade1 int    `excel:"等级1"`
	Grade2 int    `excel:"等级2"`
	Grade3 int    `excel:"等级3"`
	Grade4 int    `excel:"等级4"`
	Grade5 int    `excel:"等级5"`
}

func (r *benchIfaceRow) ExcelField(header, raw string) (bool, error) {
	switch header {
	case "等级1":
		n, err := parseBenchGrade(raw)
		r.Grade1 = n
		return true, err
	case "等级2":
		n, err := parseBenchGrade(raw)
		r.Grade2 = n
		return true, err
	case "等级3":
		n, err := parseBenchGrade(raw)
		r.Grade3 = n
		return true, err
	case "等级4":
		n, err := parseBenchGrade(raw)
		r.Grade4 = n
		return true, err
	case "等级5":
		n, err := parseBenchGrade(raw)
		r.Grade5 = n
		return true, err
	default:
		return false, nil
	}
}

func (r *benchIfaceRow) ExcelExportField(header string) (string, bool, error) {
	switch header {
	case "等级1":
		s, err := formatBenchGrade(r.Grade1)
		return s, true, err
	case "等级2":
		s, err := formatBenchGrade(r.Grade2)
		return s, true, err
	case "等级3":
		s, err := formatBenchGrade(r.Grade3)
		return s, true, err
	case "等级4":
		s, err := formatBenchGrade(r.Grade4)
		return s, true, err
	case "等级5":
		s, err := formatBenchGrade(r.Grade5)
		return s, true, err
	default:
		return "", false, nil
	}
}

// --- baseline: builtin int only (no Excel hooks) ---

type benchBuiltinRow struct {
	Name   string `excel:"姓名"`
	Grade1 int    `excel:"等级1"`
	Grade2 int    `excel:"等级2"`
	Grade3 int    `excel:"等级3"`
	Grade4 int    `excel:"等级4"`
	Grade5 int    `excel:"等级5"`
}

var (
	benchHeaders = map[int]string{
		0: "姓名", 1: "等级1", 2: "等级2", 3: "等级3", 4: "等级4", 5: "等级5",
	}
	benchCellsAB = []string{"张三", "A", "B", "A", "B", "A"}
	benchCellsNum = []string{"张三", "1", "2", "1", "2", "1"}
)

func mustMapping[T any](b *testing.B) bind.Mapping {
	b.Helper()
	var zero T
	sc, err := schema.New(zero)
	if err != nil {
		b.Fatal(err)
	}
	m, err := bind.MatchColumns(sc, benchHeaders)
	if err != nil {
		b.Fatal(err)
	}
	return m
}

func BenchmarkBindRow_ExcelReflect(b *testing.B) {
	m := mustMapping[benchReflectRow](b)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		row, err := bind.BindRow[benchReflectRow](m, benchCellsAB)
		if err != nil {
			b.Fatal(err)
		}
		if row.Grade1 != 1 || row.Grade5 != 1 {
			b.Fatalf("row: %+v", row)
		}
	}
}

func BenchmarkBindRow_ExcelIface(b *testing.B) {
	m := mustMapping[benchIfaceRow](b)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		row, err := bind.BindRow[benchIfaceRow](m, benchCellsAB)
		if err != nil {
			b.Fatal(err)
		}
		if row.Grade1 != 1 || row.Grade5 != 1 {
			b.Fatalf("row: %+v", row)
		}
	}
}

func BenchmarkBindRow_BuiltinOnly(b *testing.B) {
	m := mustMapping[benchBuiltinRow](b)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		row, err := bind.BindRow[benchBuiltinRow](m, benchCellsNum)
		if err != nil {
			b.Fatal(err)
		}
		if row.Grade1 != 1 || row.Grade5 != 1 {
			b.Fatalf("row: %+v", row)
		}
	}
}

func BenchmarkExportRow_ExcelReflect(b *testing.B) {
	sc, err := schema.New(benchReflectRow{})
	if err != nil {
		b.Fatal(err)
	}
	row := benchReflectRow{Name: "张三", Grade1: 1, Grade2: 2, Grade3: 1, Grade4: 2, Grade5: 1}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cells, err := bind.ExportRow(sc, row)
		if err != nil {
			b.Fatal(err)
		}
		if cells[1] != "A" || cells[2] != "B" {
			b.Fatalf("cells: %v", cells)
		}
	}
}

func BenchmarkExportRow_ExcelIface(b *testing.B) {
	sc, err := schema.New(benchIfaceRow{})
	if err != nil {
		b.Fatal(err)
	}
	row := benchIfaceRow{Name: "张三", Grade1: 1, Grade2: 2, Grade3: 1, Grade4: 2, Grade5: 1}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cells, err := bind.ExportRow(sc, row)
		if err != nil {
			b.Fatal(err)
		}
		if cells[1] != "A" || cells[2] != "B" {
			b.Fatalf("cells: %v", cells)
		}
	}
}

func BenchmarkExportRow_BuiltinOnly(b *testing.B) {
	sc, err := schema.New(benchBuiltinRow{})
	if err != nil {
		b.Fatal(err)
	}
	row := benchBuiltinRow{Name: "张三", Grade1: 1, Grade2: 2, Grade3: 1, Grade4: 2, Grade5: 1}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cells, err := bind.ExportRow(sc, row)
		if err != nil {
			b.Fatal(err)
		}
		if cells[1] != "1" || cells[2] != "2" {
			b.Fatalf("cells: %v", cells)
		}
	}
}

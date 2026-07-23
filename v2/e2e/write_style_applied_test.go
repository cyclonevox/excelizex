// 业务场景：Write Template 时 style tag 应用到表头与数据列。
package e2e_test

import (
	"fmt"
	"testing"

	excelizex "github.com/cyclonevox/excelizex/v2"
	"github.com/cyclonevox/excelizex/v2/layout"
)

type styledImportRow struct {
	Name  string `excel:"姓名" style:"header-red,body"`
	Age   int    `excel:"年龄" style:"header,body"`
	Grade int    `excel:"年级" style:"header,body-blue"`
}

func (r *styledImportRow) ExcelGrade(raw string) error {
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

func (r *styledImportRow) ExcelExportGrade() (string, error) {
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

func TestWriteStyleApplied(t *testing.T) {
	wb := excelizex.New()
	defer wb.Close()
	if err := excelizex.Write[styledImportRow](wb.Sheet("样式").
		WithLayout(layout.HeaderData{})).
		Template().
		Apply(); err != nil {
		t.Fatal(err)
	}
	f := wb.File()
	styleID, err := f.GetCellStyle("样式", "A1")
	if err != nil {
		t.Fatal(err)
	}
	if styleID == 0 {
		t.Fatal("expected header style on A1")
	}
	styleID, err = f.GetCellStyle("样式", "C2")
	if err != nil {
		t.Fatal(err)
	}
	if styleID == 0 {
		t.Fatal("expected body style on grade column")
	}
}

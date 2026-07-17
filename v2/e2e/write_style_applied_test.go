// 业务场景：Write Template 时 style tag 应用到表头与数据列。
package e2e_test

import (
	"testing"

	excelizex "github.com/cyclonevox/excelizex/v2"
	"github.com/cyclonevox/excelizex/v2/e2e/fixture"
	"github.com/cyclonevox/excelizex/v2/layout"
)

type styledImportRow struct {
	Name  string `excel:"姓名" style:"header-red,body"`
	Age   int    `excel:"年龄" style:"header,body"`
	Grade int    `excel:"年级" conv:"grade" style:"header,body-blue"`
}

func TestWriteStyleApplied(t *testing.T) {
	wb := excelizex.New()
	defer wb.Close()
	if err := excelizex.Write[styledImportRow](wb.Sheet("样式").
		WithLayout(layout.HeaderData{})).
		Convert("grade", fixture.GradeExport).
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

// 业务场景：用户 Excel 数据区中间夹空行，导入时正确跳过。
package e2e_test

import (
	"context"
	"testing"

	excelizex "github.com/cyclonevox/excelizex/v2"
	"github.com/cyclonevox/excelizex/v2/e2e/fixture"
	"github.com/cyclonevox/excelizex/v2/layout"
)

func TestDirtyEmptyRowsSkipped(t *testing.T) {
	wb := fixture.OpenTestdata(t, "students_empty_rows.xlsx")
	defer wb.Close()

	rows, res, err := excelizex.Read[fixture.StudentImportRow](wb.Sheet(fixture.SheetStudentImport).
		WithLayout(layout.NoticeHeaderData{})).
		Convert("grade", fixture.GradeImport).
		Validate(fixture.StructValidator()).
		Collect(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if res.HasErrors() {
		t.Fatalf("errors: %v", res.Errors())
	}
	if len(rows) != 2 || rows[0].Name != "张三" || rows[1].Name != "李四" {
		t.Fatalf("rows: %+v", rows)
	}
}

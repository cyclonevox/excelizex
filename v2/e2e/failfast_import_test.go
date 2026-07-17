// 业务场景：FailFast 模式下遇到首条坏行即停止 Collect。
package e2e_test

import (
	"context"
	"testing"

	excelizex "github.com/cyclonevox/excelizex/v2"
	"github.com/cyclonevox/excelizex/v2/e2e/fixture"
	"github.com/cyclonevox/excelizex/v2/layout"
)

func TestFailFastImport(t *testing.T) {
	wb := fixture.OpenTestdata(t, "students_notice_failfast.xlsx")
	defer wb.Close()

	rows, res, err := excelizex.Read[fixture.StudentImportRow](wb.Sheet(fixture.SheetStudentImport).
		WithLayout(layout.NoticeHeaderData{})).
		Convert("grade", fixture.GradeImport).
		Validate(fixture.StructValidator()).
		SetFailFast().
		Collect(context.Background())
	if err == nil {
		t.Fatal("expected fail-fast error")
	}
	if len(rows) != 1 || rows[0].Name != "张三" {
		t.Fatalf("fail-fast rows: %+v", rows)
	}
	if len(res.Errors()) != 1 {
		t.Fatalf("errors: %v", res.Errors())
	}
}

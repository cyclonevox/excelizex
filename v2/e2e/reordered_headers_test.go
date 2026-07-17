// 业务场景：用户拖乱表头顺序并多加无关列、备注列，库仍能正确绑定。
package e2e_test

import (
	"context"
	"testing"

	excelizex "github.com/cyclonevox/excelizex/v2"
	"github.com/cyclonevox/excelizex/v2/e2e/fixture"
	"github.com/cyclonevox/excelizex/v2/layout"
)

func TestReorderedHeadersWithExtraColumn(t *testing.T) {
	wb := fixture.OpenTestdata(t, "students_reordered.xlsx")
	defer wb.Close()

	// 夹具 notice 文案与 NoticeFillStudents 不同，仅显式 WithLayout。
	rows, res, err := excelizex.Read[fixture.ReorderedRow](wb.Sheet(fixture.SheetStudentImport).
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
	if len(rows) != 1 {
		t.Fatalf("rows: %d", len(rows))
	}
	if rows[0].Name != "张三" || rows[0].Age != 30 || rows[0].Grade != 1 || rows[0].Extra != "ok" {
		t.Fatalf("row: %+v", rows[0])
	}
}

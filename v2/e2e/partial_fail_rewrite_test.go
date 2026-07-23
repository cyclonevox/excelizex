// 业务场景：部分行校验失败 → WriteErrors 回写错误原因 → 修正后再导入成功。
package e2e_test

import (
	"context"
	"testing"

	excelizex "github.com/cyclonevox/excelizex/v2"
	"github.com/cyclonevox/excelizex/v2/e2e/fixture"
	"github.com/cyclonevox/excelizex/v2/layout"
)

func TestPartialFailRewriteAndReimport(t *testing.T) {
	wb := fixture.OpenTestdata(t, "students_notice_partial_fail.xlsx")

	_, res, err := excelizex.Read[fixture.StudentImportRow](wb.Sheet(fixture.SheetStudentImport).
		WithLayout(layout.NoticeHeaderData{})).
		Validate(fixture.StructValidator()).
		Collect(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	failCount := len(res.Errors())
	if failCount < 2 {
		t.Fatalf("expected >=2 errors, got %d: %v", failCount, res.Errors())
	}
	if err := wb.WriteErrors(res); err != nil {
		t.Fatal(err)
	}
	out := fixture.SaveToBytes(t, wb)
	wb2 := fixture.OpenBytes(t, out)
	defer wb2.Close()

	rows, res2, err := excelizex.Read[fixture.StudentImportRow](wb2.Sheet(fixture.SheetStudentImport).
		WithLayout(layout.NoticeHeaderData{})).
		Validate(fixture.StructValidator()).
		Collect(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 0 {
		t.Fatalf("reopened ok rows: %d", len(rows))
	}
	if len(res2.Errors()) != failCount {
		t.Fatalf("reopened error rows: got %d want %d", len(res2.Errors()), failCount)
	}
	raw, err := wb2.File().GetRows(fixture.SheetStudentImport)
	if err != nil {
		t.Fatal(err)
	}
	if len(raw) < 2 {
		t.Fatal("missing header row")
	}
	lastHeader := raw[1][len(raw[1])-1]
	if lastHeader != "错误原因" {
		t.Fatalf("error column header: %q", lastHeader)
	}

	// 业务方修正失败行后重新导入（已提交夹具 students_notice_fixed.xlsx）
	wb3 := fixture.OpenTestdata(t, "students_notice_fixed.xlsx")
	defer wb3.Close()
	rows3, res3, err := excelizex.Read[fixture.StudentImportRow](wb3.Sheet(fixture.SheetStudentImport).
		WithLayout(layout.NoticeHeaderData{})).
		Validate(fixture.StructValidator()).
		Collect(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if res3.HasErrors() {
		t.Fatalf("reimport errors: %v", res3.Errors())
	}
	if len(rows3) != 2 || rows3[1].Name != "钱七" {
		t.Fatalf("reimport rows: %+v", rows3)
	}
}

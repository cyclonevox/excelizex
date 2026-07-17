// 业务场景：复刻原 testdata 考生表（notice：张三 ok + 李四 bad age；header：王五）。
package e2e_test

import (
	"context"
	"testing"

	excelizex "github.com/cyclonevox/excelizex/v2"
	"github.com/cyclonevox/excelizex/v2/e2e/fixture"
	"github.com/cyclonevox/excelizex/v2/layout"
)

func TestLegacyNoticeStudentsImport(t *testing.T) {
	buf := fixture.BuildLegacyNoticeStudents(t)
	wb := fixture.OpenBytes(t, buf)
	defer wb.Close()

	rows, res, err := excelizex.Read[fixture.LegacyStudentRow](wb.Sheet(fixture.SheetStudentImport).
		WithLayout(layout.NoticeHeaderData{})).
		Convert("grade", fixture.GradeImport).
		Validate(fixture.StructValidator()).
		Collect(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].Name != "张三" || rows[0].Grade != 1 {
		t.Fatalf("rows: %+v", rows)
	}
	if len(res.Errors()) != 1 {
		t.Fatalf("errors: %v", res.Errors())
	}
}

func TestLegacyHeaderStudentsImport(t *testing.T) {
	buf := fixture.BuildLegacyHeaderStudents(t)
	wb := fixture.OpenBytes(t, buf)
	defer wb.Close()

	rows, res, err := excelizex.Read[fixture.LegacyStudentRow](wb.Sheet(fixture.SheetStudentImport).
		WithLayout(layout.HeaderData{})).
		Convert("grade", fixture.GradeImport).
		Collect(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].Name != "王五" {
		t.Fatalf("rows: %+v", rows)
	}
	if res.HasErrors() {
		t.Fatalf("unexpected errors: %v", res.Errors())
	}
}

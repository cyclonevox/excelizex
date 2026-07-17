// 业务场景：考务系统批量导入考生名单（happy path）。
// 流程：生成导入表 → Open → Read Collect（conv + Validate hook）→ 断言成功行。
package e2e_test

import (
	"context"
	"strings"
	"testing"

	excelizex "github.com/cyclonevox/excelizex/v2"
	"github.com/cyclonevox/excelizex/v2/e2e/fixture"
	"github.com/cyclonevox/excelizex/v2/layout"
)

func TestStudentBatchImportHappyPath(t *testing.T) {
	buf := fixture.BuildDirtyNoticeImport(t, [][]string{
		{"张三", "110101199001011234", "18", "A"},
		{"李四", "110101199002021234", "20", "B"},
	})
	wb := fixture.OpenBytes(t, buf)
	defer wb.Close()

	rows, res, err := excelizex.Read[fixture.StudentImportRow](wb.Sheet(fixture.SheetStudentImport).
		WithLayout(layout.NoticeHeaderData{}).
		WithNotice(fixture.NoticeFillStudents)).
		Convert("grade", fixture.GradeImport).
		Validate(fixture.StructValidator()).
		Collect(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Errors()) != 0 {
		t.Fatalf("unexpected errors: %v", res.Errors())
	}
	if len(rows) != 2 {
		t.Fatalf("rows: got %d want 2", len(rows))
	}
	if rows[0].Name != "张三" || rows[0].Age != 18 || rows[0].Grade != 1 {
		t.Fatalf("row0: %+v", rows[0])
	}
	if rows[1].Name != "李四" || rows[1].Grade != 2 {
		t.Fatalf("row1: %+v", rows[1])
	}
}

func TestStudentBatchImportEmptyName(t *testing.T) {
	buf := fixture.BuildDirtyNoticeImport(t, [][]string{
		{"张三", "110101199001011234", "18", "A"},
		{"", "110101199002021234", "20", "B"},
	})
	wb := fixture.OpenBytes(t, buf)
	defer wb.Close()

	rows, res, err := excelizex.Read[fixture.StudentImportRow](wb.Sheet(fixture.SheetStudentImport).
		WithLayout(layout.NoticeHeaderData{}).
		WithNotice(fixture.NoticeFillStudents)).
		Convert("grade", fixture.GradeImport).
		Validate(fixture.StructValidator()).
		Collect(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 {
		t.Fatalf("ok rows: got %d want 1", len(rows))
	}
	if rows[0].Name != "张三" {
		t.Fatalf("row0: %+v", rows[0])
	}
	errs := res.Errors()
	if len(errs) != 1 {
		t.Fatalf("errors: got %d want 1: %v", len(errs), errs)
	}
	if errs[0].Row != 4 {
		t.Fatalf("error row: got %d want 4", errs[0].Row)
	}
	msg := strings.Join(errs[0].Messages, "; ")
	if !strings.Contains(msg, "Name") && !strings.Contains(msg, "required") {
		t.Fatalf("expected tag-driven validation message, got %q", msg)
	}
}

// 业务场景：FailFast 模式下遇到首条坏行即停止 Collect。
package e2e_test

import (
	"context"
	"testing"

	excelizex "github.com/cyclonevox/excelizex/v2"
	"github.com/cyclonevox/excelizex/v2/e2e/fixture"
)

func TestFailFastImport(t *testing.T) {
	buf := fixture.BuildDirtyNoticeImport(t, [][]string{
		{"张三", "", "18", "A"},
		{"李四", "", "bad-age", "A"},
		{"王五", "", "20", "A"},
	})
	wb := fixture.OpenBytes(t, buf)
	defer wb.Close()

	rows, res, err := excelizex.Read[fixture.StudentImportRow](wb.Sheet(fixture.SheetStudentImport)).
		Convert("grade", fixture.GradeImport).
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

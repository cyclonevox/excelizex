// 业务场景：缺列或非法数字时返回清晰错误，不 panic。
package e2e_test

import (
	"context"
	"testing"

	excelizex "github.com/cyclonevox/excelizex/v2"
	"github.com/cyclonevox/excelizex/v2/e2e/fixture"
)

func TestMissingColumnAndBadType(t *testing.T) {
	buf := fixture.BuildMissingColumnFile(t)
	wb := fixture.OpenBytes(t, buf)
	defer wb.Close()

	_, _, err := excelizex.Read[fixture.StudentImportRow](wb.Sheet("Sheet1")).
		Collect(context.Background())
	if err == nil {
		t.Fatal("expected missing column error")
	}

	buf2 := fixture.BuildDirtyNoticeImport(t, [][]string{{"张三", "", "not-int", "A"}})
	wb2 := fixture.OpenBytes(t, buf2)
	defer wb2.Close()
	_, res, err := excelizex.Read[fixture.StudentImportRow](wb2.Sheet(fixture.SheetStudentImport)).
		Convert("grade", fixture.GradeImport).
		Collect(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !res.HasErrors() {
		t.Fatal("expected type conversion error in result")
	}
}

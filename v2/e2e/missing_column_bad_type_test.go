// 业务场景：缺列或非法数字时返回清晰错误，不 panic。
package e2e_test

import (
	"context"
	"testing"

	excelizex "github.com/cyclonevox/excelizex/v2"
	"github.com/cyclonevox/excelizex/v2/e2e/fixture"
	"github.com/cyclonevox/excelizex/v2/layout"
)

func TestMissingColumnAndBadType(t *testing.T) {
	wb := fixture.OpenTestdata(t, "students_missing_column.xlsx")
	defer wb.Close()

	// 缺列在绑定阶段失败，Validate 尚未执行。
	_, _, err := excelizex.Read[fixture.StudentImportRow](wb.Sheet("Sheet1").
		WithLayout(layout.NoticeHeaderData{})).
		Collect(context.Background())
	if err == nil {
		t.Fatal("expected missing column error")
	}

	wb2 := fixture.OpenTestdata(t, "students_notice_bad_type.xlsx")
	defer wb2.Close()
	// 年龄列类型转换失败，在 Validate 之前即记入 Result。
	_, res, err := excelizex.Read[fixture.StudentImportRow](wb2.Sheet(fixture.SheetStudentImport).
		WithLayout(layout.NoticeHeaderData{})).
		Convert("grade", fixture.GradeImport).
		Collect(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !res.HasErrors() {
		t.Fatal("expected type conversion error in result")
	}
}

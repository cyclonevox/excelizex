// 业务场景：发模板给业务方（Template + Dropdown + Protect）→ 落盘 → 重开填表 → Read。
package e2e_test

import (
	"context"
	"testing"

	excelizex "github.com/cyclonevox/excelizex/v2"
	"github.com/cyclonevox/excelizex/v2/e2e/fixture"
)

func TestTemplateDistributeFillRead(t *testing.T) {
	dir := t.TempDir()
	path := fixture.WriteTemplateDistribute(t, dir)

	wb, closeFn := fixture.OpenPath(t, path)
	defer closeFn()

	dvs, err := wb.File().GetDataValidations(fixture.SheetStudentImport)
	if err != nil {
		t.Fatal(err)
	}
	if len(dvs) == 0 {
		t.Fatal("expected dropdown validation on template")
	}
	if err := wb.File().UnprotectSheet(fixture.SheetStudentImport, "dist-secret"); err != nil {
		t.Fatalf("unprotect: %v", err)
	}
	fixture.FillTemplateRows(t, wb, [][]string{
		{"张三", "A"},
		{"李四", "B"},
	})

	rows, res, err := excelizex.Read[fixture.TemplateDistributeRow](wb.Sheet(fixture.SheetStudentImport)).
		Validate(fixture.StructValidator()).
		Collect(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if res.HasErrors() {
		t.Fatalf("errors: %v", res.Errors())
	}
	if len(rows) != 2 || rows[0].Name != "张三" || rows[1].Level != "B" {
		t.Fatalf("rows: %+v", rows)
	}
}

// 业务场景：无 notice 的 HeaderData 布局导出成绩表后再导入。
package e2e_test

import (
	"context"
	"testing"

	excelizex "github.com/cyclonevox/excelizex/v2"
	"github.com/cyclonevox/excelizex/v2/e2e/fixture"
	"github.com/cyclonevox/excelizex/v2/layout"
)

func TestHeaderDataExportReimport(t *testing.T) {
	wb := fixture.OpenTestdata(t, "scores_header.xlsx")
	defer wb.Close()

	rows, res, err := excelizex.Read[fixture.ScoreRow](wb.Sheet("成绩").WithLayout(layout.HeaderData{})).
		Collect(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if res.HasErrors() {
		t.Fatalf("errors: %v", res.Errors())
	}
	if len(rows) != 2 || rows[0].Name != "李四" || rows[1].Score != 88 {
		t.Fatalf("rows: %+v", rows)
	}
}

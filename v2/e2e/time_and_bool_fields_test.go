// 业务场景：启用状态（bool）与入学日期（time.Time）字段的 write + read round-trip。
package e2e_test

import (
	"context"
	"testing"

	excelizex "github.com/cyclonevox/excelizex/v2"
	"github.com/cyclonevox/excelizex/v2/e2e/fixture"
	"github.com/cyclonevox/excelizex/v2/layout"
)

func TestTimeAndBoolFieldsRoundTrip(t *testing.T) {
	wb := fixture.OpenTestdata(t, "time_bool.xlsx")
	defer wb.Close()

	rows, res, err := excelizex.Read[fixture.TimeBoolRow](wb.Sheet(fixture.SheetStudentImport).
		WithLayout(layout.NoticeHeaderData{})).
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
	if !rows[0].Active || rows[0].Joined.Format("2006-01-02") != "2024-09-01" {
		t.Fatalf("row0: %+v", rows[0])
	}
	if rows[1].Active {
		t.Fatalf("row1 active: %+v", rows[1])
	}
}

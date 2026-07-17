// 业务场景：嵌套地址 struct 用 ,inline flatten 后 write + read round-trip。
package e2e_test

import (
	"context"
	"testing"

	excelizex "github.com/cyclonevox/excelizex/v2"
	"github.com/cyclonevox/excelizex/v2/e2e/fixture"
	"github.com/cyclonevox/excelizex/v2/layout"
)

func TestInlineAddressWriteRead(t *testing.T) {
	wb := fixture.OpenTestdata(t, "inline_address.xlsx")
	defer wb.Close()

	rows, res, err := excelizex.Read[fixture.InlineAddressRow](wb.Sheet("地址导入").
		WithLayout(layout.NoticeHeaderData{})).
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
	if rows[0].Name != "张三" || rows[0].Addr.City != "北京" || rows[0].Addr.Street != "长安街" {
		t.Fatalf("row: %+v", rows[0])
	}

	// 库 round-trip 导出再导入
	wbOut := excelizex.New()
	defer wbOut.Close()
	if err := excelizex.Write[fixture.InlineAddressRow](wbOut.Sheet("地址导入").
		WithLayout(layout.NoticeHeaderData{})).
		Rows(rows[0]).
		Apply(); err != nil {
		t.Fatal(err)
	}
	wb2 := fixture.RoundTripSave(t, wbOut)
	rows2, res2, err := excelizex.Read[fixture.InlineAddressRow](wb2.Sheet("地址导入").
		WithLayout(layout.NoticeHeaderData{})).
		Collect(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if res2.HasErrors() || len(rows2) != 1 || rows2[0].Addr.City != "北京" {
		t.Fatalf("round-trip: rows=%+v errors=%v", rows2, res2.Errors())
	}
}

// 业务场景：动态科目列、分组表头、流式导出大批量成绩。
package e2e_test

import (
	"context"
	"fmt"
	"testing"

	excelizex "github.com/cyclonevox/excelizex"
	"github.com/cyclonevox/excelizex/layout"
)

type examRow struct {
	Name  string            `excel:"姓名"`
	Extra map[string]string `excel:",extend"`
}

func TestDynamicSubjectColumnsRoundTrip(t *testing.T) {
	wb := excelizex.New()
	defer wb.Close()
	if err := excelizex.Write[examRow](wb.Sheet("成绩").WithLayout(layout.HeaderData{})).
		Extend("语文", "数学", "英语").
		Rows(
			examRow{Name: "张三", Extra: map[string]string{"语文": "90", "数学": "80", "英语": "85"}},
			examRow{Name: "李四", Extra: map[string]string{"语文": "70", "数学": "88", "英语": "92"}},
		).
		Apply(); err != nil {
		t.Fatal(err)
	}

	got, res, err := excelizex.Read[examRow](wb.Sheet("成绩").WithLayout(layout.HeaderData{})).
		Collect(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if res.HasErrors() {
		t.Fatalf("errors: %v", res.Errors())
	}
	if len(got) != 2 || got[0].Extra["英语"] != "85" || got[1].Extra["数学"] != "88" {
		t.Fatalf("rows: %+v", got)
	}
}

func TestStreamExportThenReimport(t *testing.T) {
	wb := excelizex.New()
	defer wb.Close()
	const n = 500
	if err := excelizex.Write[examRow](wb.Sheet("成绩").WithLayout(layout.HeaderData{})).
		Extend("分数").
		Stream(func(yield func(examRow) error) error {
			for i := 0; i < n; i++ {
				if err := yield(examRow{
					Name:  fmt.Sprintf("s%03d", i),
					Extra: map[string]string{"分数": fmt.Sprintf("%d", 60+i%40)},
				}); err != nil {
					return err
				}
			}

			return nil
		}).
		Apply(); err != nil {
		t.Fatal(err)
	}

	got, res, err := excelizex.Read[examRow](wb.Sheet("成绩").WithLayout(layout.HeaderData{})).
		Collect(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if res.HasErrors() || len(got) != n {
		t.Fatalf("len=%d errors=%v", len(got), res.Errors())
	}
	if got[0].Name != "s000" || got[n-1].Extra["分数"] != fmt.Sprintf("%d", 60+(n-1)%40) {
		t.Fatalf("first=%+v last=%+v", got[0], got[n-1])
	}
}

func TestGroupedHeaderImport(t *testing.T) {
	type row struct {
		Name    string `excel:"姓名"`
		IDCard  string `excel:"身份证"`
		Chinese int    `excel:"语文"`
		Math    int    `excel:"数学"`
	}
	lyt := layout.Grouped{
		HasNotice: true,
		Groups: []layout.Group{
			{Title: "考生", Span: 2},
			{Title: "成绩", Span: 2},
		},
	}
	wb := excelizex.New()
	defer wb.Close()
	if err := excelizex.Write[row](wb.Sheet("导入").WithLayout(lyt).WithNotice("分组表头导入")).
		Rows(row{Name: "张三", IDCard: "110101199001011234", Chinese: 91, Math: 88}).
		Apply(); err != nil {
		t.Fatal(err)
	}

	got, res, err := excelizex.Read[row](wb.Sheet("导入").WithLayout(layout.Stacked{HasNotice: true})).
		Collect(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if res.HasErrors() || len(got) != 1 || got[0].Math != 88 {
		t.Fatalf("rows=%+v errors=%v", got, res.Errors())
	}
}

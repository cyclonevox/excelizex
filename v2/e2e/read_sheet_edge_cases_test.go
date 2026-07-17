// 业务场景：notice 字段不强制 notice 行；错误 sheet 名返回清晰错误。
package e2e_test

import (
	"bytes"
	"context"
	"testing"

	excelizex "github.com/cyclonevox/excelizex/v2"
	"github.com/cyclonevox/excelizex/v2/e2e/fixture"
	"github.com/xuri/excelize/v2"
)

type rowWithNotice struct {
	Notice string `excel:"notice"`
	Name   string `excel:"姓名"`
}

func TestReadNoticeTagDoesNotRequireNoticeRow(t *testing.T) {
	f := excelize.NewFile()
	_ = f.SetSheetRow("Sheet1", "A2", &[]string{"姓名"})
	_ = f.SetSheetRow("Sheet1", "A3", &[]string{"张三"})
	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		t.Fatal(err)
	}

	wb := fixture.OpenBytes(t, &buf)
	defer wb.Close()

	rows, res, err := excelizex.Read[rowWithNotice](wb.Sheet("Sheet1")).
		Collect(context.Background())
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if len(rows) != 1 || rows[0].Name != "张三" {
		t.Fatalf("rows: %+v", rows)
	}
	if res.HasErrors() {
		t.Fatalf("unexpected errors: %v", res.Errors())
	}
}

func TestReadWrongSheetName(t *testing.T) {
	buf := fixture.BuildDirtyNoticeImport(t, [][]string{
		{"张三", "110101199001011234", "18", "A"},
	})
	wb := fixture.OpenBytes(t, buf)
	defer wb.Close()

	_, _, err := excelizex.Read[fixture.StudentImportRow](wb.Sheet("不存在")).
		Collect(context.Background())
	if err == nil {
		t.Fatal("expected error for missing sheet")
	}
}

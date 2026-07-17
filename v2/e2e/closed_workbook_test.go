// 业务场景：Workbook 关闭后 Collect/Apply/WriteErrors/Save 返回 workbook: closed，不 panic。
package e2e_test

import (
	"context"
	"strings"
	"testing"

	excelizex "github.com/cyclonevox/excelizex/v2"
	"github.com/cyclonevox/excelizex/v2/e2e/fixture"
	"github.com/cyclonevox/excelizex/v2/layout"
)

func TestClosedWorkbookOperations(t *testing.T) {
	buf := fixture.BuildDirtyNoticeImport(t, [][]string{
		{"张三", "", "18", "A"},
		{"", "", "19", "A"},
	})
	wb := fixture.OpenBytes(t, buf)

	_, res, err := excelizex.Read[fixture.StudentImportRow](wb.Sheet(fixture.SheetStudentImport)).
		Convert("grade", fixture.GradeImport).
		Validate(fixture.StructValidator()).
		Collect(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !res.HasErrors() {
		t.Fatal("expected read errors for WriteErrors test")
	}
	if err := wb.Close(); err != nil {
		t.Fatal(err)
	}

	_, _, err = excelizex.Read[fixture.StudentImportRow](wb.Sheet(fixture.SheetStudentImport)).
		Convert("grade", fixture.GradeImport).
		Collect(context.Background())
	if err == nil {
		t.Fatal("Collect after Close: expected error")
	}
	if !strings.Contains(err.Error(), "workbook: closed") {
		t.Fatalf("Collect after Close: got %v", err)
	}

	_, err = excelizex.Read[fixture.StudentImportRow](wb.Sheet(fixture.SheetStudentImport)).
		Convert("grade", fixture.GradeImport).
		Each(context.Background(), func(ctx excelizex.Context, row fixture.StudentImportRow) error {
			return nil
		})
	if err == nil {
		t.Fatal("Each after Close: expected error")
	}
	if !strings.Contains(err.Error(), "workbook: closed") {
		t.Fatalf("Each after Close: got %v", err)
	}

	err = excelizex.Write[fixture.ScoreRow](wb.Sheet("数据").WithLayout(layout.HeaderData{})).
		Rows(fixture.ScoreRow{Name: "李四", Score: 95}).
		Apply()
	if err == nil {
		t.Fatal("Apply after Close: expected error")
	}
	if !strings.Contains(err.Error(), "workbook: closed") {
		t.Fatalf("Apply after Close: got %v", err)
	}

	err = wb.WriteErrors(res)
	if err == nil {
		t.Fatal("WriteErrors after Close: expected error")
	}
	if !strings.Contains(err.Error(), "workbook: closed") {
		t.Fatalf("WriteErrors after Close: got %v", err)
	}

	var out strings.Builder
	err = wb.Save(&out)
	if err == nil {
		t.Fatal("Save after Close: expected error")
	}
	if !strings.Contains(err.Error(), "workbook: closed") {
		t.Fatalf("Save after Close: got %v", err)
	}
}

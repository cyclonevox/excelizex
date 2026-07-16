package excelizex_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	excelizex "github.com/cyclonevox/excelizex/v2"
	"github.com/cyclonevox/excelizex/v2/layout"
	"github.com/cyclonevox/excelizex/v2/validate"
)

func testdataPath(name string) string {
	return filepath.Join("testdata", name)
}

func gradeConverter(raw string) (any, error) {
	switch raw {
	case "A":
		return 1, nil
	case "B":
		return 2, nil
	default:
		return 0, fmt.Errorf("unknown grade %q", raw)
	}
}

func TestReadTestdataNoticeXLSX(t *testing.T) {
	f, err := os.Open(testdataPath("students_notice.xlsx"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	wb, err := excelizex.Open(f)
	if err != nil {
		t.Fatal(err)
	}
	defer wb.Close()

	rows, res, err := excelizex.Read[StudentRow](wb.Sheet("考生导入").WithLayout(layout.NoticeHeaderData{})).
		Convert("grade", gradeConverter).
		Validate(validate.Required{}).
		Collect(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].Name != "张三" || rows[0].Grade != 1 {
		t.Fatalf("rows: %+v", rows)
	}
	if len(res.Errors()) != 1 {
		t.Fatalf("errors: %v", res.Errors())
	}
}

func TestReadTestdataHeaderXLSX(t *testing.T) {
	f, err := os.Open(testdataPath("students_header.xlsx"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	wb, err := excelizex.Open(f)
	if err != nil {
		t.Fatal(err)
	}
	defer wb.Close()

	rows, res, err := excelizex.Read[StudentRow](wb.Sheet("考生导入").WithLayout(layout.HeaderData{})).
		Convert("grade", gradeConverter).
		Collect(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].Name != "王五" {
		t.Fatalf("rows: %+v", rows)
	}
	if res.HasErrors() {
		t.Fatalf("unexpected errors: %v", res.Errors())
	}
}

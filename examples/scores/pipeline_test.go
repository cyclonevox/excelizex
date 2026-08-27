package main

import (
	"context"
	"io"
	"path/filepath"
	"testing"

	excelizex "github.com/cyclonevox/excelizex"
)

func TestScoresPipeline(t *testing.T) {
	dir := t.TempDir()
	if err := run(dir, io.Discard); err != nil {
		t.Fatal(err)
	}

	wb, closer, err := openWorkbook(filepath.Join(dir, "3-汇总导出.xlsx"))
	if err != nil {
		t.Fatal(err)
	}
	defer closer()
	defer wb.Close()

	rows, res, err := excelizex.Read[scoreRow](wb.Sheet(sheetScores).
		WithLayout(scoreLayout())).
		Collect(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if res.HasErrors() {
		t.Fatalf("errors: %v", res.Errors())
	}
	if len(rows) != 2 || rows[0].Extra["语文"] != "90" || rows[1].Class != "3班" {
		t.Fatalf("rows: %+v", rows)
	}

	dvs, err := wb.File().GetDataValidations(sheetScores)
	if err != nil {
		t.Fatal(err)
	}
	if len(dvs) < 2 {
		t.Fatalf("expected grade and class validations, got %d", len(dvs))
	}
}

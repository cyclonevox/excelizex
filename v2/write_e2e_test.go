package excelizex_test

import (
	"bytes"
	"context"
	"fmt"
	"testing"
	"time"

	excelizex "github.com/cyclonevox/excelizex/v2"
	"github.com/cyclonevox/excelizex/v2/layout"
	"github.com/cyclonevox/excelizex/v2/validate"
	"github.com/xuri/excelize/v2"
)

type writeStudentRow struct {
	Notice string    `excel:"notice"`
	Name   string    `excel:"姓名" validate:"required" style:"header-red,body"`
	Age    int       `excel:"年龄" style:"header,body"`
	Active bool      `excel:"启用"`
	Joined time.Time `excel:"入学日期" time:"2006-01-02"`
	Grade  int       `excel:"等级" conv:"grade" style:"header,body-blue"`
	Class  string    `excel:"班级"`
}

func gradeExport(v any) (string, error) {
	switch n := v.(type) {
	case int:
		switch n {
		case 1:
			return "A", nil
		case 2:
			return "B", nil
		default:
			return "", fmt.Errorf("unknown grade %d", n)
		}
	default:
		return "", fmt.Errorf("bad grade type")
	}
}

func gradeImport(raw string) (any, error) {
	switch raw {
	case "A":
		return 1, nil
	case "B":
		return 2, nil
	default:
		return 0, fmt.Errorf("unknown grade %q", raw)
	}
}

func TestWriteReadRoundtrip(t *testing.T) {
	joined, _ := time.Parse("2006-01-02", "2024-09-01")
	rows := []writeStudentRow{
		{Name: "张三", Age: 18, Active: true, Joined: joined, Grade: 1, Class: "一班"},
		{Name: "李四", Age: 19, Active: false, Joined: joined, Grade: 2, Class: "二班"},
	}

	wb := excelizex.New()
	err := excelizex.Write[writeStudentRow](wb.Sheet("考生导入").
		WithLayout(layout.NoticeHeaderData{}).
		WithNotice("请按模板填写考生信息")).
		Convert("grade", gradeExport).
		Dropdown("班级", []string{"一班", "二班", "三班"}).
		Rows(rows...).
		Apply()
	if err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	if err := wb.Save(&buf); err != nil {
		t.Fatal(err)
	}

	wb2, err := excelizex.Open(&buf)
	if err != nil {
		t.Fatal(err)
	}
	defer wb2.Close()

	got, res, err := excelizex.Read[writeStudentRow](wb2.Sheet("考生导入")).
		Convert("grade", gradeImport).
		Validate(validate.Required{}).
		Collect(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Errors()) != 0 {
		t.Fatalf("unexpected read errors: %v", res.Errors())
	}
	if len(got) != 2 {
		t.Fatalf("rows: got %d want 2", len(got))
	}
	if got[0].Name != "张三" || got[0].Age != 18 || !got[0].Active || got[0].Grade != 1 || got[0].Class != "一班" {
		t.Fatalf("row0: %+v", got[0])
	}
	if got[1].Name != "李四" || got[1].Grade != 2 {
		t.Fatalf("row1: %+v", got[1])
	}
}

func TestWriteTemplateOnly(t *testing.T) {
	wb := excelizex.New()
	if err := excelizex.Write[writeStudentRow](wb.Sheet("模板").
		WithLayout(layout.NoticeHeaderData{}).
		WithNotice("空白模板")).
		Template().
		Apply(); err != nil {
		t.Fatal(err)
	}
	rows, err := wb.File().GetRows("模板")
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) < 2 {
		t.Fatalf("rows: %d", len(rows))
	}
	if rows[0][0] != "空白模板" {
		t.Fatalf("notice: %q", rows[0][0])
	}
	if rows[1][0] != "姓名" {
		t.Fatalf("header: %v", rows[1])
	}
}

func TestWriteDropdownValidation(t *testing.T) {
	wb := excelizex.New()
	if err := excelizex.Write[writeStudentRow](wb.Sheet("导入").
		WithLayout(layout.HeaderData{})).
		Dropdown("班级", []string{"一班", "二班"}).
		Template().
		Apply(); err != nil {
		t.Fatal(err)
	}
	dvs, err := wb.File().GetDataValidations("导入")
	if err != nil {
		t.Fatal(err)
	}
	if len(dvs) == 0 {
		t.Fatal("expected data validation")
	}
	found := false
	for _, dv := range dvs {
		if dv != nil && dv.Formula1 != "" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("validations: %+v", dvs)
	}
}

func TestWriteStyleApplied(t *testing.T) {
	wb := excelizex.New()
	if err := excelizex.Write[writeStudentRow](wb.Sheet("样式").
		WithLayout(layout.HeaderData{})).
		Template().
		Apply(); err != nil {
		t.Fatal(err)
	}
	f := wb.File()
	styleID, err := f.GetCellStyle("样式", "A1")
	if err != nil {
		t.Fatal(err)
	}
	if styleID == 0 {
		t.Fatal("expected header style on A1")
	}
	styleID, err = f.GetCellStyle("样式", "E2")
	if err != nil {
		t.Fatal(err)
	}
	if styleID == 0 {
		t.Fatal("expected body style on grade column")
	}
}

func TestWriteProtectSheet(t *testing.T) {
	wb := excelizex.New()
	pwd := "secret"
	if err := excelizex.Write[writeStudentRow](wb.Sheet("保护").
		WithLayout(layout.HeaderData{})).
		Template().
		Protect(pwd).
		Apply(); err != nil {
		t.Fatal(err)
	}
	if err := wb.File().UnprotectSheet("保护", pwd); err != nil {
		t.Fatalf("unprotect: %v", err)
	}
}

func TestWriteHeaderDataLayout(t *testing.T) {
	wb := excelizex.New()
	row := writeStudentRow{Name: "王五", Age: 20, Grade: 1}
	if err := excelizex.Write[writeStudentRow](wb.Sheet("数据").
		WithLayout(layout.HeaderData{})).
		Convert("grade", gradeExport).
		Rows(row).
		Apply(); err != nil {
		t.Fatal(err)
	}
	val, err := wb.File().GetCellValue("数据", "A2")
	if err != nil {
		t.Fatal(err)
	}
	if val != "王五" {
		t.Fatalf("name: %q", val)
	}
}

func TestExportToHelper(t *testing.T) {
	wb := excelizex.New()
	b := excelizex.Write[writeStudentRow](wb.Sheet("x"))
	excelizex.ExportTo(b, "grade", func(n int) (string, error) {
		return fmt.Sprintf("%d", n), nil
	})
	if b == nil {
		t.Fatal("nil builder")
	}
}

// Ensure excelize import is used for compile-time API checks in protect test.
var _ = excelize.NewFile

package testdata

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/xuri/excelize/v2"
)

// TestGenXLSX writes committed test fixtures (run once: go test ./testdata -run Gen -count=1).
func TestGenXLSX(t *testing.T) {
	if os.Getenv("GEN_TESTDATA") != "1" {
		t.Skip("set GEN_TESTDATA=1 to regenerate fixtures")
	}
	dir := "."
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}

	writeNoticeSheet(t, filepath.Join(dir, "students_notice.xlsx"))
	writeHeaderSheet(t, filepath.Join(dir, "students_header.xlsx"))
}

func writeNoticeSheet(t *testing.T, path string) {
	t.Helper()
	f := excelize.NewFile()
	const sheet = "考生导入"
	_ = f.SetSheetName("Sheet1", sheet)
	_ = f.SetCellStr(sheet, "A1", "请按模板填写考生信息")
	_ = f.SetSheetRow(sheet, "A2", &[]string{"姓名", "年龄", "等级"})
	_ = f.SetSheetRow(sheet, "A3", &[]string{"张三", "18", "A"})
	_ = f.SetSheetRow(sheet, "A4", &[]string{"李四", "bad", "B"})
	if err := f.SaveAs(path); err != nil {
		t.Fatal(err)
	}
}

func writeHeaderSheet(t *testing.T, path string) {
	t.Helper()
	f := excelize.NewFile()
	const sheet = "考生导入"
	_ = f.SetSheetName("Sheet1", sheet)
	_ = f.SetSheetRow(sheet, "A1", &[]string{"姓名", "年龄", "等级"})
	_ = f.SetSheetRow(sheet, "A2", &[]string{"王五", "20", "A"})
	if err := f.SaveAs(path); err != nil {
		t.Fatal(err)
	}
}

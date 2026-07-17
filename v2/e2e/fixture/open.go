package fixture

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	excelizex "github.com/cyclonevox/excelizex/v2"
)

// OpenBytes 从内存 buffer 打开工作簿。
func OpenBytes(t *testing.T, buf *bytes.Buffer) *excelizex.Workbook {
	t.Helper()
	wb, err := excelizex.Open(buf)
	if err != nil {
		t.Fatal(err)
	}

	return wb
}

// OpenTestdata 打开 e2e/testdata/ 下已提交的 .xlsx 夹具。
func OpenTestdata(t *testing.T, name string) *excelizex.Workbook {
	t.Helper()
	path, err := filepath.Abs(TestdataPath(name))
	if err != nil {
		t.Fatal(err)
	}
	wb, closeFn := OpenPath(t, path)
	t.Cleanup(closeFn)

	return wb
}

// OpenPath 从磁盘路径打开工作簿；返回 wb 与应在测试结束时调用的 close 函数。
func OpenPath(t *testing.T, path string) (*excelizex.Workbook, func()) {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	wb, err := excelizex.Open(f)
	if err != nil {
		_ = f.Close()
		t.Fatal(err)
	}
	closeFn := func() {
		_ = wb.Close()
		_ = f.Close()
	}

	return wb, closeFn
}

// SaveToBytes 将工作簿序列化到 buffer。
func SaveToBytes(t *testing.T, wb *excelizex.Workbook) *bytes.Buffer {
	t.Helper()
	var buf bytes.Buffer
	if err := wb.Save(&buf); err != nil {
		t.Fatal(err)
	}

	return &buf
}

// SaveWorkbookToPath 将工作簿保存到磁盘路径。
func SaveWorkbookToPath(t *testing.T, wb *excelizex.Workbook, path string) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := wb.Save(f); err != nil {
		t.Fatal(err)
	}
}

// RoundTripSave 保存后重新打开，模拟「导出 → 上传 → 再导入」。
func RoundTripSave(t *testing.T, wb *excelizex.Workbook) *excelizex.Workbook {
	t.Helper()
	buf := SaveToBytes(t, wb)
	wb2 := OpenBytes(t, buf)
	t.Cleanup(func() { _ = wb2.Close() })

	return wb2
}

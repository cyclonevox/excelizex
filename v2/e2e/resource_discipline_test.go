// 资源纪律：os.Open + defer f.Close + defer wb.Close；双 Close 安全。
package e2e_test

import (
	"os"
	"testing"

	"github.com/cyclonevox/excelizex/v2/e2e/fixture"
)

func TestResourceDisciplineOpenPath(t *testing.T) {
	src, err := os.Open(fixture.TestdataPath("students_notice_ok.xlsx"))
	if err != nil {
		t.Fatal(err)
	}
	defer src.Close()

	tmp, err := os.CreateTemp(t.TempDir(), "excelizex-e2e-*.xlsx")
	if err != nil {
		t.Fatal(err)
	}
	filePath := tmp.Name()
	if _, err := tmp.ReadFrom(src); err != nil {
		t.Fatal(err)
	}
	if err := tmp.Close(); err != nil {
		t.Fatal(err)
	}

	f, err := os.Open(filePath)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	wb, closeFn := fixture.OpenPath(t, filePath)
	closeFn()
	if err := wb.Close(); err != nil {
		t.Fatalf("second close: %v", err)
	}
}

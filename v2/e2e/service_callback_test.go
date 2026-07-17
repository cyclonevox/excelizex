// 业务场景：Each 回调模拟业务服务 svc.Create，服务错误记入 Result（真实导入任务风格）。
package e2e_test

import (
	"context"
	"fmt"
	"sync"
	"testing"

	excelizex "github.com/cyclonevox/excelizex/v2"
	"github.com/cyclonevox/excelizex/v2/e2e/fixture"
)

type fakeImportSvc struct {
	mu    sync.Mutex
	calls []string
}

func (s *fakeImportSvc) Create(ctx context.Context, row fixture.StudentImportRow) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.calls = append(s.calls, row.Name)
	if row.Name == "重复考生" {
		return fmt.Errorf("duplicate student %q", row.Name)
	}

	return nil
}

func (s *fakeImportSvc) callCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	return len(s.calls)
}

func TestServiceCallbackImportStyle(t *testing.T) {
	buf := fixture.BuildDirtyNoticeImport(t, [][]string{
		{"张三", "", "18", "A"},
		{"李四", "", "20", "B"},
		{"重复考生", "", "21", "A"},
		{"王五", "", "22", "B"},
	})
	wb := fixture.OpenBytes(t, buf)
	defer wb.Close()

	svc := &fakeImportSvc{}
	res, err := excelizex.Read[fixture.StudentImportRow](wb.Sheet(fixture.SheetStudentImport)).
		Convert("grade", fixture.GradeImport).
		Validate(fixture.StructValidator()).
		Each(context.Background(), func(ctx excelizex.Context, row fixture.StudentImportRow) error {
			return svc.Create(ctx, row)
		})
	if err != nil {
		t.Fatal(err)
	}
	if svc.callCount() != 4 {
		t.Fatalf("svc calls: %d want 4", svc.callCount())
	}
	if !res.HasErrors() {
		t.Fatal("expected service error in result")
	}
	errs := res.Errors()
	if len(errs) != 1 {
		t.Fatalf("errors: %v", errs)
	}
	if errs[0].Messages[0] != `duplicate student "重复考生"` {
		t.Fatalf("error msg: %v", errs[0].Messages)
	}
}

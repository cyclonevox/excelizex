// 业务场景：导入任务被取消（context.Cancel）时 Each 及时退出，不 hang。
package e2e_test

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	excelizex "github.com/cyclonevox/excelizex/v2"
	"github.com/cyclonevox/excelizex/v2/e2e/fixture"
	"github.com/cyclonevox/excelizex/v2/layout"
)

func TestCancelImportDuringEach(t *testing.T) {
	const n = 50
	data := make([][]string, n)
	for i := 0; i < n; i++ {
		data[i] = []string{fmt.Sprintf("考生%d", i+1), "", "25", "A"}
	}
	buf := fixture.BuildDirtyNoticeImport(t, data)
	wb := fixture.OpenBytes(t, buf)
	defer wb.Close()

	ctx, cancel := context.WithCancel(context.Background())
	var processed atomic.Int32
	go func() {
		time.Sleep(20 * time.Millisecond)
		cancel()
	}()

	_, err := excelizex.Read[fixture.StudentImportRow](wb.Sheet(fixture.SheetStudentImport).
		WithLayout(layout.NoticeHeaderData{})).
		Convert("grade", fixture.GradeImport).
		Validate(fixture.StructValidator()).
		Each(ctx, func(ctx excelizex.Context, row fixture.StudentImportRow) error {
			processed.Add(1)
			time.Sleep(5 * time.Millisecond)

			return nil
		}, excelizex.Concurrency(4))
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err: %v processed=%d", err, processed.Load())
	}
	if processed.Load() >= int32(n) {
		t.Fatalf("expected early stop, processed all %d rows", n)
	}
}

func TestCancelImportDuringCollect(t *testing.T) {
	wb := fixture.OpenTestdata(t, "students_notice_ok.xlsx")
	defer wb.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	rows, res, err := excelizex.Read[fixture.StudentImportRow](wb.Sheet(fixture.SheetStudentImport).
		WithLayout(layout.NoticeHeaderData{})).
		Convert("grade", fixture.GradeImport).
		Validate(fixture.StructValidator()).
		Collect(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err: %v", err)
	}
	if rows != nil {
		t.Fatalf("expected nil or empty partial rows on cancel, got %d", len(rows))
	}
	if res == nil {
		t.Fatal("expected result metadata")
	}
}

// 业务场景：大批量考生导入，Each 并发处理 ≥100 行，race 下无重复。
package e2e_test

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	excelizex "github.com/cyclonevox/excelizex/v2"
	"github.com/cyclonevox/excelizex/v2/e2e/fixture"
	"github.com/cyclonevox/excelizex/v2/layout"
)

func TestConcurrentBatchImport(t *testing.T) {
	const n = 120
	data := make([][]string, n)
	for i := 0; i < n; i++ {
		data[i] = []string{
			fmt.Sprintf("考生%d", i+1),
			fmt.Sprintf("1101011990%06d", i+1),
			fmt.Sprintf("%d", 20+i%10),
			"A",
		}
	}
	buf := fixture.BuildDirtyNoticeImport(t, data)
	wb := fixture.OpenBytes(t, buf)
	defer wb.Close()

	var seen sync.Map
	var count atomic.Int32
	_, err := excelizex.Read[fixture.StudentImportRow](wb.Sheet(fixture.SheetStudentImport).
		WithLayout(layout.NoticeHeaderData{})).
		Validate(fixture.StructValidator()).
		Each(context.Background(), func(ctx excelizex.Context, row fixture.StudentImportRow) error {
			if _, loaded := seen.LoadOrStore(row.Name, struct{}{}); loaded {
				t.Errorf("duplicate row %q", row.Name)
			}
			count.Add(1)
			time.Sleep(time.Microsecond)

			return nil
		}, excelizex.Concurrency(8))
	if err != nil {
		t.Fatal(err)
	}
	if count.Load() != n {
		t.Fatalf("processed %d want %d", count.Load(), n)
	}
}

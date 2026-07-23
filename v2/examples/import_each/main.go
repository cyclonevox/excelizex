// 业务场景：大批量导入，Each 并发调用业务服务 Create。
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"sync"

	excelizex "github.com/cyclonevox/excelizex/v2"
	"github.com/cyclonevox/excelizex/v2/examples/internal/demo"
	"github.com/cyclonevox/excelizex/v2/layout"
)

type importSvc struct {
	mu    sync.Mutex
	names []string
}

func (s *importSvc) Create(_ context.Context, row demo.StudentRow) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.names = append(s.names, row.Name)

	return nil
}

func main() {
	path := flag.String("file", "../../e2e/testdata/students_notice_ok.xlsx", "导入 xlsx 路径")
	concurrency := flag.Int("c", 4, "并发度")
	flag.Parse()

	f, err := os.Open(*path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "open: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	wb, err := excelizex.Open(f)
	if err != nil {
		fmt.Fprintf(os.Stderr, "workbook: %v\n", err)
		os.Exit(1)
	}
	defer wb.Close()

	svc := &importSvc{}
	res, err := excelizex.Read[demo.StudentRow](wb.Sheet(demo.SheetStudentImport).
		WithLayout(layout.NoticeHeaderData{}).
		WithNotice(demo.NoticeFillStudents)).
		Validate(demo.NewPlaygroundValidator()).
		Each(context.Background(), func(ctx excelizex.Context, row demo.StudentRow) error {
			return svc.Create(ctx, row)
		}, excelizex.Concurrency(*concurrency))
	if err != nil {
		fmt.Fprintf(os.Stderr, "each: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("服务入库 %d 行", len(svc.names))
	if res.HasErrors() {
		fmt.Printf("，%d 行失败", len(res.Errors()))
	}
	fmt.Println()
	for _, name := range svc.names {
		fmt.Printf("  created: %s\n", name)
	}
}

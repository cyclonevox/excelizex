// 业务场景：打开用户上传的导入表，Collect 聚合成功行，失败行 WriteErrors 回写。
package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	excelizex "github.com/cyclonevox/excelizex/v2"
	"github.com/cyclonevox/excelizex/v2/examples/internal/demo"
	"github.com/cyclonevox/excelizex/v2/layout"
)

func main() {
	path := flag.String("file", "../../e2e/testdata/students_notice_empty_name.xlsx", "导入 xlsx 路径")
	out := flag.String("out", "", "可选：将带错误列的文件保存到此路径")
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

	rows, res, err := excelizex.Read[demo.StudentRow](wb.Sheet(demo.SheetStudentImport).
		WithLayout(layout.NoticeHeaderData{}).
		WithNotice(demo.NoticeFillStudents)).
		Validate(demo.NewPlaygroundValidator()).
		Collect(context.Background())
	if err != nil {
		fmt.Fprintf(os.Stderr, "read: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("成功 %d 行，失败 %d 行\n", len(rows), len(res.Errors()))
	for _, row := range rows {
		fmt.Printf("  ok: %s age=%d grade=%d\n", row.Name, row.Age, row.Grade)
	}
	for _, e := range res.Errors() {
		fmt.Printf("  err row %d: %v\n", e.Row, e.Messages)
	}

	if !res.HasErrors() || *out == "" {
		return
	}
	if err := wb.WriteErrors(res); err != nil {
		fmt.Fprintf(os.Stderr, "write errors: %v\n", err)
		os.Exit(1)
	}
	outFile, err := os.Create(*out)
	if err != nil {
		fmt.Fprintf(os.Stderr, "create out: %v\n", err)
		os.Exit(1)
	}
	defer outFile.Close()
	if err := wb.Save(outFile); err != nil {
		fmt.Fprintf(os.Stderr, "save: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("已回写错误列到 %s\n", *out)
}

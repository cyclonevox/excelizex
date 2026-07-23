// 业务场景：将已有考生数据导出为 Excel（Rows 写数据区）。
package main

import (
	"flag"
	"fmt"
	"os"

	excelizex "github.com/cyclonevox/excelizex/v2"
	"github.com/cyclonevox/excelizex/v2/examples/internal/demo"
	"github.com/cyclonevox/excelizex/v2/layout"
)

func main() {
	out := flag.String("out", "考生导出.xlsx", "输出路径")
	flag.Parse()

	rows := []demo.StudentRow{
		{Name: "张三", IDCard: "110101199001011234", Age: 18, Grade: 1},
		{Name: "李四", IDCard: "110101199002021234", Age: 20, Grade: 2},
	}

	wb := excelizex.New()
	defer wb.Close()

	if err := excelizex.Write[demo.StudentRow](wb.Sheet(demo.SheetStudentImport).
		WithLayout(layout.NoticeHeaderData{}).
		WithNotice(demo.NoticeFillStudents)).
		Dropdown("年级", []string{"A", "B"}).
		Rows(rows...).
		Apply(); err != nil {
		fmt.Fprintf(os.Stderr, "write: %v\n", err)
		os.Exit(1)
	}

	f, err := os.Create(*out)
	if err != nil {
		fmt.Fprintf(os.Stderr, "create: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()
	if err := wb.Save(f); err != nil {
		fmt.Fprintf(os.Stderr, "save: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("已导出 %d 行到 %s\n", len(rows), *out)
}

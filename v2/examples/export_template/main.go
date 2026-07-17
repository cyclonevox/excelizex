// 业务场景：生成带提示、下拉与保护的导入模板，下发给业务方填写。
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
	out := flag.String("out", "考生模板.xlsx", "输出路径")
	password := flag.String("password", "import-secret", "工作表保护密码")
	flag.Parse()

	wb := excelizex.New()
	defer wb.Close()

	if err := excelizex.Write[demo.StudentRow](wb.Sheet(demo.SheetStudentImport).
		WithLayout(layout.NoticeHeaderData{}).
		WithNotice(demo.NoticeFillStudents)).
		Convert("grade", demo.GradeExport).
		Dropdown("年级", []string{"A", "B"}).
		Protect(*password).
		Template().
		Apply(); err != nil {
		fmt.Fprintf(os.Stderr, "write template: %v\n", err)
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

	fmt.Printf("模板已保存到 %s（保护密码 %q）\n", *out, *password)
}

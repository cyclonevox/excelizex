package excelizex_test

import (
	"context"
	"fmt"
	"os"

	excelizex "github.com/cyclonevox/excelizex/v2"
	"github.com/cyclonevox/excelizex/v2/layout"
	"github.com/cyclonevox/excelizex/v2/validate"
)

// StudentRow is a typical import DTO.
type StudentRow struct {
	Name  string `excel:"姓名" validate:"required"`
	Age   int    `excel:"年龄"`
	Grade int    `excel:"等级" conv:"grade"`
}

// Example demonstrates a short business import flow (~20 lines).
func ExampleRead() {
	f, err := os.Open("testdata/students_notice.xlsx")
	if err != nil {
		fmt.Println("open:", err)

		return
	}
	defer f.Close()

	wb, err := excelizex.Open(f)
	if err != nil {
		fmt.Println("workbook:", err)

		return
	}
	defer wb.Close()

	rows, res, err := excelizex.Read[StudentRow](wb.Sheet("考生导入").WithLayout(layout.NoticeHeaderData{})).
		Convert("grade", func(raw string) (any, error) {
			switch raw {
			case "A":
				return 1, nil
			case "B":
				return 2, nil
			default:
				return 0, fmt.Errorf("unknown grade %q", raw)
			}
		}).
		Validate(validate.Required{}).
		Collect(context.Background())
	if err != nil {
		fmt.Println("read:", err)

		return
	}

	fmt.Printf("imported %d rows, %d errors\n", len(rows), len(res.Errors()))
	for _, row := range rows {
		fmt.Printf("%s age=%d grade=%d\n", row.Name, row.Age, row.Grade)
	}
	// Output:
	// imported 1 rows, 1 errors
	// 张三 age=18 grade=1
}

// ExampleWrite demonstrates template generation and data export.
func ExampleWrite() {
	wb := excelizex.New()
	rows := []StudentRow{{Name: "张三", Age: 18, Grade: 1}}
	if err := excelizex.Write[StudentRow](wb.Sheet("考生导入").
		WithLayout(layout.NoticeHeaderData{}).
		WithNotice("请填写考生信息")).
		Convert("grade", func(v any) (string, error) {
			switch n := v.(type) {
			case int:
				if n == 1 {
					return "A", nil
				}
			}

			return "", fmt.Errorf("bad grade")
		}).
		Dropdown("等级", []string{"A", "B"}).
		Rows(rows...).
		Apply(); err != nil {
		fmt.Println("write:", err)

		return
	}
	fmt.Println("written ok")
	// Output:
	// written ok
}

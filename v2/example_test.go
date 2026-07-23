package excelizex_test

import (
	"bytes"
	"context"
	"fmt"

	excelizex "github.com/cyclonevox/excelizex/v2"
	"github.com/cyclonevox/excelizex/v2/layout"
	playvalidator "github.com/go-playground/validator/v10"
)

type examplePlaygroundValidator struct {
	v *playvalidator.Validate
}

func newExamplePlaygroundValidator() examplePlaygroundValidator {
	return examplePlaygroundValidator{v: playvalidator.New()}
}

func (p examplePlaygroundValidator) Validate(row any) error {
	return p.v.Struct(row)
}

// ExampleRead demonstrates Write → buffer → Open → Read (no committed xlsx).
func ExampleRead() {
	wb := excelizex.New()
	defer wb.Close()
	if err := excelizex.Write[StudentRow](wb.Sheet("考生导入").
		WithLayout(layout.NoticeHeaderData{}).
		WithNotice("请按模板填写考生信息")).
		Rows(StudentRow{Name: "张三", Age: 18, Grade: 1}).
		Apply(); err != nil {
		fmt.Println("write:", err)

		return
	}
	// 模拟用户继续填写一行非法年龄
	_ = wb.File().SetSheetRow("考生导入", "A4", &[]string{"李四", "bad", "B"})

	var buf bytes.Buffer
	if err := wb.Save(&buf); err != nil {
		fmt.Println("save:", err)

		return
	}
	wb2, err := excelizex.Open(&buf)
	if err != nil {
		fmt.Println("open:", err)

		return
	}
	defer wb2.Close()

	rows, res, err := excelizex.Read[StudentRow](wb2.Sheet("考生导入").WithLayout(layout.NoticeHeaderData{})).
		Validate(newExamplePlaygroundValidator()).
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
	if wb != nil {
		defer wb.Close()
	}
	rows := []StudentRow{{Name: "张三", Age: 18, Grade: 1}}
	if err := excelizex.Write[StudentRow](wb.Sheet("考生导入").
		WithLayout(layout.NoticeHeaderData{}).
		WithNotice("请填写考生信息")).
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

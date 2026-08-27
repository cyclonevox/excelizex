package excelizex_test

import (
	"bytes"
	"context"
	"fmt"

	excelizex "github.com/cyclonevox/excelizex"
	"github.com/cyclonevox/excelizex/layout"
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

type exampleScoreRow struct {
	Name  string            `excel:"姓名"`
	Extra map[string]string `excel:",extend"`
}

type exampleRegionRow struct {
	Name     string `excel:"姓名"`
	Province string `excel:"省"`
	City     string `excel:"市"`
}

type exampleGroupedRow struct {
	Name    string `excel:"姓名"`
	Age     int    `excel:"年龄"`
	Chinese int    `excel:"语文"`
	Math    int    `excel:"数学"`
}

// ExampleWrite_extend shows dynamic extra columns via excel:",extend".
func ExampleWrite_extend() {
	wb := excelizex.New()
	defer wb.Close()
	if err := excelizex.Write[exampleScoreRow](wb.Sheet("成绩").WithLayout(layout.HeaderData{})).
		Extend("语文", "数学").
		Rows(exampleScoreRow{Name: "张三", Extra: map[string]string{"语文": "90", "数学": "80"}}).
		Apply(); err != nil {
		fmt.Println("write:", err)

		return
	}
	rows, _, err := excelizex.Read[exampleScoreRow](wb.Sheet("成绩").WithLayout(layout.HeaderData{})).
		Collect(context.Background())
	if err != nil {
		fmt.Println("read:", err)

		return
	}
	fmt.Printf("%s 语文=%s 数学=%s\n", rows[0].Name, rows[0].Extra["语文"], rows[0].Extra["数学"])
	// Output:
	// 张三 语文=90 数学=80
}

// ExampleWrite_stream shows streaming export without holding all rows on the builder.
func ExampleWrite_stream() {
	wb := excelizex.New()
	defer wb.Close()
	src := []exampleScoreRow{
		{Name: "张三", Extra: map[string]string{"语文": "90"}},
		{Name: "李四", Extra: map[string]string{"语文": "88"}},
	}
	if err := excelizex.Write[exampleScoreRow](wb.Sheet("成绩").WithLayout(layout.HeaderData{})).
		Extend("语文").
		Stream(func(yield func(exampleScoreRow) error) error {
			for _, row := range src {
				if err := yield(row); err != nil {
					return err
				}
			}

			return nil
		}).
		Apply(); err != nil {
		fmt.Println("write:", err)

		return
	}
	rows, _, err := excelizex.Read[exampleScoreRow](wb.Sheet("成绩").WithLayout(layout.HeaderData{})).
		Collect(context.Background())
	if err != nil {
		fmt.Println("read:", err)

		return
	}
	fmt.Printf("streamed %d rows\n", len(rows))
	// Output:
	// streamed 2 rows
}

// ExampleWrite_cascade shows a parent/child dropdown (省 → 市).
func ExampleWrite_cascade() {
	wb := excelizex.New()
	defer wb.Close()
	if err := excelizex.Write[exampleRegionRow](wb.Sheet("地区").WithLayout(layout.HeaderData{})).
		Cascade("市", "省", map[string][]string{
			"浙江": {"杭州", "宁波"},
			"江苏": {"南京", "苏州"},
		}).
		Template().
		Apply(); err != nil {
		fmt.Println("write:", err)

		return
	}
	fmt.Println("cascade template ok")
	// Output:
	// cascade template ok
}

// ExampleWrite_grouped shows a two-row grouped header.
func ExampleWrite_grouped() {
	wb := excelizex.New()
	defer wb.Close()
	lyt := layout.Grouped{
		Groups: []layout.Group{
			{Title: "基本信息", Span: 2},
			{Title: "成绩", Span: 2},
		},
	}
	if err := excelizex.Write[exampleGroupedRow](wb.Sheet("成绩").WithLayout(lyt)).
		Rows(exampleGroupedRow{Name: "张三", Age: 18, Chinese: 90, Math: 88}).
		Apply(); err != nil {
		fmt.Println("write:", err)

		return
	}
	rows, _, err := excelizex.Read[exampleGroupedRow](wb.Sheet("成绩").WithLayout(lyt)).
		Collect(context.Background())
	if err != nil {
		fmt.Println("read:", err)

		return
	}
	fmt.Printf("%s 语文=%d\n", rows[0].Name, rows[0].Chinese)
	// Output:
	// 张三 语文=90
}

package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	excelizex "github.com/cyclonevox/excelizex"
	"github.com/cyclonevox/excelizex/layout"
)

func scoreLayout() layout.Grouped {
	return layout.Grouped{
		HasNotice: true,
		Groups: []layout.Group{
			{Title: "考生", Span: 3},
			{Title: "成绩", Span: 3},
		},
	}
}

func run(dir string, log io.Writer) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	tpl := filepath.Join(dir, "1-成绩模板.xlsx")
	filled := filepath.Join(dir, "2-已填成绩.xlsx")
	summary := filepath.Join(dir, "3-汇总导出.xlsx")

	logf(log, "① 下发成绩模板（分组表头 + 年级→班级两级下拉 + 动态科目列）")
	if err := writeTemplate(tpl); err != nil {
		return err
	}
	logf(log, "    → %s", tpl)

	logf(log, "② 班主任按科目填分后交回")
	if err := fillScores(tpl, filled); err != nil {
		return err
	}
	logf(log, "    → %s", filled)

	logf(log, "③ 读入动态科目列，再用 Stream 汇总导出")
	rows, err := importScores(filled)
	if err != nil {
		return err
	}
	for _, row := range rows {
		logf(log, "      %s %s/%s 语文=%s 数学=%s 英语=%s",
			row.Name, row.Grade, row.Class,
			row.Extra["语文"], row.Extra["数学"], row.Extra["英语"])
	}
	if err := exportSummary(rows, summary); err != nil {
		return err
	}
	logf(log, "    → %s（%d 人）", summary, len(rows))

	return nil
}

func writeTemplate(path string) error {
	wb := excelizex.New()
	defer wb.Close()
	if err := writeScores(wb).Template().Apply(); err != nil {
		return err
	}

	return saveWorkbook(wb, path)
}

func fillScores(_, out string) error {
	wb := excelizex.New()
	defer wb.Close()
	rows := []scoreRow{
		{Name: "张三", Grade: "高一", Class: "1班", Extra: map[string]string{"语文": "90", "数学": "88", "英语": "85"}},
		{Name: "李四", Grade: "高二", Class: "3班", Extra: map[string]string{"语文": "76", "数学": "92", "英语": "80"}},
	}
	if err := writeScores(wb).Rows(rows...).Apply(); err != nil {
		return err
	}

	return saveWorkbook(wb, out)
}

func importScores(path string) ([]scoreRow, error) {
	wb, closer, err := openWorkbook(path)
	if err != nil {
		return nil, err
	}
	defer closer()
	defer wb.Close()

	rows, res, err := excelizex.Read[scoreRow](wb.Sheet(sheetScores).
		WithLayout(scoreLayout()).
		WithNotice(noticeText)).
		Collect(context.Background())
	if err != nil {
		return nil, err
	}
	if res.HasErrors() {
		return nil, fmt.Errorf("import errors: %v", res.Errors())
	}

	return rows, nil
}

func exportSummary(rows []scoreRow, path string) error {
	wb := excelizex.New()
	defer wb.Close()
	if err := writeScores(wb).
		Stream(func(yield func(scoreRow) error) error {
			for _, row := range rows {
				if err := yield(row); err != nil {
					return err
				}
			}

			return nil
		}).
		Apply(); err != nil {
		return err
	}

	return saveWorkbook(wb, path)
}

func writeScores(wb *excelizex.Workbook) *excelizex.WriteBuilder[scoreRow] {
	return excelizex.Write[scoreRow](wb.Sheet(sheetScores).
		WithLayout(scoreLayout()).
		WithNotice(noticeText)).
		Extend(subjects...).
		Cascade("班级", "年级", gradeClass)
}

func openWorkbook(path string) (*excelizex.Workbook, func(), error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	wb, err := excelizex.Open(f)
	if err != nil {
		_ = f.Close()

		return nil, nil, err
	}

	return wb, func() { _ = f.Close() }, nil
}

func saveWorkbook(wb *excelizex.Workbook, path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	return wb.Save(f)
}

func logf(w io.Writer, format string, args ...any) {
	if w == nil {
		return
	}
	fmt.Fprintf(w, format+"\n", args...)
}

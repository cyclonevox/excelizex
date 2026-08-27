package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	excelizex "github.com/cyclonevox/excelizex"
	"github.com/cyclonevox/excelizex/layout"
)

func run(dir string, log io.Writer) (*enrollment, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	svc := newEnrollment()
	tpl := filepath.Join(dir, "1-考生模板.xlsx")
	filled := filepath.Join(dir, "2-班主任交回.xlsx")
	returned := filepath.Join(dir, "3-退回-带错误列.xlsx")
	fixed := filepath.Join(dir, "4-班主任改完.xlsx")
	roster := filepath.Join(dir, "5-已录取名单.xlsx")

	logf(log, "① 教务处下发导入模板（提示行 + 表头 + 年级下拉）")
	if err := writeTemplate(tpl); err != nil {
		return nil, err
	}
	logf(log, "    → %s", tpl)

	logf(log, "② 班主任按模板填写后交回（故意留下 3 处填错）")
	if err := fillAsTeacher(tpl, filled); err != nil {
		return nil, err
	}
	logf(log, "    → %s", filled)

	logf(log, "③ 系统导入：校验 + 并发写入报名库，失败行回写成错误列")
	res, err := importFile(filled, svc, 4)
	if err != nil {
		return nil, err
	}
	logImport(log, svc, res)
	if err := writeErrorFile(filled, res, returned); err != nil {
		return nil, err
	}
	logf(log, "    → 退回 %s", returned)

	logf(log, "④ 班主任按错误列改退回表（这份表里只剩失败行）")
	if err := fixReturned(returned, fixed); err != nil {
		return nil, err
	}
	logf(log, "    → %s", fixed)

	logf(log, "⑤ 再导改完的失败行；已录取的人仍在报名库里")
	res2, err := importFile(fixed, svc, 4)
	if err != nil {
		return nil, err
	}
	logImport(log, svc, res2)

	logf(log, "⑥ 导出已录取名单")
	if err := exportRoster(svc.List(), roster); err != nil {
		return nil, err
	}
	logf(log, "    → %s（%d 人）", roster, svc.Count())
	return svc, nil
}

func writeTemplate(path string) error {
	wb := excelizex.New()
	defer wb.Close()
	if err := excelizex.Write[student](wb.Sheet(sheetImport).
		WithLayout(layout.NoticeHeaderData{}).
		WithNotice(noticeText)).
		Dropdown("年级", []string{"A", "B"}).
		Template().
		Apply(); err != nil {
		return err
	}
	return saveWorkbook(wb, path)
}

func fillAsTeacher(templatePath, out string) error {
	wb, closer, err := openWorkbook(templatePath)
	if err != nil {
		return err
	}
	defer closer()
	defer wb.Close()

	rows := []student{
		{Name: "张三", IDCard: "110101199001011234", Age: 18, Grade: 1},
		{Name: "李四", IDCard: "110101199002021234", Age: 19, Grade: 2},
		{Name: "", IDCard: "110101199003031234", Age: 17, Grade: 1}, // 缺姓名
		{Name: "王五", IDCard: "短", Age: 20, Grade: 1},                // 身份证不够 18 位
		{Name: "赵六", IDCard: "110101199004041234", Age: 16, Grade: 1},
	}
	if err := excelizex.Write[student](wb.Sheet(sheetImport).
		WithLayout(layout.NoticeHeaderData{}).
		WithNotice(noticeText)).
		Dropdown("年级", []string{"A", "B"}).
		Rows(rows...).
		Apply(); err != nil {
		return err
	}
	// 用户绕过下拉，把年级填成 C。
	f := wb.File()
	if err := f.SetCellValue(sheetImport, "A8", "钱七"); err != nil {
		return err
	}
	if err := f.SetCellValue(sheetImport, "B8", "110101199005051234"); err != nil {
		return err
	}
	if err := f.SetCellValue(sheetImport, "C8", 21); err != nil {
		return err
	}
	if err := f.SetCellValue(sheetImport, "D8", "C"); err != nil {
		return err
	}
	return saveWorkbook(wb, out)
}

func importFile(path string, svc *enrollment, workers int) (*excelizex.Result, error) {
	wb, closer, err := openWorkbook(path)
	if err != nil {
		return nil, err
	}
	defer closer()
	defer wb.Close()

	return excelizex.Read[student](wb.Sheet(sheetImport).
		WithLayout(layout.NoticeHeaderData{}).
		WithNotice(noticeText)).
		Validate(newValidator()).
		Each(context.Background(), func(ctx excelizex.Context, row student) error {
			return svc.Create(ctx, row)
		}, excelizex.Concurrency(workers))
}

func writeErrorFile(src string, res *excelizex.Result, out string) error {
	if res == nil || !res.HasErrors() {
		return fmt.Errorf("expected row errors to rewrite")
	}
	wb, closer, err := openWorkbook(src)
	if err != nil {
		return err
	}
	defer closer()
	defer wb.Close()
	if err := wb.WriteErrors(res); err != nil {
		return err
	}
	return saveWorkbook(wb, out)
}

func fixReturned(src, out string) error {
	wb, closer, err := openWorkbook(src)
	if err != nil {
		return err
	}
	defer closer()
	defer wb.Close()

	f := wb.File()
	rows, err := f.GetRows(sheetImport)
	if err != nil {
		return err
	}
	// WriteErrors 只留下失败行，从数据起始行开始改格子。
	for i := 2; i < len(rows); i++ {
		cells := rows[i]
		name, id, grade := cell(cells, 0), cell(cells, 1), cell(cells, 3)
		rowNum := i + 1
		if strings.TrimSpace(name) == "" {
			if err := f.SetCellValue(sheetImport, fmt.Sprintf("A%d", rowNum), "周九"); err != nil {
				return err
			}
		}
		if len(id) != 18 {
			if err := f.SetCellValue(sheetImport, fmt.Sprintf("B%d", rowNum), "110101199006061234"); err != nil {
				return err
			}
		}
		if grade == "C" {
			if err := f.SetCellValue(sheetImport, fmt.Sprintf("D%d", rowNum), "B"); err != nil {
				return err
			}
		}
	}
	return saveWorkbook(wb, out)
}

func cell(row []string, i int) string {
	if i >= len(row) {
		return ""
	}
	return row[i]
}

func exportRoster(rows []student, path string) error {
	wb := excelizex.New()
	defer wb.Close()
	if err := excelizex.Write[student](wb.Sheet(sheetImport).
		WithLayout(layout.NoticeHeaderData{}).
		WithNotice("已录取考生名单")).
		Dropdown("年级", []string{"A", "B"}).
		Rows(rows...).
		Apply(); err != nil {
		return err
	}
	return saveWorkbook(wb, path)
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

func logImport(log io.Writer, svc *enrollment, res *excelizex.Result) {
	fail := 0
	if res != nil {
		fail = len(res.Errors())
	}
	logf(log, "    报名库 %d 人，本轮失败 %d 行", svc.Count(), fail)
	for _, row := range svc.List() {
		logf(log, "      录取 %s  %s  年级=%d", row.Name, row.IDCard, row.Grade)
	}
	if res == nil {
		return
	}
	for _, e := range res.Errors() {
		logf(log, "      失败 第 %d 行  %v", e.Row, e.Messages)
	}
}

func logf(w io.Writer, format string, args ...any) {
	if w == nil {
		return
	}
	fmt.Fprintf(w, format+"\n", args...)
}

package fixture

import (
	"bytes"
	"path/filepath"
	"testing"
	"time"

	excelizex "github.com/cyclonevox/excelizex/v2"
	"github.com/cyclonevox/excelizex/v2/layout"
	"github.com/xuri/excelize/v2"
)

const (
	SheetStudentImport = "考生导入"
	NoticeFillStudents = "请按模板填写考生信息"
)

// BuildDirtyNoticeImport 用 excelize 模拟用户上传的「脏」导入表（提示行 + 表头 + 数据）。
// 静态场景请用 OpenTestdata；本函数保留给动态行数或改写后重导等测试。
func BuildDirtyNoticeImport(t *testing.T, dataRows [][]string) *bytes.Buffer {
	t.Helper()
	f := excelize.NewFile()
	sheet := SheetStudentImport
	idx, _ := f.GetSheetIndex(sheet)
	if idx == -1 {
		_, _ = f.NewSheet(sheet)
		_ = f.DeleteSheet("Sheet1")
	}
	_ = f.SetCellStr(sheet, "A1", NoticeFillStudents)
	_ = f.SetSheetRow(sheet, "A2", &[]string{"姓名", "身份证", "年龄", "年级"})
	for i, row := range dataRows {
		addr, _ := excelize.JoinCellName("A", 3+i)
		_ = f.SetSheetRow(sheet, addr, &row)
	}
	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		t.Fatal(err)
	}

	return &buf
}

// BuildReorderedHeadersFile 用户拖乱表头并多加无关列、备注列。静态场景请用 OpenTestdata("students_reordered.xlsx")。
func BuildReorderedHeadersFile(t *testing.T) *bytes.Buffer {
	t.Helper()
	f := excelize.NewFile()
	sheet := SheetStudentImport
	_, _ = f.NewSheet(sheet)
	_ = f.DeleteSheet("Sheet1")
	_ = f.SetCellStr(sheet, "A1", "说明文字可不同")
	_ = f.SetSheetRow(sheet, "A2", &[]string{"年级", "无关列", "姓名", "年龄", "备注"})
	_ = f.SetSheetRow(sheet, "A3", &[]string{"A", "ignored", "张三", "30", "ok"})
	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		t.Fatal(err)
	}

	return &buf
}

// BuildMissingColumnFile 缺列场景：只有姓名列。静态场景请用 OpenTestdata("students_missing_column.xlsx")。
func BuildMissingColumnFile(t *testing.T) *bytes.Buffer {
	t.Helper()
	f := excelize.NewFile()
	_ = f.SetCellStr("Sheet1", "A1", "n")
	_ = f.SetSheetRow("Sheet1", "A2", &[]string{"姓名"})
	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		t.Fatal(err)
	}

	return &buf
}

// BuildInlineAddressFile 嵌套 inline 地址导入表。静态场景请用 OpenTestdata("inline_address.xlsx")。
func BuildInlineAddressFile(t *testing.T) *bytes.Buffer {
	t.Helper()
	f := excelize.NewFile()
	sheet := "地址导入"
	_, _ = f.NewSheet(sheet)
	_ = f.DeleteSheet("Sheet1")
	_ = f.SetCellStr(sheet, "A1", "请填写地址")
	_ = f.SetSheetRow(sheet, "A2", &[]string{"姓名", "城市", "街道"})
	_ = f.SetSheetRow(sheet, "A3", &[]string{"张三", "北京", "长安街"})
	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		t.Fatal(err)
	}

	return &buf
}

// BuildDirtyEmptyRowsFile 数据区中间夹空行。静态场景请用 OpenTestdata("students_empty_rows.xlsx")。
func BuildDirtyEmptyRowsFile(t *testing.T) *bytes.Buffer {
	t.Helper()
	f := excelize.NewFile()
	sheet := SheetStudentImport
	_, _ = f.NewSheet(sheet)
	_ = f.DeleteSheet("Sheet1")
	_ = f.SetCellStr(sheet, "A1", NoticeFillStudents)
	_ = f.SetSheetRow(sheet, "A2", &[]string{"姓名", "身份证", "年龄", "年级"})
	_ = f.SetSheetRow(sheet, "A3", &[]string{"张三", "110101199001011234", "18", "A"})
	_ = f.SetSheetRow(sheet, "A4", &[]string{"", "", ""})
	_ = f.SetSheetRow(sheet, "A5", &[]string{"  ", "\t", ""})
	_ = f.SetSheetRow(sheet, "A6", &[]string{"李四", "110101199002021234", "20", "B"})
	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		t.Fatal(err)
	}

	return &buf
}

// WriteStudentBatch 用 excelizex.Write 生成标准考生导入表（库 round-trip）。
func WriteStudentBatch(t *testing.T, rows ...StudentImportRow) *bytes.Buffer {
	t.Helper()
	wb := excelizex.New()
	t.Cleanup(func() { _ = wb.Close() })
	if err := excelizex.Write[StudentImportRow](wb.Sheet(SheetStudentImport).
		WithLayout(layout.NoticeHeaderData{}).
		WithNotice(NoticeFillStudents)).
		Rows(rows...).
		Apply(); err != nil {
		t.Fatal(err)
	}

	return SaveToBytes(t, wb)
}

// WriteHeaderDataScores 无 notice 的 HeaderData 导出。
func WriteHeaderDataScores(t *testing.T, rows ...ScoreRow) *bytes.Buffer {
	t.Helper()
	wb := excelizex.New()
	t.Cleanup(func() { _ = wb.Close() })
	if err := excelizex.Write[ScoreRow](wb.Sheet("成绩").WithLayout(layout.HeaderData{})).
		Rows(rows...).
		Apply(); err != nil {
		t.Fatal(err)
	}

	return SaveToBytes(t, wb)
}

// WriteTimeBoolSheet 时间 + 布尔字段 round-trip 表。
func WriteTimeBoolSheet(t *testing.T, rows ...TimeBoolRow) *bytes.Buffer {
	t.Helper()
	wb := excelizex.New()
	t.Cleanup(func() { _ = wb.Close() })
	if err := excelizex.Write[TimeBoolRow](wb.Sheet(SheetStudentImport).
		WithLayout(layout.NoticeHeaderData{}).
		WithNotice("请填写启用状态与入学日期")).
		Rows(rows...).
		Apply(); err != nil {
		t.Fatal(err)
	}

	return SaveToBytes(t, wb)
}

// WriteTemplateDistribute 生成带下拉与保护的模板工作簿，保存到 dir 并返回路径。
func WriteTemplateDistribute(t *testing.T, dir string) string {
	t.Helper()
	wb := excelizex.New()
	t.Cleanup(func() { _ = wb.Close() })
	const pwd = "dist-secret"
	if err := excelizex.Write[TemplateDistributeRow](wb.Sheet(SheetStudentImport).
		WithLayout(layout.NoticeHeaderData{}).
		WithNotice("请从下拉选择年级后填写")).
		Dropdown("年级", []string{"A", "B"}).
		Protect(pwd).
		Template().
		Apply(); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "考生模板.xlsx")
	SaveWorkbookToPath(t, wb, path)

	return path
}

// FillTemplateRows 在已打开的工作簿上写入数据行（模拟业务方填表）。
func FillTemplateRows(t *testing.T, wb *excelizex.Workbook, rows [][]string) {
	t.Helper()
	f := wb.File()
	for i, row := range rows {
		addr, _ := excelize.JoinCellName("A", 3+i)
		if err := f.SetSheetRow(SheetStudentImport, addr, &row); err != nil {
			t.Fatal(err)
		}
	}
}

// ParseJoined 解析入学日期测试辅助。
func ParseJoined(s string) time.Time {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		panic(err)
	}

	return t
}

// legacyNoticeStudents 复刻原 students_notice.xlsx：张三 ok + 李四 bad age。
func legacyNoticeStudents() *bytes.Buffer {
	f := excelize.NewFile()
	sheet := SheetStudentImport
	_ = f.SetSheetName("Sheet1", sheet)
	_ = f.SetCellStr(sheet, "A1", NoticeFillStudents)
	_ = f.SetSheetRow(sheet, "A2", &[]string{"姓名", "年龄", "等级"})
	_ = f.SetSheetRow(sheet, "A3", &[]string{"张三", "18", "A"})
	_ = f.SetSheetRow(sheet, "A4", &[]string{"李四", "bad", "B"})
	var buf bytes.Buffer
	_ = f.Write(&buf)

	return &buf
}

// legacyHeaderStudents 复刻原 students_header.xlsx：无 notice，王五单行。
func legacyHeaderStudents() *bytes.Buffer {
	f := excelize.NewFile()
	sheet := SheetStudentImport
	_ = f.SetSheetName("Sheet1", sheet)
	_ = f.SetSheetRow(sheet, "A1", &[]string{"姓名", "年龄", "等级"})
	_ = f.SetSheetRow(sheet, "A2", &[]string{"王五", "20", "A"})
	var buf bytes.Buffer
	_ = f.Write(&buf)

	return &buf
}

// BuildLegacyNoticeStudents 原 students_notice.xlsx 等价表。静态场景请用 OpenTestdata("students_notice_legacy.xlsx")。
func BuildLegacyNoticeStudents(t *testing.T) *bytes.Buffer {
	t.Helper()

	return legacyNoticeStudents()
}

// BuildLegacyHeaderStudents 原 students_header.xlsx 等价表。静态场景请用 OpenTestdata("students_header_legacy.xlsx")。
func BuildLegacyHeaderStudents(t *testing.T) *bytes.Buffer {
	t.Helper()

	return legacyHeaderStudents()
}

// LegacyNoticeStudentsBuffer 供 example 等无 testing.T 场景使用。
func LegacyNoticeStudentsBuffer() *bytes.Buffer {
	return legacyNoticeStudents()
}


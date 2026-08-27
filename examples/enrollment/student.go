package main

import "fmt"

const (
	sheetImport = "考生导入"
	noticeText  = "请按模板填写考生信息；年级请从下拉选择 A 或 B。"
)

// student 考务导入 DTO：Excel 里年级是 A/B，业务里是 1/2。
type student struct {
	Notice string `excel:"notice"`
	Name   string `excel:"姓名" validate:"required"`
	IDCard string `excel:"身份证" validate:"required,len=18"`
	Age    int    `excel:"年龄" validate:"required,gte=1,lte=100"`
	Grade  int    `excel:"年级" validate:"required,oneof=1 2"`
}

func (s *student) ExcelGrade(raw string) error {
	n, err := parseGrade(raw)
	if err != nil {
		return err
	}
	s.Grade = n
	return nil
}

func (s *student) ExcelExportGrade() (string, error) {
	return formatGrade(s.Grade)
}

func parseGrade(raw string) (int, error) {
	switch raw {
	case "A":
		return 1, nil
	case "B":
		return 2, nil
	default:
		return 0, fmt.Errorf("年级仅支持 A/B，收到 %q", raw)
	}
}

func formatGrade(n int) (string, error) {
	switch n {
	case 1:
		return "A", nil
	case 2:
		return "B", nil
	default:
		return "", fmt.Errorf("未知年级 %d", n)
	}
}

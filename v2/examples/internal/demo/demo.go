// Package demo 提供 examples 共用的 DTO、年级转换钩子与 validator 适配。
package demo

import (
	"fmt"

	playvalidator "github.com/go-playground/validator/v10"
)

const (
	SheetStudentImport = "考生导入"
	NoticeFillStudents = "请按模板填写考生信息"
)

// StudentRow 考生导入 DTO（与 e2e fixture 同构，examples 不依赖 e2e）。
type StudentRow struct {
	Notice string `excel:"notice"`
	Name   string `excel:"姓名" validate:"required"`
	IDCard string `excel:"身份证"`
	Age    int    `excel:"年龄"`
	Grade  int    `excel:"年级"`
}

func (r *StudentRow) ExcelGrade(raw string) error {
	switch raw {
	case "A":
		r.Grade = 1
		return nil
	case "B":
		r.Grade = 2
		return nil
	case "":
		r.Grade = 0
		return nil
	default:
		return fmt.Errorf("unknown grade %q", raw)
	}
}

func (r *StudentRow) ExcelExportGrade() (string, error) {
	switch r.Grade {
	case 1:
		return "A", nil
	case 2:
		return "B", nil
	case 0:
		return "", nil
	default:
		return "", fmt.Errorf("unknown grade %d", r.Grade)
	}
}

// PlaygroundValidator 对接 go-playground/validator，与 example_test.go 同模式。
type PlaygroundValidator struct {
	v *playvalidator.Validate
}

func NewPlaygroundValidator() PlaygroundValidator {
	return PlaygroundValidator{v: playvalidator.New()}
}

func (p PlaygroundValidator) Validate(row any) error {
	return p.v.Struct(row)
}

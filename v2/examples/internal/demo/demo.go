// Package demo 提供 examples 共用的 DTO、转换器与 validator 适配。
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
	Grade  int    `excel:"年级" conv:"grade"`
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

func GradeImport(raw string) (any, error) {
	switch raw {
	case "A":
		return 1, nil
	case "B":
		return 2, nil
	default:
		return 0, fmt.Errorf("unknown grade %q", raw)
	}
}

func GradeExport(v any) (string, error) {
	switch n := v.(type) {
	case int:
		switch n {
		case 1:
			return "A", nil
		case 2:
			return "B", nil
		default:
			return "", fmt.Errorf("unknown grade %d", n)
		}
	default:
		return "", fmt.Errorf("bad grade type")
	}
}

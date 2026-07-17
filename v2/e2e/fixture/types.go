package fixture

import (
	"fmt"
	"time"
)

// StudentImportRow 考生批量导入 DTO（NoticeHeaderData + 年级 conv）。
type StudentImportRow struct {
	Notice string `excel:"notice"`
	Name   string `excel:"姓名" validate:"required"`
	IDCard string `excel:"身份证"`
	Age    int    `excel:"年龄"`
	Grade  int    `excel:"年级" conv:"grade"`
}

// ReorderedRow 表头顺序与 struct 字段不一致时的绑定 DTO。
type ReorderedRow struct {
	Age    int    `excel:"年龄"`
	Name   string `excel:"姓名" validate:"required"`
	Grade  int    `excel:"年级" conv:"grade"`
	Extra  string `excel:"备注"`
	Unused string `excel:"-"`
}

// Address 嵌套地址（inline flatten）。
type Address struct {
	City   string `excel:"城市"`
	Street string `excel:"街道"`
}

// InlineAddressRow 嵌套 inline 地址行。
type InlineAddressRow struct {
	Name string  `excel:"姓名"`
	Addr Address `excel:",inline"`
}

// ScoreRow 无 notice 的 HeaderData 导出/导入行。
type ScoreRow struct {
	Name  string `excel:"姓名"`
	Score int    `excel:"分数"`
}

// TemplateDistributeRow 发模板给业务方的行（含等级下拉）。
type TemplateDistributeRow struct {
	Notice string `excel:"notice"`
	Name   string `excel:"姓名" validate:"required"`
	Level  string `excel:"年级"`
}

// TimeBoolRow 时间 + 布尔字段行（builtin convert）。
type TimeBoolRow struct {
	Notice string    `excel:"notice"`
	Name   string    `excel:"姓名" validate:"required"`
	Active bool      `excel:"启用"`
	Joined time.Time `excel:"入学日期" time:"2006-01-02"`
}

// LegacyStudentRow 复刻原 testdata StudentRow（等级列名，无身份证列）。
type LegacyStudentRow struct {
	Name  string `excel:"姓名" validate:"required"`
	Age   int    `excel:"年龄"`
	Grade int    `excel:"等级" conv:"grade"`
}

// GradeImport A/B → int。
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

// GradeExport int → A/B。
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

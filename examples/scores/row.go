package main

// scoreRow 期末成绩 DTO：固定考生列 + 动态科目分。
type scoreRow struct {
	Name  string            `excel:"姓名"`
	Grade string            `excel:"年级"`
	Class string            `excel:"班级"`
	Extra map[string]string `excel:",extend"`
}

const (
	sheetScores = "期末成绩"
	noticeText  = "年级、班级请从下拉选择；成绩列按科目填写。"
)

var (
	subjects   = []string{"语文", "数学", "英语"}
	gradeClass = map[string][]string{
		"高一": {"1班", "2班"},
		"高二": {"3班", "4班"},
	}
)

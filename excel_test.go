package excelizex

import (
	"testing"
)

func TestSetPullDown(t *testing.T) {
	t.Run("TestGen", func(t *testing.T) {
		var ttt testStruct
		sheet := NewSheet().SetHeaderByStruct(ttt).
			SetOptions("性别", []any{"男", "女", "坤"}).
			SetOptions("C", []any{"hello", "world"}).
			SetName("下拉菜单测试")

		if err := New().AddSheets(sheet).SaveAs("./test_file/pull_down.xlsx"); err != nil {
			t.Fatal(err)
		}
	})
}

type ImportSchoolReq struct {
	// 名字
	Name string `json:"name" validate:"required,min=2,max=63" excel:"学校名称"`
	// 学校学段
	SchoolStage string `json:"schoolStage" validate:"required" excel:"学校类别"`
	// 区域名称
	AreaName string `json:"areaCode" validate:"required" excel:"所属区域"`
	// 管理员编号
	ContactorName string `json:"contactorName" validate:"omitempty" excel:"用户名"`
	// 管理员手机号
	ContactorPhone string `json:"contactorPhone" validate:"omitempty" excel:"学校管理员手机号"`
	// 管理员身份证号
	ContactorIdCard string `json:"contactorIdCard" validate:"required,len=18" excel:"学校管理员身份证号"`
	// 是否开启智能行为分析
	Analyse string `json:"analyse" validate:"required" excel:"是否开启智能行为分析"`
}

func TestSetNotice(t *testing.T) {
	t.Run("TestGen", func(t *testing.T) {
		sheet := NewSheet().SetHeaderByStruct(new(ImportSchoolReq)).
			SetOptions("所属区域", []any{"男", "女", "坤"}).
			SetNotice(`填表说明：
1. 请严格按照本模板填写导入用户信息，不要修改表头信息（本表1、2行）
2. 用户姓名、手机号是必填项。用户姓名需在2 - 30个字符之间，头尾不允许空格；手机号按照11位手机号填写，头尾不允许空格，不需加国家码前缀。
3. 身份证为非必填项。若需填写请严格按照国家有效身份证填写，头尾不允许空格。
4. 用户组为非必填项。用户组名称需在2 - 30个字符之间，头尾不允许空格。若填写此项，则代表该用户加入指定用户组。若该用户组已经在本机构存在，则意为加入该用户组（不影响现有用户组成员）`,
			).SetName("下拉菜单测试")

		if err := New().AddSheets(sheet).SaveAs("./test_file/pull_down.xlsx", "test1234"); err != nil {
			t.Fatal(err)
		}
	})
}

func TestGenExcel(t *testing.T) {
	sheet := NewSheet().
		SetName("导入学校").
		SetNotice(
			"*表格中所有字段均为必填项，其中“学校类别、所属区域、是否开启智能行为分析”请下拉选择。否则可能会导致数据导入失败；\n"+
				"若导入数据的学校管理员已存在时，则数据不会覆盖系统数据。",
		).
		SetHeaderByStruct(new(ImportSchoolReq)).
		SetOptions(
			"学校类1",
			[]any{"小学5年制", "小学6年制", "普通初中3年制", "普通初中4年制", "普通高中",
				"九年一贯制", "十二年一贯制", "完全中学"}).
		SetOptions("是否开启智能行为分析", []any{"是", "否"})

	if err := New().AddSheets(sheet).SaveAs("./test_file/ImportSchoolReq.xlsx", "test1234"); err != nil {
		t.Fatal(err)
	}
}

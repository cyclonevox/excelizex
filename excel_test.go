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

func TestSetNotice(t *testing.T) {
	t.Run("TestGen", func(t *testing.T) {
		var ttt testStruct
		sheet := NewSheet().SetHeaderByStruct(ttt).
			SetOptions("性别", []any{"男", "女", "坤"}).
			SetNotice(`填表说明：
1. 请严格按照本模板填写导入用户信息，不要修改表头信息（本表1、2行）
2. 用户姓名、手机号是必填项。用户姓名需在2 - 30个字符之间，头尾不允许空格；手机号按照11位手机号填写，头尾不允许空格，不需加国家码前缀。
3. 身份证为非必填项。若需填写请严格按照国家有效身份证填写，头尾不允许空格。
4. 用户组为非必填项。用户组名称需在2 - 30个字符之间，头尾不允许空格。若填写此项，则代表该用户加入指定用户组。若该用户组已经在本机构存在，则意为加入该用户组（不影响现有用户组成员）`,
			).SetName("下拉菜单测试")

		if err := New().AddSheets(sheet).SaveAs("./test_file/pull_down.xlsx"); err != nil {
			t.Fatal(err)
		}
	})
}

package excelizex

import (
	"testing"
)

func TestSetPullDown(t *testing.T) {
	t.Run("TestGen", func(t *testing.T) {
		var ttt testStruct
		sheet := NewSheet().SetHeaderByStruct(ttt).
			SetOptions(NewPullDown().AddOptions("B", []any{"男", "女", "坤"})).
			SetName("下拉菜单测试")

		if err := New().AddSheets(sheet).SaveAs("./test_file/pull_down.xlsx"); err != nil {
			t.Fatal(err)
		}
	})
}

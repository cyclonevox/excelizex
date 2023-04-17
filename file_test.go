package excelizex

import (
	"strconv"
	"testing"
)

func TestGenerateFile(t *testing.T) {
	t.Run("no_data_no_style", func(t *testing.T) {
		err := New().AddSheet("helloWorld", new(TestNoStyle)).SaveAs("./test/no_data_no_style.xlsx")
		if err != nil {
			t.Errorf("has_data_no_style: %s", err)
		}
	})

	t.Run("no_data_has_style", func(t *testing.T) {
		ttt := new(TestHasStyle)
		ttt.Notice = "你好世界你好世界你好世界你好世界你好世界你好世界你好世界"
		err := New().AddSheet("helloWorld", ttt).SaveAs("./test/no_data_has_style.xlsx")
		if err != nil {
			t.Errorf("has_data_has_style: %s", err)
		}
	})

	t.Run("has_data_no_style", func(t *testing.T) {
		var hasdata []*TestNoStyle
		for i := 0; i < 100; i++ {
			hasdata = append(
				hasdata,
				&TestNoStyle{
					Notice: "你好世界你好世界你好世界你好世界你好世界你好世界你好世界",
					Name:   "hello" + strconv.FormatInt(int64(i), 10),
					Phone:  i,
				},
			)
		}

		err := New().AddSheet("helloWorld", hasdata).SaveAs("./test/has_data_no_style.xlsx")
		if err != nil {
			t.Errorf("has_data_no_style: %s", err)
		}
	})

	t.Run("has_data_has_style", func(t *testing.T) {
		var hasdata []*TestHasStyle
		for i := 0; i < 100; i++ {
			hasdata = append(
				hasdata,
				&TestHasStyle{
					Notice: "你好世界你好世界你好世界你好世界你好世界你好世界你好世界你好世界你好世界",
					Name:   "hello" + strconv.FormatInt(int64(i), 10),
					Phone:  i,
					Id:     i,
				},
			)
		}

		err := New().AddSheet("helloWorld", hasdata).SaveAs("./test/has_data_has_style.xlsx", "1")
		if err != nil {
			t.Errorf("has_data_has_style: %s", err)
		}
	})

	t.Run("has_data_has_style_has_pull_down", func(t *testing.T) {
		var hasdata []*TestHasStyle
		for i := 0; i < 100; i++ {
			hasdata = append(
				hasdata,
				&TestHasStyle{
					Notice: "你好世界你好世界你好世界你好世界你好世界你好世界你好世界你好世界你好世界",
					Name:   "hello" + strconv.FormatInt(int64(i), 10),
					Phone:  i,
					Id:     i,
				},
			)
		}

		err := New().AddSheet("考生导入", hasdata,
			NewOptions("学生姓名", []string{"tom", "jerry"}),
			NewOptions("学生号码", []string{"13380039232", "13823021932", "17889032312"}),
			NewOptions("学生编号", []string{"1", "2", "3"}),
		).SaveAs("./test/has_data_has_style_has_pull_down.xlsx")
		if err != nil {
			t.Errorf("has_data_has_style: %s", err)
		}
	})
}

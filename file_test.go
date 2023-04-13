package excelizex

import (
	"strconv"
	"testing"
)

func TestGenerateFile(t *testing.T) {
	t.Run("no_data_no_style", func(t *testing.T) {
		err := New().AddSheets(NewSheet("helloWorld", new(TestNoStyle))).SaveAs("./test/no_data_no_style.xlsx")
		if err != nil {
			t.Errorf("has_data_no_style: %s", err)
		}
	})

	t.Run("no_data_has_style", func(t *testing.T) {
		ttt := new(TestHasStyle)
		ttt.Notice = "你好世界你好世界你好世界你好世界你好世界你好世界你好世界"
		err := New().AddSheets(NewSheet("helloWorld", ttt)).SaveAs("./test/no_data_has_style.xlsx")
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

		err := New().AddSheets(NewSheet("helloWorld", hasdata)).SaveAs("./test/has_data_no_style.xlsx")
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

		err := New().AddSheets(NewSheet("helloWorld", hasdata)).SaveAs("./test/has_data_has_style.xlsx", "1")
		if err != nil {
			t.Errorf("has_data_has_style: %s", err)
		}
	})
}

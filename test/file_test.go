package test

import (
	"strconv"
	"testing"

	"github.com/cyclonevox/excelizex"
)

func TestGenerateFile(t *testing.T) {

	t.Run("no_data_no_style", func(t *testing.T) {
		err := excelizex.New().AddSheet("helloWorld", new(noStyle)).SaveAs("./xlsx/no_data_no_style.xlsx")
		if err != nil {
			t.Errorf("has_data_no_style: %s", err)
		}
	})

	t.Run("no_data_has_style", func(t *testing.T) {
		ttt := new(hasStyle)
		ttt.Notice = "你好世界你好世界你好世界你好世界你好世界你好世界你好世界"
		err := excelizex.New().AddSheet("helloWorld", ttt).SaveAs("./xlsx/no_data_has_style.xlsx")
		if err != nil {
			t.Errorf("has_data_has_style: %s", err)
		}
	})

	t.Run("has_data_no_style", func(t *testing.T) {
		var hasdata []*noStyle
		for i := 0; i < 100; i++ {
			hasdata = append(
				hasdata,
				&noStyle{
					Notice: "你好世界你好世界你好世界你好世界你好世界你好世界你好世界",
					Name:   "hello" + strconv.FormatInt(int64(i), 10),
					Phone:  i,
				},
			)
		}

		err := excelizex.New().AddSheet("helloWorld", hasdata).SaveAs("./xlsx/has_data_no_style.xlsx")
		if err != nil {
			t.Errorf("has_data_no_style: %s", err)
		}
	})

	t.Run("has_data_has_style", func(t *testing.T) {
		var hasdata []*hasStyle
		for i := 0; i < 100; i++ {
			hasdata = append(
				hasdata,
				&hasStyle{
					Notice: "你好世界你好世界你好世界你好世界你好世界你好世界你好世界你好世界你好世界",
					Name:   "hello" + strconv.FormatInt(int64(i), 10),
					Phone:  100 + i,
					Id:     i,
				},
			)
		}

		err := excelizex.New().AddSheet("helloWorld", hasdata).SaveAs("./xlsx/has_data_has_style.xlsx", "1")
		if err != nil {
			t.Errorf("has_data_has_style: %s", err)
		}
	})

	t.Run("has_data_has_style_has_pull_down", func(t *testing.T) {
		var hasdata []*hasStyle
		for i := 0; i < 100; i++ {
			hasdata = append(
				hasdata,
				&hasStyle{
					Notice: "你好世界你好世界你好世界你好世界你好世界你好世界你好世界你好世界你好世界",
					Name:   "hello" + strconv.FormatInt(int64(i), 10),
					Phone:  100 + i,
					Id:     i,
				},
			)
		}

		err := excelizex.New().AddSheet("考生导入", hasdata,
			excelizex.NewOptions("学生姓名", []string{"tom", "jerry"}),
			excelizex.NewOptions("学生号码", []string{"13380039232", "13823021932", "17889032312"}),
			excelizex.NewOptions("学生编号", []string{"1", "2", "3"}),
		).SaveAs("./xlsx/has_data_has_style_has_pull_down.xlsx")
		if err != nil {
			t.Errorf("has_data_has_style: %s", err)
		}
	})

	t.Run("has_data_style_pull_down_dynamic", func(t *testing.T) {
		var hasdata []*hasStyleHasDynamic
		for i := 0; i < 100; i++ {
			hasdata = append(
				hasdata,
				&hasStyleHasDynamic{
					Notice: "你好世界你好世界你好世界你好世界你好世界你好世界你好世界你好世界你好世界",
					Name:   "hello" + strconv.FormatInt(int64(i), 10),
					Phone:  100 + i,
					Id:     i,
					ExtInfo: []excelizex.DefaultExtHeader{
						{
							HeaderName:  "他爹的年纪",
							StyleTag:    "default-header-red",
							ValidateTag: "",
							Data:        33 + i,
						},
						{
							HeaderName:  "他爷的年纪",
							StyleTag:    "default-header-red",
							ValidateTag: "",
							Data:        55 + i,
						},
					},
				},
			)
		}

		err := excelizex.New().AddSheet("考生导入", hasdata,
			excelizex.NewOptions("学生姓名", []string{"tom", "jerry"}),
			excelizex.NewOptions("学生号码", []string{"13380039232", "13823021932", "17889032312"}),
			excelizex.NewOptions("学生编号", []string{"1", "2", "3"}),
		).SaveAs("./xlsx/has_data_style_pull_down_dynamic.xlsx", "test")
		if err != nil {
			t.Errorf("has_data_has_style: %s", err)
		}
	})
}

package excelizex

import (
	"fmt"
	"testing"
)

type TestNoStyle struct {
	Notice string `excel:"notice"`
	Name   string `excel:"header|姓名"`
	Phone  int    `excel:"header|号码"`
}

func TestGenerateSheet(t *testing.T) {
	t.Run("no_data_no_style", func(t *testing.T) {
		newSheet := NewSheet("helloWorld", new(TestNoStyle))
		fmt.Println(newSheet)
	})

	t.Run("has_data_no_style", func(t *testing.T) {
		var hasdata []*TestNoStyle
		for i := 0; i < 100; i++ {
			hasdata = append(
				hasdata,
				&TestNoStyle{
					Notice: "你好世界",
					Name:   "hello" + string(rune(i)),
					Phone:  i,
				},
			)
		}

		newSheet := NewSheet("helloWorld", hasdata)
		fmt.Println(newSheet)
	})
}

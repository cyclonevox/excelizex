package excelizex

import (
	"fmt"
	"strconv"
	"testing"
)

type TestNoStyle struct {
	Notice string `excel:"notice"`
	Name   string `excel:"header|学生姓名"`
	Phone  int    `excel:"header|学生号码"`
}

type TestHasStyle struct {
	Notice string `excel:"notice" style:"default-notice"`
	Name   string `excel:"header|学生姓名" style:"default-header"`
	Phone  int    `excel:"header|学生号码" style:"default-header-red"`
	Id     int    `excel:"header|学生编号" style:"default-header-red"`
	p      []*TestNoStyle
}

func TestGenerateSheet(t *testing.T) {
	t.Run("no_data_no_style", func(t *testing.T) {
		newSheet := NewSheet("helloWorld", new(TestNoStyle))
		fmt.Println(newSheet)
	})

	t.Run("no_data_no_style_has_notice", func(t *testing.T) {
		ttt := new(TestNoStyle)
		ttt.Notice = "new(TestNoStyle)new(TestNoStyle)new(TestNoStyle)"
		newSheet := NewSheet("helloWorld", ttt)
		fmt.Println(newSheet)
	})

	t.Run("has_data_no_style", func(t *testing.T) {
		var hasdata []*TestHasStyle
		for i := 0; i < 100; i++ {
			hasdata = append(
				hasdata,
				&TestHasStyle{
					Notice: "你好世界",
					Name:   "hello" + strconv.FormatInt(int64(i), 10),
					Phone:  i,
					Id:     i,
				},
			)
		}

		newSheet := NewSheet("helloWorld", hasdata)
		fmt.Println(newSheet)
	})

	t.Run("has_data_has_style", func(t *testing.T) {
		var hasdata []*TestHasStyle
		for i := 0; i < 100; i++ {
			hasdata = append(
				hasdata,
				&TestHasStyle{
					Notice: "你好世界",
					Name:   "hello" + strconv.FormatInt(int64(i), 10),
					Phone:  i,
					Id:     i,
				},
			)
		}

		m := newMetas(hasdata)
		for _, rr := range m.data {
			t.Log(rr)
		}

		newSheet := NewSheet("helloWorld", hasdata)
		fmt.Println(newSheet)
	})
}

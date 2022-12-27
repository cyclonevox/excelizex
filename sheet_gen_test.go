package excelizex

import (
	"reflect"
	"testing"
)

type testStruct struct {
	Name      string `file:"名称" json:"sheet"`
	Sex       int    `file:"性别" json:"sex"`
	HelloWord string `file:"测试" json:"helloWord"`
}

func TestGen(t *testing.T) {
	t.Run("TestGen", func(t *testing.T) {
		var ttt testStruct
		var expectSheet = Sheet{
			Header: []string{"名称", "性别", "测试"},
		}
		sheet := genSheet(ttt)

		if !reflect.DeepEqual(expectSheet, sheet) {
			t.Fatalf("expect %+v,but %+v", expectSheet, sheet)
		}
	})
}

func TestSliceGen(t *testing.T) {
	t.Run("TestGen", func(t *testing.T) {
		ttt := []testStruct{
			{"123", 123, "456"},
			{"456", 231, "213"},
		}

		var expectSheet = Sheet{
			Header: []string{"名称", "性别", "测试"},
			Data: [][]any{
				{"123", 123, "456"},
				{"456", 231, "213"},
			},
		}
		sheet := Gen(ttt)

		if !reflect.DeepEqual(expectSheet, sheet) {
			t.Fatalf("expect %+v,but %+v", expectSheet, sheet)
		}
	})
}

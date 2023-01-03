package excelizex

import (
	"reflect"
	"testing"
)

var (
	testName   = "test_sheet"
	testNotice = "test_sheet_notice"
	testHeader = []string{"名称", "性别", "测试"}
)

type testStruct struct {
	Name       string `excel:"名称" json:"sheet"`
	Sex        string `excel:"性别" json:"sex"`
	HelloWorld string `excel:"测试" json:"helloWorld"`
}

type testStructs []testStruct

func (t testStructs) ToExpectData() [][]string {
	var ss [][]string

	ss = append(ss, []string{testNotice})
	ss = append(ss, []string{"名称", "性别", "测试"})
	for _, ts := range t {
		ss = append(ss, []string{ts.Name, ts.Sex, ts.HelloWorld})
	}

	return ss
}

type testErrStruct struct {
	Name       string `excel:"名称" json:"sheet"`
	Sex        string `excel:"性别" json:"sex"`
	HelloWorld string `excel:"测试" json:"helloWorld"`
	Err        string `excel:"错误原因" json:"err"`
}

type testErrStructs []testErrStruct

func (t testErrStructs) ToExpectData() [][]string {
	var ss [][]string

	ss = append(ss, []string{testNotice})
	ss = append(ss, []string{"名称", "性别", "测试", "错误原因"})
	for _, ts := range t {
		ss = append(ss, []string{ts.Name, ts.Sex, ts.HelloWorld, ts.Err})
	}

	return ss
}

func TestGen(t *testing.T) {
	t.Run("TestGen", func(t *testing.T) {
		var ttt testStruct
		var expectSheet = &Sheet{
			Header: []string{"名称", "性别", "测试"},
		}
		sheet := NewSheet().SetHeaderByStruct(ttt)

		if !reflect.DeepEqual(expectSheet, sheet) {
			t.Fatalf("expect %+v,but %+v", expectSheet, sheet)
		}
	})
}

func TestSliceGen(t *testing.T) {
	t.Run("TestGen", func(t *testing.T) {
		ttt := []testStruct{
			{"123", "男", "456"},
			{"456", "女", "213"},
		}

		var expectSheet = &Sheet{
			Header: []string{"名称", "性别", "测试"},
			Data: [][]any{
				{"123", "男", "456"},
				{"456", "女", "213"},
			},
		}
		sheet := genDataSheet(ttt)

		if !reflect.DeepEqual(expectSheet, sheet) {
			t.Fatalf("expect %+v,but %+v", expectSheet, sheet)
		}
	})
}

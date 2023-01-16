package excelizex

import (
	"fmt"
	"reflect"
	"testing"
)

func TestFile_SetResults(t *testing.T) {
	vava := &testStruct{}
	f := New().AddSimpleSheet(
		vava,
		Name(testName),
		Notice(testNotice),
		Data([][]any{{"测试人员1", "无", "123123123"}, {"测试人员2", "公", "helloWorld"}}),
		Options(NewPullDown().AddOptions("B", []any{"男", "女"})),
	)

	results := f.SelectSheet(testName).Read(vava, func() error {
		fmt.Println(vava)

		return nil
	})

	rows, err := f.excel().GetRows(testName)

	f, exist, err := f.SetResults(&results)
	if err != nil {
		panic(err)
	}
	if exist {
		rows, err = f.excel().GetRows(testName)
		if err != nil {
			t.Fatal("TestFile_SetResults:", " 表数据获取失败", err)
		}

		testData := []testErrStruct{
			{"测试人员1", "无", "123123123", "Sex必须是[男 女]中的一个"},
			{"测试人员2", "公", "helloWorld", "Sex必须是[男 女]中的一个"},
		}

		if err = f.SaveAs("./test_file/result1.xlsx"); err != nil {
			t.Fatal(err)
		}

		expectData := testErrStructs(testData).ToExpectData()
		if !reflect.DeepEqual(expectData, rows) {
			t.Fatalf("Expect:%+v,but%+v", expectData, rows)
		}

	}
}

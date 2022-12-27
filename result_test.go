package excelizex

import (
	"reflect"
	"testing"
)

func TestFile_SetResults(t *testing.T) {
	var testResult = Result{
		SheetName: testName,
		Notice:    testNotice,
		Header:    testHeader,
		Errors: []ErrorInfo{
			{
				ErrorRow:  2,
				RawData:   []string{"测试人员1", "无", "123123123"},
				ErrorInfo: []string{"性别只能是男或者女"},
			},
			{
				ErrorRow:  3,
				RawData:   []string{"测试人员2", "公", "helloWorld"},
				ErrorInfo: []string{"性别只能是男或者女"},
			},
		},
	}

	results := New().SetResults(&testResult)

	rows, err := results.excel().GetRows(testResult.SheetName)
	if err != nil {
		t.Fatal("TestFile_SetResults:", " 表数据获取失败", err)
	}

	testData := []testErrStruct{
		{"测试人员1", "无", "123123123", "性别只能是男或者女"},
		{"测试人员2", "公", "helloWorld", "性别只能是男或者女"},
	}

	expectData := testErrStructs(testData).ToExpectData()
	if !reflect.DeepEqual(expectData, rows) {
		t.Fatalf("Expect:%+v,but%+v", expectData, rows)
	}
}

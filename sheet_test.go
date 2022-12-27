package excelizex

import (
	"reflect"
	"testing"
)

var (
	testName   = "test_sheet"
	testNotice = "test_sheet notice"
	testHeader = []string{"test1", "test2", "test3"}
)

func TestNewSheet(t *testing.T) {
	expectSheet := &Sheet{
		Name:   testName,
		Notice: testNotice,
		Header: testHeader,
		Data:   [][]any{{"s123", "123"}, {"123123"}, {"123123"}},
	}

	newSheet := NewSheet(
		SetName(testName),
		SetNotice(testNotice),
		SetHeader(testHeader),
		SetData([][]any{{"s123", "123"}, {"123123"}, {"123123"}}),
	)

	if !reflect.DeepEqual(expectSheet, newSheet) {
		t.Fatalf("expect %+v,but %+v", expectSheet, newSheet)
	}
}

func TestSheet_writeRowIncrWrite(t *testing.T) {
	var (
		exceptRowName string
		testRowName   string
	)

	newSheet := NewSheet(
		SetName(testName),
		SetNotice(testNotice),
		SetHeader(testHeader),
		SetData([][]any{{"s123", "123"}, {"123123"}, {"123123"}}),
	)

	testRowName = newSheet.writeRowIncr()
	exceptRowName = "A1"

	if testRowName != exceptRowName {
		t.Fatalf("expect %+v,but %+v", exceptRowName, testRowName)
	}

	testRowName = newSheet.writeRowIncr(3)
	exceptRowName = "A4"

	if testRowName != exceptRowName {
		t.Fatalf("expect %+v,but %+v", exceptRowName, testRowName)
	}
}

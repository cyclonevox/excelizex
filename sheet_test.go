package excelizex

import (
	"reflect"
	"testing"
)

var (
	testName   = "test_sheet"
	testNotice = "test_sheet notice"
	testHeader = []string{"test1", "test2", "test3"}
	testData   = [][]any{{"s123", "123"}, {"123123"}, {"123123"}}
)

func TestNewSheet(t *testing.T) {
	expectSheet := &Sheet{
		Name:   testName,
		Notice: testNotice,
		Header: testHeader,
		Data:   testData,
	}

	newSheet := NewSheet(
		SetName(testName),
		SetNotice(testNotice),
		SetHeader(testHeader),
		SetData(testData),
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
		SetData(testData),
	)

	testRowName = newSheet.writeRowIncr()
	exceptRowName = "A1"

	if testRowName != exceptRowName {
		t.Fatalf("expect %+v,but %+v", exceptRowName, testRowName)
	}

	newSheet.writeRowIncr(3)
	exceptRowName = "A4"

	if testRowName != exceptRowName {
		t.Fatalf("expect %+v,but %+v", exceptRowName, testRowName)
	}
}

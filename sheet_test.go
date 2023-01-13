package excelizex

import (
	"reflect"
	"testing"
)

func TestNewSheet(t *testing.T) {
	expectSheet := &Sheet{
		Name:   testName,
		Notice: testNotice,
		Header: testHeader,
		Data:   [][]any{{"s123", "123"}, {"123123"}, {"123123"}},
	}

	newSheet := NewSheet(
		Name(testName),
		Notice(testNotice),
		Header(testHeader),
		Data([][]any{{"s123", "123"}, {"123123"}, {"123123"}}),
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
		Name(testName),
		Notice(testNotice),
		Header(testHeader),
		Data([][]any{{"s123", "123"}, {"123123"}, {"123123"}}),
	)

	testRowName = newSheet.nextWriteRow()
	exceptRowName = "A1"

	if testRowName != exceptRowName {
		t.Fatalf("expect %+v,but %+v", exceptRowName, testRowName)
	}

	testRowName = newSheet.nextWriteRow(3)
	exceptRowName = "A4"

	if testRowName != exceptRowName {
		t.Fatalf("expect %+v,but %+v", exceptRowName, testRowName)
	}
}

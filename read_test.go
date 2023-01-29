package excelizex

import (
	"os"
	"reflect"
	"strconv"
	"testing"
)

type readTestStruct struct {
	Id   int64  `excel:"埃低"`
	Name string `excel:"名称"`
	List []*struct {
		Id int64
	} `excel:"列表" excel-conv:"list"`
}

func listConvert(rawData string) (any, error) {
	i, err := strconv.ParseInt(rawData, 10, 64)
	if err != nil {
		return nil, err
	}

	return []*struct{ Id int64 }{{i}}, nil
}

func TestConvertRead(t *testing.T) {
	f, err := os.Open("./test_file/read_test.xlsx")
	if err != nil {
		panic(err)
	}
	file := New(f)
	s := new(readTestStruct)

	var expectPtr = []*readTestStruct{
		{103, "张3", []*struct{ Id int64 }{{123}}},
		{104, "张4", []*struct{ Id int64 }{{124}}},
		{105, "张5", []*struct{ Id int64 }{{125}}},
		{106, "张6", []*struct{ Id int64 }{{126}}},
		{107, "张7", []*struct{ Id int64 }{{127}}},
		{108, "张8", []*struct{ Id int64 }{{128}}},
		{109, "张9", []*struct{ Id int64 }{{129}}},
		{110, "张10", []*struct{ Id int64 }{{130}}},
	}

	var sListPtr []*readTestStruct
	file.SelectSheet("测试用表").
		SetConvert("list", listConvert).
		Read(s, func() error {
			data := *s
			sListPtr = append(sListPtr, &data)

			return nil
		})

	for index := range expectPtr {
		if !reflect.DeepEqual(sListPtr[index], expectPtr[index]) {
			t.Fatalf("index:%d,Expect:%+v,but%+v", index, sListPtr[index], expectPtr[index])
		}
	}

	var expect = []readTestStruct{
		{103, "张3", []*struct{ Id int64 }{{123}}},
		{104, "张4", []*struct{ Id int64 }{{124}}},
		{105, "张5", []*struct{ Id int64 }{{125}}},
		{106, "张6", []*struct{ Id int64 }{{126}}},
		{107, "张7", []*struct{ Id int64 }{{127}}},
		{108, "张8", []*struct{ Id int64 }{{128}}},
		{109, "张9", []*struct{ Id int64 }{{129}}},
		{110, "张10", []*struct{ Id int64 }{{130}}},
	}

	var sList []readTestStruct

	file.SelectSheet("测试用表").
		SetConvert("list", listConvert).
		Read(s, func() error {
			sList = append(sList, *s)

			return nil
		})

	for index := range expect {
		if !reflect.DeepEqual(sList[index], expect[index]) {
			t.Fatalf("index:%d,Expect:%+v,but%+v", index, sList[index], expect[index])
		}
	}
}

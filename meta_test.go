package excelizex

import (
	"fmt"
	"testing"
)

type TestParseMeta struct {
	Notice    string             `excel:"notice" style:"default-notice"`
	Name      string             `excel:"header|学生姓名" style:"default-header"`
	Phone     string             `excel:"header|学生号码" style:"default-header-red"`
	Id        int                `excel:"header|学生编号" style:"default-header-red"`
	SingleExt DefaultExtHeader   `excel:"extend"`
	ExtInfo   []DefaultExtHeader `excel:"extend"`
	Parent    ParentInfo         `excel:"extend"`
	Parents   []ParentInfo       `excel:"extend"`
}

type ParentInfo struct {
	Name  string `excel:"header|家人姓名" style:"default-header-red"`
	Phone int    `excel:"header|家人手机号" style:"default-header-red"`
	Sex   string `excel:"header|性别"`
}

func TestManipulateMetaRaw(t *testing.T) {
	tpm := TestParseMeta{
		Notice:    "test test test test",
		SingleExt: DefaultExtHeader{StyleTag: "default-header", HeaderName: "999"},
		ExtInfo: []DefaultExtHeader{
			{
				StyleTag:   "default-header",
				HeaderName: "123",
			},
			{
				StyleTag:   "default-header",
				HeaderName: "456",
			},
		},
		Parent:  ParentInfo{},
		Parents: make([]ParentInfo, 5),
	}

	t.Run("new_meta_raw_list", func(t *testing.T) {
		r := newMetas(tpm)
		for _, rr := range r.raws {
			fmt.Println(*rr)
		}
	})

}

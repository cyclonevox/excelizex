package validate_test

import (
	"testing"

	"github.com/cyclonevox/excelizex/v2/validate"
)

type row struct {
	Name string `excel:"姓名" validate:"required"`
	Age  int    `excel:"年龄"`
}

func TestRequired(t *testing.T) {
	v := validate.Required{}
	if err := v.ValidateRow(row{Name: "张三"}); err != nil {
		t.Fatal(err)
	}
	if err := v.ValidateRow(row{}); err == nil {
		t.Fatal("expected required error")
	}
}

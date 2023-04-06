package excelizex

import (
	"fmt"
	"reflect"
	"testing"
)

type testetst struct {
	asd  int
	sad  int
	asds int
}

func getFuncParam(any any) any {
	a := reflect.New(reflect.TypeOf(any).Elem()).Interface()

	ttt := a.(*testetst)
	ttt.asd = 1
	ttt.sad = 2
	ttt.asds = 3

	return ttt
}

func TestName(t *testing.T) {
	var ttt *testetst
	a := getFuncParam(ttt)

	fmt.Println("hello", a)
}

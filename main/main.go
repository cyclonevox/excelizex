package main

import (
	"excelizex/validatorx"
	"fmt"
	_ "github.com/go-playground/validator/v10"
)

type A struct {
	B string `tag1:"b" tag2:"B"` //这引号里面的就是tag
	C string `tag1:"c" tag2:"C" excel:"身份证" excel-conv:"id" validate:"id_card"`
}

func main() {
	var a A
	a.C = "123123"
	vvv := validatorx.New()
	_, sss := vvv.Struct(a)
	fmt.Println(sss)
}

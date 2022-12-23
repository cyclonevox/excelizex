package main

import (
	"excelizex/validatorx"
	"fmt"
	_ "github.com/go-playground/validator/v10"
)

type A struct {
	B string `tag1:"b" tag2:"B"` //这引号里面的就是tag
	C string `tag1:"c" tag2:"C" validate:"id_card"`
}

func main() {
	var a A
	vvv := validatorx.New()
	_, sss := vvv.Val(a.C, "id_card")
	fmt.Println(sss)
}

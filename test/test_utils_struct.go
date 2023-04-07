package test

import (
	"github.com/cyclonevox/validatorx"
	"github.com/go-playground/validator/v10"
)

var validate validation

type validation struct {
	v *validator.Validate
}

func newValidation() *validation {
	if validate.v == nil {
		validate.v = validatorx.New()
	}

	return &validate
}

func (v *validation) Validate(i interface{}) error {
	return v.v.Struct(i)
}

type batchData struct {
	Id    string `excel:"userId" excel-conv:"id-string"`
	Phone string `excel:"phoneNumber"`
}

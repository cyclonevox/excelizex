package excelizex

import (
	"github.com/cyclonevox/validatorx"
	"github.com/go-playground/validator/v10"
)

type validate struct {
	validator *validator.Validate
}

func newValidate() *validate {
	return &validate{
		validator: validatorx.New(),
	}
}

func (v *validate) Validate(i interface{}) error {
	return v.validator.Struct(i)
}

func (v *validate) ValidateVal(i interface{}, tag string) error {
	return v.validator.Var(i, tag)
}
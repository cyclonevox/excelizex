package main

import playvalidator "github.com/go-playground/validator/v10"

type rowValidator struct {
	v *playvalidator.Validate
}

func newValidator() rowValidator {
	return rowValidator{v: playvalidator.New()}
}

func (r rowValidator) Validate(row any) error {
	return r.v.Struct(row)
}

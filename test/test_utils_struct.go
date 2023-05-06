package test

import (
	"github.com/cyclonevox/excelizex"
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
	Id    string `excel:"header|userId" excel-conv:"id-string"`
	Phone string `excel:"header|phoneNumber"`
}

type batchDataHasDynamic struct {
	Id    string `excel:"header|userId" excel-conv:"id-string"`
	Phone string `excel:"header|phoneNumber"`
}

type noStyle struct {
	Notice string `excel:"notice"`
	Name   string `excel:"header|学生姓名"`
	Phone  int    `excel:"header|学生号码"`
}

type hasStyle struct {
	Notice string `excel:"notice" style:"default-notice"`
	Name   string `excel:"header|学生姓名" style:"default-header"`
	Phone  int    `excel:"header|学生号码" style:"default-header-red"`
	Id     int    `excel:"header|学生编号" style:"default-header-red"`
}

type hasStyleHasDynamic struct {
	Notice  string                       `excel:"notice" style:"default-notice"`
	Name    string                       `excel:"header|学生姓名" style:"default-header"`
	Phone   int                          `excel:"header|学生号码" style:"default-header-red"`
	Id      int                          `excel:"header|学生编号" style:"default-header-red"`
	ExtInfo []excelizex.DefaultExtHeader `excel:"extend"`
}

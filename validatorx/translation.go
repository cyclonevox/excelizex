package validatorx

import (
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
)

func initTranslation(validate *validator.Validate, chinese ut.Translator) (err error) {
	if err = validate.RegisterTranslation(
		"id_card",
		chinese,
		func(ut ut.Translator) error {
			return ut.Add("id_card", "身份证不符合规范", true)
		},
		func(ut ut.Translator, fe validator.FieldError) string {
			t, _ := ut.T("id_card")
			return t
		},
	); nil != err {
		return
	}

	if err = validate.RegisterTranslation(
		"mobile",
		chinese,
		func(ut ut.Translator) error {
			return ut.Add("mobile", "手机号不符合规范", true)
		},
		func(ut ut.Translator, fe validator.FieldError) string {
			t, _ := ut.T("mobile")
			return t
		},
	); nil != err {
		return
	}

	return
}

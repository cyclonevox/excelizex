package validatorx

import "github.com/go-playground/validator/v10"

func initValidation(validate *validator.Validate) (err error) {
	if err = validate.RegisterValidation("mobile", checkMobile); nil != err {
		return
	}
	if err = validate.RegisterValidation("without_special_symbol", checkWithoutSpecialSymbol); nil != err {
		return
	}
	if err = validate.RegisterValidation("id_card", checkIdCard); nil != err {
		return
	}
	if err = validate.RegisterValidation("id_card_15", checkIdCard15Len); nil != err {
		return
	}

	return
}

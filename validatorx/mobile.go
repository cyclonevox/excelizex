package validatorx

import (
	"regexp"

	"github.com/go-playground/validator/v10"
)

// CheckMobile 检查手机号
func CheckMobile(mobile string) bool {
	regular := `^[+]86[-]1([3,4,5,6,7,8,9][0-9])\d{8}$`
	reg := regexp.MustCompile(regular)

	return reg.MatchString(mobile)
}

func checkMobile(fl validator.FieldLevel) bool {
	return CheckMobile(fl.Field().String())
}

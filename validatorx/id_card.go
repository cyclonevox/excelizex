package validatorx

import (
	"regexp"
	"strconv"

	"github.com/go-playground/validator/v10"
)

const (
	_15LenIdCardRegex = `^[1-8]\d{7}(?:0\d|10|11|12)(?:0[1-9]|[1-2][\d]|30|31)\d{3}$`
	_18LenIdCardRegex = `^[1-8]\d{5}((((((19|20)\d{2})(0[13-9]|1[012])(0[1-9]|[12]\d|30))|(((19|20)\d{2})(0[13578]|1[02])31)|
					((19|20)\d{2})02(0[1-9]|1\d|2[0-8])|((((19|20)([13579][26]|[2468][048]|0[48]))|(2000))0229))\d{3})|
					((((\d{2})(0[13-9]|1[012])(0[1-9]|[12]\d|30))|((\d{2})(0[13578]|1[02])31)
					|((\d{2})02(0[1-9]|1\d|2[0-8]))|(([13579][26]|[2468][048]|0[048])0229))\d{2}))(\d|X|x)$`
)

// IdCard 18位身份证号码校验. 校验了年月日合法性，以及最后一位是否合法
func IdCard(str string) bool {
	b := regexp.MustCompile(_18LenIdCardRegex).MatchString(str)

	return b && checkLastNum(str)
}

// IdCard15Len 15位身份证号码校验，校验了年月日合法性.
func IdCard15Len(str string) bool {
	return regexp.MustCompile(_15LenIdCardRegex).MatchString(str)
}

// checkIdCard 判断身份证号码是否有效(仅15位)
func checkIdCard15Len(fl validator.FieldLevel) bool {
	valid := IdCard15Len(fl.Field().String())
	return valid
}

// checkIdCard 判断身份证号码是否有效(仅18位)
func checkIdCard(fl validator.FieldLevel) bool {
	valid := IdCard(fl.Field().String())
	return valid
}

var (
	// 根据 2^(index-1) % 11 得出位置:权重 哈希表 index从右至左算
	weight = []int{7, 9, 10, 5, 8, 4, 2, 1, 6, 3, 7, 9, 10, 5, 8, 4, 2}
	// 校验码值换算表
	codes = []string{"1", "0", "X", "9", "8", "7", "6", "5", "4", "3", "2"}
)

// checkLastNum 判断18位身份证号码的最后一位是否合法
func checkLastNum(idCard string) bool {
	sum := 0
	for i := 0; i < 17; i++ {
		n, _ := strconv.Atoi(idCard[i : i+1])
		sum += n * weight[i]
	}

	sum = sum % 11

	return codes[sum] == idCard[17:]
}

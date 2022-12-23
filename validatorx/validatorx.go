package validatorx

import (
	"github.com/go-playground/locales/en"
	"github.com/go-playground/locales/zh"
	"github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	zhLang "github.com/go-playground/validator/v10/translations/zh"
)

var v V

type V struct {
	validate *validator.Validate
	chinese  ut.Translator
}

// New 创建新的验证器
func New() *V {
	return &v
}

// 创建内置验证器
// 单例设计模式
func newValidate() (err error) {
	//声明翻译对象
	uni := ut.New(en.New(), zh.New())
	//设置翻译语言
	chinese, _ := uni.GetTranslator("zh")

	v = V{
		validate: validator.New(),
		chinese:  chinese,
	}

	if err = zhLang.RegisterDefaultTranslations(v.validate, v.chinese); nil != err {
		return
	}
	if err = initValidation(v.validate); nil != err {
		return
	}
	if err = initTranslation(v.validate, v.chinese); nil != err {
		return
	}

	return
}

func (v *V) Struct(i interface{}) (error, map[string]string) {

	if err := v.validate.Struct(i); err != nil {

		return err, err.(validator.ValidationErrors).Translate(v.chinese)
	}

	return nil, nil
}

func (v *V) Val(i interface{}, tag string) (error, map[string]string) {

	if err := v.validate.Var(i, tag); err != nil {

		return err, err.(validator.ValidationErrors).Translate(v.chinese)
	}

	return nil, nil
}

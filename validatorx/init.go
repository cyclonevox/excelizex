package validatorx

func init() {
	if err := newValidate(); nil != err {
		panic(err)
	}
}

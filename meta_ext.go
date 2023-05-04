package excelizex

type DefaultExtHeader struct {
	// 表头的值
	HeaderName string
	// 样式 用于存储样式Tag内容
	StyleTag string
	// 验证字段 用于存储validate tag中的数据
	ValidateTag string
}

func (d DefaultExtHeader) ExtHeader() string {
	return d.HeaderName
}

func (d DefaultExtHeader) ExtValidateTag() string {
	return d.ValidateTag
}

func (d DefaultExtHeader) ExtStyleTag() string {
	return d.StyleTag
}

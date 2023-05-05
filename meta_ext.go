package excelizex

type DefaultExtHeader struct {
	// 表头的值
	HeaderName string
	// 样式 用于存储样式Tag内容
	StyleTag string
	// 验证字段 用于存储validate tag中的数据
	ValidateTag string
	// 数据字段 用于存储数据 可直接生成在表中
	Data any
}

func (d DefaultExtHeader) ExtData() any {
	return d.Data
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

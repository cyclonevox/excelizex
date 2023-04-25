package excelizex

var defExtHeader = DefaultExtHeader{}

type DefaultExtHeader struct {
	// 部分
	ExtPart Part
	// 样式 用于存储样式Tag内容
	ExtStyleTag string
	// 验证字段 用于存储validate tag中的数据
	ExtValidateTag string
	// 列索引
	ExtColIndex int
	// Notice的值/表头的值
	ExtCellValue string
}

func (d DefaultExtHeader) Part() Part {
	return d.ExtPart
}

func (d DefaultExtHeader) ColIndex() int {
	return d.ExtColIndex
}

func (d DefaultExtHeader) CellValue() string {
	return d.ExtCellValue
}

func (d DefaultExtHeader) ValidateTag() string {
	return d.ExtValidateTag
}

func (d DefaultExtHeader) StyleTag() string {
	return d.ExtStyleTag
}

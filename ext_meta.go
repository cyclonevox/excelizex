package excelizex

type ExtHeader interface {
	HeaderName() string
	ValidateTag() string
	StyleTag() string
	Data() any
}

var defExtHeader = DefaultExtHeader{}

type DefaultExtHeader struct {
	name     string
	validate string
	style    string
	data     any
}

func (d DefaultExtHeader) HeaderName() string {
	return d.name
}

func (d DefaultExtHeader) ValidateTag() string {
	return d.validate
}

func (d DefaultExtHeader) StyleTag() string {
	return d.style
}

func (d DefaultExtHeader) Data() any {
	return d.data
}

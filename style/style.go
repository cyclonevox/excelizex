package style

import "github.com/xuri/excelize/v2"

type Payload struct {
	StyleID int
	Style
}

type Style interface {
	Name() string
	SetName(name string) Style
	Style() *excelize.Style

	Append(style Style) Style
}

type DefaultStyle struct {
	name  string
	style *styleMap
}

func (d *DefaultStyle) Append(style Style) Style {
	d.style.saveToMap(style.Style())
	d.name += "+" + style.Name()

	return d
}

func (d *DefaultStyle) Name() string {
	return d.name
}

func (d *DefaultStyle) Style() *excelize.Style {
	return d.style.mapToStyle()
}

func (d *DefaultStyle) SetName(name string) Style {
	d.name = name

	return d
}

func (d *DefaultStyle) initStyle(style *excelize.Style) {
	d.style.saveToMap(style)
}

func NewDefaultStyle(name string, style *excelize.Style) *DefaultStyle {
	ds := &DefaultStyle{
		name:  name,
		style: newStyleMap(),
	}

	ds.initStyle(style)

	return ds
}

var DefaultNoticeStyle = DefaultRedFont.Append(AlignmentWrapText).Append(DefaultLocked).SetName("default-notice")
var DefaultHeaderRedStyle = DefaultRedFont.Append(DefaultLocked).SetName("default-header-red")
var DefaultHeaderStyle = DefaultLocked.SetName("default-header")
var DefaultDataStyle = DefaultNumFmtText.SetName("default-all")

var DefaultRedFont = NewDefaultStyle("red-font", redFont)
var AlignmentWrapText = NewDefaultStyle("alignment", &excelize.Style{Alignment: &excelize.Alignment{WrapText: true}})
var DefaultNumFmtText = NewDefaultStyle("numFmtText", numFmtText)
var DefaultLocked = NewDefaultStyle("default-locked", locked)

var locked = &excelize.Style{Protection: &excelize.Protection{Locked: true}}
var numFmtText = &excelize.Style{NumFmt: 49}
var redFont = &excelize.Style{
	Font: &excelize.Font{
		Bold:   true,
		Family: "微软雅黑",
		Size:   11,
		Color:  "#FF0000",
	},
}

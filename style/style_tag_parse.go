package style

import (
	"strconv"
	"strings"
)

type TagParse string

type Parsed struct {
	StyleNames []string
	Cell       CellRange
	AutoWide   bool
}

type CellRange struct {
	StartCell Cell
	EndCell   Cell
}

type Cell struct {
	Col string
	Row int
}

func (c Cell) Format() string {
	return c.Col + strconv.FormatInt(int64(c.Row), 10)
}

func (t TagParse) Parse() (style Parsed) {
	p := strings.Split(string(t), " ")
	if len(p) == 1 || (len(p) == 2 && p[1] == "auto-wide") {
		style.AutoWide = true
	}

	if len(p) == 2 && p[1] == "no-auto-wide" {
		style.AutoWide = false
	}

	styleTags := strings.Split(p[0], ";")
	for _, s := range styleTags {
		style.parse(s)
	}

	return
}

func (p *Parsed) parse(style string) {
	styleParams := strings.Split(style, ",")
	styleList := strings.Split(styleParams[0], "+")

	p.StyleNames = styleList
}

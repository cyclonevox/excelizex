package style

import "github.com/xuri/excelize/v2"

var DefaultNoticeStyle = &excelize.Style{
	Font: &excelize.Font{
		Bold:   true,
		Family: "微软雅黑",
		Size:   11,
		Color:  "#FF0000",
	},
	Alignment:  &excelize.Alignment{WrapText: true},
	Protection: &excelize.Protection{Locked: true},
}

var NumFmtText = &excelize.Style{NumFmt: 49}

var StyleLocked = &excelize.Style{Protection: &excelize.Protection{Locked: true}}

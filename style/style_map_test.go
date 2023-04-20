package style

import (
	"testing"

	"github.com/xuri/excelize/v2"
)

func Test_Create_Style_Map(t *testing.T) {
	testStyle1 := &excelize.Style{
		Font: &excelize.Font{
			Bold:   true,
			Family: "微软雅黑",
			Size:   11,
			Color:  "#FF0000",
		},
	}

	testStyle2 := &excelize.Style{
		Alignment: &excelize.Alignment{WrapText: true},
	}

	testStyle3 := &excelize.Style{
		Alignment: &excelize.Alignment{WrapText: true},
	}

	testStyle4 := &excelize.Style{NumFmt: 49, Protection: &excelize.Protection{Locked: false}}

	t.Run("create_style_map", func(t *testing.T) {
		sm := newStyleMap()
		sm.saveToMap(testStyle1)
		style := sm.mapToStyle()
		t.Logf("style %+v", style)
		sm.saveToMap(testStyle2)
		style = sm.mapToStyle()
		t.Logf("style %+v", style)
		sm.saveToMap(testStyle3)
		style = sm.mapToStyle()
		t.Logf("style %+v", style)
		sm.saveToMap(testStyle4)
		style = sm.mapToStyle()
		t.Logf("style %+v", style)

		return
	})
}

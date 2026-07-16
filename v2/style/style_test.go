package style_test

import (
	"testing"

	"github.com/cyclonevox/excelizex/v2/style"
	"github.com/xuri/excelize/v2"
)

func TestAppendMerge(t *testing.T) {
	a := style.New("a", &excelize.Style{Font: &excelize.Font{Bold: true}})
	b := style.New("b", &excelize.Style{NumFmt: 49})
	merged := a.Append(b).(*style.DefaultStyle)
	es := merged.ExcelStyle()
	if es.Font == nil || !es.Font.Bold {
		t.Fatal("expected bold font")
	}
	if es.NumFmt != 49 {
		t.Fatalf("numfmt: %d", es.NumFmt)
	}
}

func TestRegistryResolveAppend(t *testing.T) {
	f := excelize.NewFile()
	reg := style.NewRegistry(f)
	_ = reg.RegisterDefaults()
	id, err := reg.Resolve("header", "locked")
	if err != nil {
		t.Fatal(err)
	}
	if id < 1 {
		t.Fatalf("style id: %d", id)
	}
	id2, err := reg.Resolve("header", "locked")
	if err != nil || id2 != id {
		t.Fatalf("cache miss: %d %d err=%v", id, id2, err)
	}
}

func TestSplitRole(t *testing.T) {
	parts := style.SplitRole([]string{"header-red+locked", "body"}, 0)
	if len(parts) != 2 || parts[0] != "header-red" || parts[1] != "locked" {
		t.Fatalf("parts: %v", parts)
	}
}

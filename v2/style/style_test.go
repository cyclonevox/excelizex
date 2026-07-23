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

func TestAppendDeepMergeKeepsFill(t *testing.T) {
	a := style.New("body-blue", &excelize.Style{
		NumFmt: 49,
		Fill: excelize.Fill{
			Type:    "pattern",
			Pattern: 1,
			Color:   []string{"#DAEEF3"},
		},
		Protection: &excelize.Protection{Locked: false},
	})
	b := style.New("locked", &excelize.Style{Protection: &excelize.Protection{Locked: true}})
	es := a.Append(b).ExcelStyle()
	if es.Fill.Type != "pattern" || len(es.Fill.Color) == 0 || es.Fill.Color[0] != "#DAEEF3" {
		t.Fatalf("fill lost: %+v", es.Fill)
	}
	if es.NumFmt != 49 {
		t.Fatalf("numfmt lost: %d", es.NumFmt)
	}
	if es.Protection == nil || !es.Protection.Locked {
		t.Fatalf("protection: %+v", es.Protection)
	}
}

func TestRegisterInvalidatesStyleIDCache(t *testing.T) {
	f := excelize.NewFile()
	reg := style.NewRegistry(f)
	_ = reg.Register(style.New("custom", &excelize.Style{NumFmt: 1}))
	id1, err := reg.Resolve("custom")
	if err != nil {
		t.Fatal(err)
	}
	_ = reg.Register(style.New("custom", &excelize.Style{NumFmt: 49}))
	id2, err := reg.Resolve("custom")
	if err != nil {
		t.Fatal(err)
	}
	if id1 == id2 {
		t.Fatal("expected style id cache invalidation after re-register")
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

func TestStyleBodyBluePlusLockedKeepsFillAndProtection(t *testing.T) {
	f := excelize.NewFile()
	t.Cleanup(func() { _ = f.Close() })
	reg := style.NewRegistry(f)
	if err := reg.RegisterDefaults(); err != nil {
		t.Fatal(err)
	}
	id, err := reg.Resolve("body-blue", "locked")
	if err != nil {
		t.Fatal(err)
	}
	if err := f.SetCellStyle("Sheet1", "A1", "A1", id); err != nil {
		t.Fatal(err)
	}
	styleID, err := f.GetCellStyle("Sheet1", "A1")
	if err != nil {
		t.Fatal(err)
	}
	st, err := f.GetStyle(styleID)
	if err != nil {
		t.Fatal(err)
	}
	if st.Fill.Type == "" || len(st.Fill.Color) == 0 || st.Fill.Color[0] == "" {
		t.Fatalf("expected body-blue fill retained, got %+v", st.Fill)
	}
	if st.NumFmt != 49 {
		t.Fatalf("expected NumFmt 49, got %d", st.NumFmt)
	}
	if st.Protection == nil || !st.Protection.Locked {
		t.Fatalf("expected locked protection, got %+v", st.Protection)
	}
}

package convert_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/cyclonevox/excelizex/v2/convert"
)

func TestFromBuiltin(t *testing.T) {
	reg := convert.ExportRegistry{}
	v := reflect.ValueOf(true)
	s, err := convert.From(v, "", "", reg)
	if err != nil || s != "是" {
		t.Fatalf("bool: %q err=%v", s, err)
	}
	tm := time.Date(2024, 9, 1, 0, 0, 0, 0, time.UTC)
	s, err = convert.From(reflect.ValueOf(tm), "", "2006-01-02", reg)
	if err != nil || s != "2024-09-01" {
		t.Fatalf("time: %q err=%v", s, err)
	}
}

func TestExportTo(t *testing.T) {
	reg := convert.ExportRegistry{}
	convert.ExportTo(reg, "x", func(n int) (string, error) {
		return "n", nil
	})
	s, err := reg["x"](42)
	if err != nil || s != "n" {
		t.Fatalf("export: %q err=%v", s, err)
	}
}

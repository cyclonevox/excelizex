package convert_test

import (
	"fmt"
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

func TestExportToString(t *testing.T) {
	reg := convert.ExportRegistry{}
	convert.ExportToString(reg, "upper", func(s string) (string, error) {
		if s == "" {
			return "(empty)", nil
		}

		return s + "!", nil
	})

	tests := []struct {
		name string
		in   any
		want string
	}{
		{name: "value", in: "hi", want: "hi!"},
		{name: "nil", in: nil, want: "(empty)"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := reg["upper"](tt.in)
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Fatalf("got %q want %q", got, tt.want)
			}
		})
	}
}

func TestExportToEdgeCases(t *testing.T) {
	reg := convert.ExportRegistry{}
	convert.ExportTo(reg, "id", func(n int) (string, error) {
		return fmt.Sprintf("%d", n), nil
	})

	s, err := reg["id"]((*int)(nil))
	if err != nil || s != "0" {
		t.Fatalf("nil ptr: %q err=%v", s, err)
	}

	convert.ExportTo(reg, "strict", func(n int) (string, error) {
		return "ok", nil
	})
	if _, err := reg["strict"]("not-int"); err == nil {
		t.Fatal("expected type mismatch error")
	}
}

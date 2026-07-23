package convert_test

import (
	"encoding"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/cyclonevox/excelizex/v2/convert"
)

type customInt int

func (c *customInt) UnmarshalText(text []byte) error {
	n, err := strconv.Atoi(string(text))
	if err != nil {
		return err
	}
	*c = customInt(n)

	return nil
}

var _ encoding.TextUnmarshaler = (*customInt)(nil)

type grade int

func (g *grade) UnmarshalText(text []byte) error {
	switch string(text) {
	case "A":
		*g = 1
		return nil
	case "B":
		*g = 2
		return nil
	default:
		return strconv.ErrSyntax
	}
}

var _ encoding.TextUnmarshaler = (*grade)(nil)

func TestBuiltinTypes(t *testing.T) {
	tests := []struct {
		name   string
		raw    string
		target any
		want   any
	}{
		{"string", "hello", "", "hello"},
		{"int", "42", 0, 42},
		{"int bad", "x", 0, nil},
		{"float", "3.14", 0.0, 3.14},
		{"bool true", "是", false, true},
		{"bool false", "否", false, false},
		{"bool bad", "maybe", false, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dst := reflect.New(reflect.TypeOf(tt.target)).Elem()
			err := convert.To(tt.raw, dst, "")
			if tt.want == nil {
				if err == nil {
					t.Fatal("expected error")
				}

				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if dst.Interface() != tt.want {
				t.Fatalf("got %v want %v", dst.Interface(), tt.want)
			}
		})
	}
}

func TestTimeLayout(t *testing.T) {
	var dst time.Time
	v := reflect.ValueOf(&dst).Elem()
	if err := convert.To("2024-06-01", v, "2006-01-02"); err != nil {
		t.Fatal(err)
	}
	if dst.Year() != 2024 || dst.Month() != time.June {
		t.Fatalf("time: %v", dst)
	}
}

func TestTextUnmarshaler(t *testing.T) {
	var c customInt
	v := reflect.ValueOf(&c).Elem()
	if err := convert.To("7", v, ""); err != nil {
		t.Fatal(err)
	}
	if c != 7 {
		t.Fatalf("got %d", c)
	}
}

func TestTextUnmarshalerGrade(t *testing.T) {
	var g grade
	v := reflect.ValueOf(&g).Elem()
	if err := convert.To("A", v, ""); err != nil {
		t.Fatal(err)
	}
	if g != 1 {
		t.Fatalf("got %d", g)
	}
}

func TestBuiltinUintAndTime(t *testing.T) {
	var u uint
	v := reflect.ValueOf(&u).Elem()
	if err := convert.To("99", v, ""); err != nil {
		t.Fatal(err)
	}
	if u != 99 {
		t.Fatalf("uint: %d", u)
	}

	var tm time.Time
	tv := reflect.ValueOf(&tm).Elem()
	if err := convert.To("", tv, ""); err != nil {
		t.Fatal(err)
	}
	if !tm.IsZero() {
		t.Fatalf("empty time: %v", tm)
	}
	if err := convert.To("2024/06/01", tv, ""); err != nil {
		t.Fatal(err)
	}
	if tm.Year() != 2024 || tm.Month() != time.June {
		t.Fatalf("common layout time: %v", tm)
	}
}

func TestAssignConvertible(t *testing.T) {
	var dst int
	v := reflect.ValueOf(&dst).Elem()
	if err := convert.To("3", v, ""); err != nil {
		t.Fatal(err)
	}
	if dst != 3 {
		t.Fatalf("got %d", dst)
	}
}

package convert_test

import (
	"encoding"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/cyclonevox/excelizex/v2/convert"
)

func TestFromBuiltin(t *testing.T) {
	v := reflect.ValueOf(true)
	s, err := convert.From(v, "")
	if err != nil || s != "是" {
		t.Fatalf("bool: %q err=%v", s, err)
	}
	tm := time.Date(2024, 9, 1, 0, 0, 0, 0, time.UTC)
	s, err = convert.From(reflect.ValueOf(tm), "2006-01-02")
	if err != nil || s != "2024-09-01" {
		t.Fatalf("time: %q err=%v", s, err)
	}
}

type gradeLabel int

func (g gradeLabel) MarshalText() ([]byte, error) {
	return []byte(fmt.Sprintf("G%d", g)), nil
}

var _ encoding.TextMarshaler = gradeLabel(0)

func TestTextMarshaler(t *testing.T) {
	s, err := convert.From(reflect.ValueOf(gradeLabel(2)), "")
	if err != nil {
		t.Fatal(err)
	}
	if s != "G2" {
		t.Fatalf("got %q want G2", s)
	}
}

type gradeChar int

func (g gradeChar) MarshalText() ([]byte, error) {
	switch g {
	case 1:
		return []byte("A"), nil
	case 2:
		return []byte("B"), nil
	default:
		return nil, fmt.Errorf("unknown grade %d", g)
	}
}

func TestTextMarshalerGrade(t *testing.T) {
	s, err := convert.From(reflect.ValueOf(gradeChar(1)), "")
	if err != nil {
		t.Fatal(err)
	}
	if s != "A" {
		t.Fatalf("got %q want A", s)
	}
}

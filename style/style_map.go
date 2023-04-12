package style

import (
	"encoding/json"
	"github.com/xuri/excelize/v2"
	"reflect"
)

type styleMap struct {
	m map[string]interface{}
}

func newStyleMap() *styleMap {
	return &styleMap{m: map[string]interface{}{}}
}

// saveToMap will not change the saved params
func (s *styleMap) saveToMap(style *excelize.Style) {
	if s.m == nil {
		panic("call newStyleMap() first")
	}
	if style == nil {
		panic("style nil")
	}

	var (
		marshal []byte
		err     error
	)
	styleCache := map[string]interface{}{}
	if marshal, err = json.Marshal(style); err != nil {
		return
	}
	if err = json.Unmarshal(marshal, &styleCache); err != nil {
		panic(err)
	}

	for k, v := range styleCache {
		if v == nil {
			continue
		}

		vValue := reflect.ValueOf(v)
		if vValue.Kind() == reflect.Pointer {
			vValue = vValue.Elem()
		}

		if !vValue.IsZero() {
			s.m[k] = v
		}
	}

	return
}

func (s *styleMap) mapToStyle() (style *excelize.Style) {
	var (
		marshal []byte
		err     error
	)
	if marshal, err = json.Marshal(s.m); err != nil {
		panic(err)
	}

	style = new(excelize.Style)
	if err = json.Unmarshal(marshal, style); err != nil {
		panic(err)
	}

	return
}

func (s *styleMap) addBorder(border []excelize.Border)            { s.m["border"] = border }
func (s *styleMap) addFill(fill excelize.Fill)                    { s.m["fill"] = fill }
func (s *styleMap) addFont(font *excelize.Font)                   { s.m["font"] = font }
func (s *styleMap) addAlignment(alignment *excelize.Alignment)    { s.m["alignment"] = alignment }
func (s *styleMap) addProtection(protection *excelize.Protection) { s.m["protection"] = protection }
func (s *styleMap) addNumFmt(numFmt int)                          { s.m["number_format"] = numFmt }
func (s *styleMap) addDecimalPlaces(decimalPlaces int)            { s.m["decimal_places"] = decimalPlaces }
func (s *styleMap) lang(lang string)                              { s.m["lang"] = lang }
func (s *styleMap) negRed(negred bool)                            { s.m["negred"] = negred }
func (s *styleMap) addCustomNumFmt(customNumberFormat *string) {
	s.m["custom_number_format"] = customNumberFormat
}

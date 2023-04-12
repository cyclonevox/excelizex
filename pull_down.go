package excelizex

import "reflect"

// 下拉结构，包含对应的列和具体的选项
type pullDown struct {
	target  map[string]*PullDownData
	options []*PullDownData
}

type PullDownData struct {
	col  string
	data []any
}

func newPullDown() *pullDown {
	return &pullDown{target: make(map[string]*PullDownData)}
}

// Merge 合并一个下拉对象至目标下拉对象中
func (p *pullDown) merge(pd *pullDown) *pullDown {
	for axis, pullDownData := range pd.target {
		p.addOptions(axis, pullDownData.data)
	}

	return p
}

func (p *pullDown) addOptions(col string, options any) *pullDown {
	var (
		typ  = reflect.TypeOf(options)
		val  = reflect.ValueOf(options)
		list []any
	)
	if typ.Kind() == reflect.Slice {
		for i := 0; i < val.Len(); i++ {
			list = append(list, val.Index(i).Interface())
		}
	} else {
		list = append(list, val.Interface())
	}

	if pd, ok := p.target[col]; ok {
		pd.data = append(pd.data, list...)
	} else {
		pdd := &PullDownData{col: col, data: list}
		p.options = append(p.options, pdd)
		p.target[col] = pdd
	}

	return p
}

func (p *pullDown) sheet(name string) *sheet {
	return NewSheet(name + OptionsSaveTable).initSheet(p.data())
}

func (p *pullDown) data() [][]any {
	d := make([][]any, 0, len(p.options))

	for _, o := range p.options {
		d = append(d, o.data)
	}

	return d
}

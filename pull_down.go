package excelizex

// 下拉结构，包含对应的列和具体的选项
type pullDown struct {
	target  map[string]*PullDownData
	options []*PullDownData
}

type PullDownData struct {
	col  string
	data []any
}

func NewPullDown() *pullDown {
	return &pullDown{target: make(map[string]*PullDownData)}
}

// Merges 合并多个下拉对象至目标下拉对象中
func (p *pullDown) Merges(pds []*pullDown) *pullDown {
	for _, pd := range pds {
		p.Merge(pd)
	}

	return p
}

// Merge 合并一个下拉对象至目标下拉对象中
func (p *pullDown) Merge(pd *pullDown) *pullDown {
	for axis, pullDownData := range pd.target {
		p.AddOptions(axis, pullDownData.data)
	}

	return p
}

func (p *pullDown) AddOptions(col string, options []any) *pullDown {
	for _, o := range options {
		p.AddOption(col, o)
	}

	return p
}

func (p *pullDown) AddOption(col string, option any) {
	if pd, ok := p.target[col]; ok {
		pd.data = append(pd.data, option)
	} else {
		pdd := &PullDownData{col: col, data: []any{option}}
		p.options = append(p.options, pdd)
		p.target[col] = pdd
	}
}

func (p *pullDown) sheet(name string) *Sheet {
	return NewSheet(Name(name + OptionsSaveTable)).SetData(p.data())
}

func (p *pullDown) data() [][]any {
	d := make([][]any, 0, len(p.options))

	for _, o := range p.options {
		d = append(d, o.data)
	}

	return d
}

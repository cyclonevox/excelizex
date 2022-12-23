package excelizex

type Sheet struct {
	// 表名
	Name string `json:"name"`
	// 顶栏提示
	Notice string `json:"notice"`
	// 表头
	Header []string `json:"header"`
	// 数据
	Data [][]any `json:"data"`
}

func NewSheet(options ...SheetOption) *Sheet {
	sheet := new(Sheet)
	for _, option := range options {
		option(sheet)
	}

	return sheet
}

func (s *Sheet) SetName(name string) {
	s.Name = name
}

func (s *Sheet) SetNotice(notice string) {
	s.Notice = notice
}

func (s *Sheet) SetHeader(header []string) {
	s.Header = header
}

func (s *Sheet) SetData(data [][]any) {
	s.Data = data
}

func (s *Sheet) Excel() *excel {
	if s.Name == "" {
		panic("need a sheet name at least")
	}

	return New().AddSheets(*s)
}

type SheetOption = func(*Sheet)

func SetName(name string) SheetOption {
	return func(sheet *Sheet) {
		sheet.Name = name
	}
}

func SetHeader(header []string) SheetOption {
	return func(s *Sheet) {
		s.Header = header
	}
}

func SetData(data [][]any) SheetOption {
	return func(s *Sheet) {
		s.Data = data
	}
}

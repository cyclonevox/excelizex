package excelizex

type Result struct {
	sector       int
	dataStartRow int
	SheetName    string
	Notice       []string
	Header       []string
	HeaderNotice [][]string
	Errors       []ErrorInfo
}

type ErrorInfo struct {
	ErrorRow  int
	RawData   []string
	ErrorInfo []string
}

func (r *Result) Next() bool {
	return len(r.Errors)+len(r.HeaderNotice) >= r.sector+1
}

func (r *Result) Data() (data any) {
	if r.sector < len(r.HeaderNotice) {
		data = r.HeaderNotice[r.sector]
	} else {
		data = append(r.Errors[r.sector-r.dataStartRow+1].RawData, r.Errors[r.sector-r.dataStartRow+1].ErrorInfo...)
	}

	r.sector++
	return data
}

func (r *Result) Close() error {
	r.sector = 0

	return nil
}

// SetResults 该方法会清除原始的表的数据。并将错误数据保留以及将错误原因写入原文件中
func (f *File) SetResults(result *Result) (file *File, exist bool) {
	if result.dataStartRow == 0 || len(result.Errors) == 0 {
		return nil, false
	}
	// 流式导入数据
	if err := f.WriteInByStream(result, 1); err != nil {
		panic(err)
	}

	return f, true
}

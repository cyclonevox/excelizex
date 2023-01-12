package excelizex

type Result struct {
	sector       int
	dataStartRow int
	SheetName    string
	Notice       string
	Header       []string
	Errors       []ErrorInfo
}

type ErrorInfo struct {
	ErrorRow  int
	RawData   []string
	ErrorInfo []string
}

func (r *Result) Next() bool {
	return len(r.Errors) >= r.sector+1
}

func (r *Result) Data() (data any) {
	data = append(r.Errors[r.sector].RawData, r.Errors[r.sector].ErrorInfo...)
	r.sector++
	return data
}

func (r *Result) Close() error {
	r.sector = 0

	return nil
}

// SetResults 该方法会清除原始的表。并将错误数据保留以及将错误原因写入原文件中
func (f *File) SetResults(result *Result) *File {
	// 流式导入数据
	if err := f.WriteInByStream(result, result.dataStartRow); err != nil {
		panic(err)
	}

	return f
}

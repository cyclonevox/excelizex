package excelizex

type Result struct {
	sector       int
	DataStartRow int
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
	r.sector++
	return len(r.Errors) >= r.sector
}

func (r *Result) Data() any {
	info := r.Errors[r.sector]
	return append(info.RawData, info.ErrorInfo...)
}

func (r *Result) Close() error {
	r.sector = 0

	return nil
}

// SetResults 该方法会清除已经导入成功的数据。并将错误数据保留以及将错误原因写入原文件中
func (f *file) SetResults(sheetName string, result *Result) {
	// 去除
	f.excel().DeleteSheet(sheetName)

	f.StreamImport(result, SetName(sheetName), SetHeader(append(result.Header, "错误原因")), SetNotice(result.Notice))
}

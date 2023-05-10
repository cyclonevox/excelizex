package excelizex

import (
	"strconv"
	"sync"

	"github.com/xuri/excelize/v2"
)

type Result struct {
	totalRow     int
	errorRow     int
	dataStartRow int
	SheetName    string
	Notice       string
	Header       []string

	mu     sync.Mutex
	errors ErrorInfos
}

type ErrorInfo struct {
	ErrorRow int
	RawData  []string
	Messages []string
}

type ErrorInfos []ErrorInfo

func (r *Result) addError(info ErrorInfo) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.errors = append(r.errors, info)
}

func (f *File) removeDataLine(results *Result) (err error) {
	var rows *excelize.Rows
	if rows, err = f.excel().Rows(results.SheetName); err != nil {
		return
	}
	defer rows.Close()

	var i int
	for rows.Next() {
		i++
		if i >= results.dataStartRow {
			if err = f.excel().RemoveRow(results.SheetName, results.dataStartRow); err != nil {
				return
			}
		}
	}

	return
}

// SetResults 该方法会清除原始的表的数据。并将错误数据保留以及将错误原因写入原文件中
func (f *File) SetResults(result *Result) (file *File, exist bool, err error) {
	if result.dataStartRow == 0 || len(result.errors) == 0 {
		return
	} else {
		exist = true
	}

	// 移除所有行
	if err = f.removeDataLine(result); err != nil {
		return
	}

	// 设置头部行
	if result.Header[len(result.Header)-1] != "错误原因" {
		result.Header = append(result.Header, "错误原因")
		rowName := "A" + strconv.FormatInt(int64(result.dataStartRow-1), 10)
		if err = f.excel().SetSheetRow(result.SheetName, rowName, &result.Header); err != nil {
			return
		}
	}

	var errorInfoColumnName string
	if errorInfoColumnName, err = excelize.ColumnNumberToName(len(result.Header)); err != nil {
		return
	}

	var lastRow int
	for index, errorInfo := range result.errors {
		if lastRow == errorInfo.ErrorRow {
			continue
		} else {
			lastRow = errorInfo.ErrorRow
		}

		cellName := "A" + strconv.FormatInt(int64(result.dataStartRow+index), 10)
		if err = f.excel().SetSheetRow(result.SheetName, cellName, &errorInfo.RawData); err != nil {
			return
		}

		errorInfoCellName := errorInfoColumnName + strconv.FormatInt(int64(result.dataStartRow+index), 10)
		if err = f.excel().SetSheetRow(result.SheetName, errorInfoCellName, &errorInfo.Messages); err != nil {
			return
		}
	}

	file = f

	return
}

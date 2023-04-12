package excelizex

import (
	"errors"
	"github.com/xuri/excelize/v2"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
)

type Context interface {
	FormFile(name string) (*multipart.FileHeader, error)
}

func ExtractFromContext(ctx Context, password ...string) (excel *File, err error) {
	var fileHeader *multipart.FileHeader
	if fileHeader, err = ctx.FormFile("file"); err != nil {
		return
	}

	filename := fileHeader.Filename
	isXlsx := strings.HasSuffix(filename, ".xlsx")
	if !isXlsx {
		err = errors.New("文件格式不正确，请上传.xlsx文件")

		return
	}

	var file multipart.File
	if file, err = fileHeader.Open(); err != nil {
		return
	}
	defer file.Close()

	if excel, err = newExcelFormIo(file); err != nil {
		return
	}

	if len(password) > 0 {
		if _, err = excel.Unlock(password[0]); nil != err {
			err = errors.New("password wrong")

			return
		}
	}

	return
}

func ExtractFromUrl(FileURL string, password ...string) (excel *File, err error) {
	var resp *http.Response
	if resp, err = http.Get(FileURL); err != nil {
		return
	}
	if excel, err = newExcelFormIo(resp.Body); err != nil {
		return
	}

	if len(password) > 0 {
		if _, err = excel.Unlock(password[0]); nil != err {
			err = errors.New("password wrong")

			return
		}
	}

	return
}

func newExcelFormIo(reader io.Reader) (*File, error) {
	if f, err := excelize.OpenReader(reader); err != nil {
		return nil, err
	} else {
		return &File{_excel: f}, nil
	}
}

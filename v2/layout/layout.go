package layout

import (
	"fmt"
	"strings"
)

// Layout describes physical sheet layout: notice, header rows, and data start.
type Layout interface {
	NoticeRow() (row int, ok bool)
	HeaderRows() (start, end int)
	DataStartRow() int
	ResolveHeaders(headerRows [][]string) (map[int]string, error)
}

// NoticeHeaderData: row1 notice (optional), row2 header, row3+ data.
type NoticeHeaderData struct{}

func (NoticeHeaderData) NoticeRow() (int, bool) { return 1, true }

func (NoticeHeaderData) HeaderRows() (int, int) { return 2, 2 }

func (NoticeHeaderData) DataStartRow() int { return 3 }

func (NoticeHeaderData) ResolveHeaders(headerRows [][]string) (map[int]string, error) {
	return singleRowHeaders(headerRows)
}

// HeaderData: row1 header, row2+ data, no notice row.
type HeaderData struct{}

func (HeaderData) NoticeRow() (int, bool) { return 0, false }

func (HeaderData) HeaderRows() (int, int) { return 1, 1 }

func (HeaderData) DataStartRow() int { return 2 }

func (HeaderData) ResolveHeaders(headerRows [][]string) (map[int]string, error) {
	return singleRowHeaders(headerRows)
}

func singleRowHeaders(headerRows [][]string) (map[int]string, error) {
	if len(headerRows) != 1 {
		return nil, fmt.Errorf("layout: expected 1 header row, got %d", len(headerRows))
	}
	out := make(map[int]string)
	for i, cell := range headerRows[0] {
		h := strings.TrimSpace(cell)
		if h == "" {
			continue
		}
		out[i] = h
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("layout: empty header row")
	}

	return out, nil
}

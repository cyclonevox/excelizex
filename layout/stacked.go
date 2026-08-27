package layout

// Stacked is a multi-row header band. Binding uses the last non-empty cell
// in each column (so a group title row above leaf names still binds on the names).
//
// Count is the number of header rows; 0 means 2.
type Stacked struct {
	HasNotice bool
	Count     int
}

func (s Stacked) headerCount() int {
	if s.Count < 1 {
		return 2
	}

	return s.Count
}

func (s Stacked) NoticeRow() (int, bool) {
	if s.HasNotice {
		return 1, true
	}

	return 0, false
}

func (s Stacked) HeaderRows() (start, end int) {
	n := s.headerCount()
	if s.HasNotice {
		return 2, 1 + n
	}

	return 1, n
}

func (s Stacked) DataStartRow() int {
	_, end := s.HeaderRows()

	return end + 1
}

func (s Stacked) ResolveHeaders(headerRows [][]string) (map[int]string, error) {
	return resolveLastNonEmpty(headerRows)
}

func (s Stacked) RenderHeaders(headers []string) ([][]any, error) {
	n := s.headerCount()
	rows := make([][]any, n)
	for i := 0; i < n-1; i++ {
		rows[i] = make([]any, len(headers))
	}
	leaf := make([]any, len(headers))
	for i, h := range headers {
		leaf[i] = h
	}
	rows[n-1] = leaf

	return rows, nil
}

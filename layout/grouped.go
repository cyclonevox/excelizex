package layout

// Group describes a top-row title covering Span consecutive leaf columns
// in schema (then extra) column order.
type Group struct {
	Title string
	Span  int
}

// Grouped is notice (optional) + group title row + leaf header row + data.
// Adjacent cells that share a group title are merged on write when the layout
// is used with excelizex.Write.
type Grouped struct {
	HasNotice bool
	Groups    []Group
}

func (g Grouped) stacked() Stacked {
	return Stacked{HasNotice: g.HasNotice, Count: 2}
}

func (g Grouped) NoticeRow() (int, bool) { return g.stacked().NoticeRow() }

func (g Grouped) HeaderRows() (start, end int) { return g.stacked().HeaderRows() }

func (g Grouped) DataStartRow() int { return g.stacked().DataStartRow() }

func (g Grouped) ResolveHeaders(headerRows [][]string) (map[int]string, error) {
	return resolveLastNonEmpty(headerRows)
}

func (g Grouped) RenderHeaders(headers []string) ([][]any, error) {
	groupRow := make([]any, len(headers))
	leafRow := make([]any, len(headers))
	col := 0
	for _, grp := range g.Groups {
		span := grp.Span
		if span < 1 {
			span = 1
		}
		for i := 0; i < span && col < len(headers); i++ {
			if i == 0 {
				groupRow[col] = grp.Title
			}
			leafRow[col] = headers[col]
			col++
		}
	}
	for col < len(headers) {
		leafRow[col] = headers[col]
		col++
	}

	return [][]any{groupRow, leafRow}, nil
}

func (g Grouped) HeaderMerges(headerStart, nCols int) []Merge {
	if nCols < 1 {
		return nil
	}
	var merges []Merge
	col := 0
	for _, grp := range g.Groups {
		span := grp.Span
		if span < 1 {
			span = 1
		}
		if col >= nCols {
			break
		}
		end := col + span - 1
		if end >= nCols {
			end = nCols - 1
		}
		if end > col && grp.Title != "" {
			merges = append(merges, Merge{Row: headerStart, From: col, To: end})
		}
		col = end + 1
	}

	return merges
}

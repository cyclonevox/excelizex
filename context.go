package excelizex

import "context"

// Context carries per-row metadata for business callbacks.
type Context struct {
	context.Context
	Row int // 1-based Excel row number
}

// WithRow returns a Context with the given row number.
func WithRow(ctx context.Context, row int) Context {
	if ctx == nil {
		ctx = context.Background()
	}

	return Context{Context: ctx, Row: row}
}

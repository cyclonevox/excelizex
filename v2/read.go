package excelizex

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/cyclonevox/excelizex/v2/bind"
	"github.com/cyclonevox/excelizex/v2/convert"
	"github.com/cyclonevox/excelizex/v2/layout"
	"github.com/cyclonevox/excelizex/v2/schema"
	"golang.org/x/sync/errgroup"
)

// Validator validates a bound row before Collect/Each invokes the business callback.
// The library does not interpret validate struct tags; wire your project's validator here.
//
// Row is passed by pointer (&row) so validators using Struct() (e.g. playground/validator)
// can set fields on the bound value.
type Validator interface {
	Validate(row any) error
}

// ReadBuilder configures and executes a typed read pipeline.
type ReadBuilder[T any] struct {
	sheet       *Sheet
	converters  convert.Registry
	validators  []Validator
	concurrency int
	failFast    bool
}

// Convert registers a named converter for this read.
func (r *ReadBuilder[T]) Convert(name string, fn convert.ConvertFunc) *ReadBuilder[T] {
	if r.converters == nil {
		r.converters = make(convert.Registry)
	}
	r.converters[name] = fn

	return r
}

// Validate adds row validators.
func (r *ReadBuilder[T]) Validate(v ...Validator) *ReadBuilder[T] {
	r.validators = append(r.validators, v...)

	return r
}

// Collect reads all rows into a slice (sequential, safe for aggregation).
func (r *ReadBuilder[T]) Collect(ctx context.Context) ([]T, *Result, error) {
	sc, err := r.schemaForRead()
	if err != nil {
		return nil, nil, err
	}
	lyt := r.sheet.layout
	if lyt == nil {
		lyt = layout.NoticeHeaderData{}
	}
	rows, err := r.sheet.wb.getRowsLocked(r.sheet.name)
	if err != nil {
		return nil, nil, fmt.Errorf("read: get rows: %w", err)
	}
	res := newResult(r.sheet.name, lyt)
	if err := r.checkNotice(rows, lyt, sc, res); err != nil {
		return nil, res, err
	}
	headerStart, headerEnd := lyt.HeaderRows()
	headerRows, err := sliceRows(rows, headerStart, headerEnd)
	if err != nil {
		return nil, res, err
	}
	headerMap, err := lyt.ResolveHeaders(headerRows)
	if err != nil {
		return nil, res, fmt.Errorf("read: resolve headers: %w", err)
	}
	mapping, err := bind.MatchColumns(sc, headerMap)
	if err != nil {
		return nil, res, fmt.Errorf("read: %w", err)
	}
	res.setHeaders(mapping.Headers)
	reg := convert.Registry(r.converters)
	dataStart := lyt.DataStartRow()
	var out []T
	for i := dataStart - 1; i < len(rows); i++ {
		if err := ctx.Err(); err != nil {
			return out, res, err
		}
		cells := rows[i]
		if isEmptyRow(cells) {
			continue
		}
		rowNum := i + 1
		row, err := bind.BindRow[T](mapping, cells, reg)
		if err != nil {
			res.addError(rowNum, cells, err.Error())
			if r.failFast {
				return out, res, err
			}
			continue
		}
		for _, v := range r.validators {
			if err := v.Validate(&row); err != nil {
				res.addError(rowNum, cells, err.Error())
				if r.failFast {
					return out, res, err
				}
				goto nextRow
			}
		}
		out = append(out, row)
	nextRow:
	}

	return out, res, nil
}

// Each invokes fn for each successfully bound and validated row.
func (r *ReadBuilder[T]) Each(ctx context.Context, fn func(Context, T) error, opts ...EachOption) (*Result, error) {
	cfg := applyReadOptions(opts)
	cfg.concurrency = resolveConcurrency(cfg, r.concurrency)
	if r.failFast {
		cfg.failFast = true
	}

	sc, err := r.schemaForRead()
	if err != nil {
		return nil, err
	}

	lyt := r.sheet.layout
	if lyt == nil {
		lyt = layout.NoticeHeaderData{}
	}

	rows, err := r.sheet.wb.getRowsLocked(r.sheet.name)
	if err != nil {
		return nil, fmt.Errorf("read: get rows: %w", err)
	}

	res := newResult(r.sheet.name, lyt)

	if err := r.checkNotice(rows, lyt, sc, res); err != nil {
		return res, err
	}

	headerStart, headerEnd := lyt.HeaderRows()
	headerRows, err := sliceRows(rows, headerStart, headerEnd)
	if err != nil {
		return res, err
	}
	headerMap, err := lyt.ResolveHeaders(headerRows)
	if err != nil {
		return res, fmt.Errorf("read: resolve headers: %w", err)
	}

	mapping, err := bind.MatchColumns(sc, headerMap)
	if err != nil {
		return res, fmt.Errorf("read: %w", err)
	}
	res.setHeaders(mapping.Headers)

	reg := convert.Registry(r.converters)
	dataStart := lyt.DataStartRow()

	type job struct {
		rowNum int
		cells  []string
	}
	var jobs []job
	for i := dataStart - 1; i < len(rows); i++ {
		cells := rows[i]
		if isEmptyRow(cells) {
			continue
		}
		jobs = append(jobs, job{rowNum: i + 1, cells: cells})
	}

	var mu sync.Mutex
	eg, egCtx := errgroup.WithContext(ctx)
	// Always set a limit: errgroup's zero value is unlimited concurrency.
	eg.SetLimit(cfg.concurrency)

	for _, j := range jobs {
		j := j
		eg.Go(func() error {
			if err := egCtx.Err(); err != nil {
				return err
			}
			row, err := bind.BindRow[T](mapping, j.cells, reg)
			if err != nil {
				mu.Lock()
				res.addError(j.rowNum, j.cells, err.Error())
				mu.Unlock()
				if cfg.failFast {
					return err
				}

				return nil
			}
			for _, v := range r.validators {
				if err := v.Validate(&row); err != nil {
					mu.Lock()
					res.addError(j.rowNum, j.cells, err.Error())
					mu.Unlock()
					if cfg.failFast {
						return err
					}

					return nil
				}
			}
			if fn == nil {
				return nil
			}
			if err := fn(WithRow(egCtx, j.rowNum), row); err != nil {
				mu.Lock()
				res.addError(j.rowNum, j.cells, err.Error())
				mu.Unlock()
				if cfg.failFast {
					return err
				}
			}

			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		return res, err
	}

	return res, nil
}

// EachMap maps each row to M before invoking the callback.
func EachMap[T, M any](r *ReadBuilder[T], ctx context.Context, mapFn func(T) (M, error), fn func(Context, M) error, opts ...EachOption) (*Result, error) {
	return r.Each(ctx, func(c Context, row T) error {
		mapped, err := mapFn(row)
		if err != nil {
			return err
		}

		return fn(c, mapped)
	}, opts...)
}

func (r *ReadBuilder[T]) schemaForRead() (schema.Schema, error) {
	if r.sheet.schema != nil {
		return *r.sheet.schema, nil
	}
	var zero T

	return schema.New(zero)
}

func (r *ReadBuilder[T]) checkNotice(rows [][]string, lyt layout.Layout, sc schema.Schema, res *Result) error {
	noticeRow, ok := lyt.NoticeRow()
	if !ok {
		return nil
	}
	if noticeRow < 1 || noticeRow > len(rows) {
		if r.sheet.notice != "" {
			return fmt.Errorf("read: missing notice row")
		}

		return nil
	}
	text := strings.TrimSpace(joinRow(rows[noticeRow-1]))
	res.setNotice(text)
	if r.sheet.notice != "" && text != r.sheet.notice {
		return fmt.Errorf("read: notice mismatch: got %q want %q", text, r.sheet.notice)
	}

	return nil
}

func sliceRows(rows [][]string, start, end int) ([][]string, error) {
	if start < 1 || end < start {
		return nil, fmt.Errorf("read: invalid header row range %d-%d", start, end)
	}
	if end > len(rows) {
		return nil, fmt.Errorf("read: header rows out of range")
	}

	return rows[start-1 : end], nil
}

func isEmptyRow(cells []string) bool {
	for _, c := range cells {
		if strings.TrimSpace(c) != "" {
			return false
		}
	}

	return true
}

func joinRow(cells []string) string {
	return strings.TrimSpace(strings.Join(cells, ""))
}

// SetConcurrency sets default concurrency for Each (can be overridden by option).
func (r *ReadBuilder[T]) SetConcurrency(n int) *ReadBuilder[T] {
	if n < 1 {
		n = 1
	}
	r.concurrency = n

	return r
}

// SetFailFast enables fail-fast mode for this builder.
func (r *ReadBuilder[T]) SetFailFast() *ReadBuilder[T] {
	r.failFast = true

	return r
}

// ConvertTo registers a typed named converter on this read builder.
func ConvertTo[T, V any](r *ReadBuilder[T], name string, fn func(string) (V, error)) *ReadBuilder[T] {
	if r.converters == nil {
		r.converters = make(convert.Registry)
	}
	convert.ConvertTo(r.converters, name, fn)

	return r
}

## Excelizex is a simple Excel library

English | [中文](README.md)

Module: `github.com/cyclonevox/excelizex`. Requires Go 1.22+ (generics). Built on [excelize](https://github.com/qax-os/excelize).

The hard part of business Excel import/export is rarely reading cells. You already have DTOs with tags, validation, and maybe nested fields; Excel only gives you strings. This library binds those DTOs, converts values, runs your validator, and calls your service.

```bash
go get github.com/cyclonevox/excelizex
```

```go
import (
    "github.com/cyclonevox/excelizex"
    "github.com/cyclonevox/excelizex/layout"
)
```

The library does not parse `validate` tags. Implement `Validator` and plug in playground / validatorx / anything else.

### Layout vs Schema

- **Schema**: logical columns (names, field paths, style names). It does not care which physical row is the header.
- **Layout**: physical shape (notice row, header rows, where data starts).

Built-in layouts:

- `layout.NoticeHeaderData`: notice, header, then data (default)
- `layout.HeaderData`: header then data, no notice row
- `layout.Stacked`: multi-row headers; the last non-empty cell in each column is the bound name
- `layout.Grouped`: group title row + leaf headers (merged on write)

Dynamic extra columns: `map[string]string` with `excel:",extend"`, plus `Extend("语文", "数学")` for column order.

Large exports: `Stream(func(yield func(T) error) error { ... })` (not together with `Rows` / `Template`).

Two-level dropdowns: `Cascade("市", "省", map[string][]string{"浙江": {"杭州", "宁波"}})`. Parent values must be valid Excel defined names (no spaces).

### Write a template or export rows

```go
type Row struct {
    Name  string `excel:"姓名" validate:"required"`
    Age   int    `excel:"年龄"`
    Grade int    `excel:"等级"`
}

wb := excelizex.New()
defer wb.Close()

err := excelizex.Write[Row](wb.Sheet("导入").
    WithLayout(layout.NoticeHeaderData{}).
    WithNotice("请填写信息")).
    Dropdown("等级", []string{"A", "B"}).
    Rows(rows...).
    Apply()

_ = wb.Save(out)
```

Use `Template()` instead of `Rows(...)` to ship an empty form. `Apply()` rebuilds the sheet, so a second apply does not leave leftover data rows.

Nested structs flatten with `excel:",inline"`.

### Read and bind to business code

```go
wb, err := excelizex.Open(f)
if err != nil {
    return err
}
defer wb.Close()

rows, res, err := excelizex.Read[Row](wb.Sheet("导入").
    WithLayout(layout.NoticeHeaderData{})).
    Validate(myValidator).
    Collect(ctx)
```

Or stream into a service with optional concurrency:

```go
_, err = excelizex.Read[Row](wb.Sheet("导入").WithLayout(layout.NoticeHeaderData{})).
    Validate(myValidator).
    Each(ctx, func(ctx excelizex.Context, row Row) error {
        return svc.Create(ctx, row)
    }, excelizex.Concurrency(8))
```

Failed rows: `res.Errors()`, then `wb.WriteErrors(res)` to give users a file they can fix.

### Field conversion

Common types (string / number / bool / time) convert automatically. Business enums belong on the DTO:

- Faster: `ExcelField(header, raw)` / `ExcelExportField(header)`
- Shorter: `ExcelGrade(raw)` / `ExcelExportGrade()`

### Notes

- Sheet names are required.
- `Open(r)` does not close `r`. Close `*os.File` yourself, and `wb.Close()` when done.
- `Each` defaults to concurrency 1. `File()` returns the raw `*excelize.File` **without** the library lock.
- Runnable examples: [`examples/enrollment`](examples/enrollment) (template → dirty import → enroll → fix → export), [`examples/scores`](examples/scores) (grouped headers, extend columns, cascade dropdowns, stream export). Godoc examples: [`example_test.go`](example_test.go).

### Previous API

Docs for the older API: [README_ENG_v1.md](README_ENG_v1.md). Source is at git tag [`v1.5.7`](https://github.com/cyclonevox/excelizex/tree/v1.5.7):

```bash
go get github.com/cyclonevox/excelizex@v1.5.7
```

It is not maintained on this branch and there is no compatibility layer.

The full walkthrough (styles, `EachMap`, conversion priority, protection) is in the [Chinese README](README.md).

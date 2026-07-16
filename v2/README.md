# Excelizex v2

基于 [excelize](https://github.com/qax-os/excelize) 的 Go Excel 导入/导出库。**v2 重写 API**：逻辑列（Schema）与物理版式（Layout）分离，链式调用，泛型 `Read[T]` / `Write[T]`。

## 安装

```bash
go get github.com/cyclonevox/excelizex/v2
```

要求 Go 1.22+（泛型）。

## Schema 与 Layout

- **Schema**：由结构体 `excel` / `conv` / `validate` 等 tag 解析出的**逻辑列**（列名、字段、转换器、校验），不关心表头在第几行。
- **Layout**：工作表**物理版式**——提示行、表头占几行、数据从哪行开始、如何把表头行收成逻辑列名。内置 `layout.NoticeHeaderData`（提示 + 单行表头 + 数据）与 `layout.HeaderData`（无提示行）。

## 读取

```go
type Row struct {
    Name string `excel:"姓名" validate:"required"`
    Age  int    `excel:"年龄"`
}

f, _ := os.Open("import.xlsx")
wb, _ := excelizex.Open(f)
defer wb.Close()

rows, res, err := excelizex.Read[Row](wb.Sheet("导入").WithLayout(layout.NoticeHeaderData{})).
    Validate(validate.Required{}).
    Collect(ctx)
// rows：成功行；res.Errors()：行级错误
```

并发逐行处理用 `Each` / `EachMap`，可选 `Concurrency(n)`。

## 写入

```go
wb := excelizex.New()
err := excelizex.Write[Row](wb.Sheet("导入").
    WithLayout(layout.NoticeHeaderData{}).
    WithNotice("请填写信息")).
    Rows(rows...).
    Apply()
```

### 仅生成模板（无数据行）

```go
excelizex.Write[Row](wb.Sheet("导入").WithLayout(layout.NoticeHeaderData{}).
    WithNotice("请填写信息")).
    Template().
    Apply()
```

### 下拉选项

```go
excelizex.Write[Row](wb.Sheet("导入").WithLayout(layout.NoticeHeaderData{})).
    Dropdown("等级", []string{"A", "B"}).
    Rows(rows...).
    Apply()
```

### 保护工作表

```go
excelizex.Write[Row](wb.Sheet("导入").WithLayout(layout.NoticeHeaderData{})).
    Protect("secret").
    Rows(rows...).
    Apply()
```

保存：`wb.Save(w)` 或 `wb.File()` 访问底层 `*excelize.File`。

## 更多示例

见 [`example_test.go`](example_test.go)（`go test -run Example`）。

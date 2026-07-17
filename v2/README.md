# Excelizex v2

基于 [excelize](https://github.com/qax-os/excelize) 的 Go Excel 导入/导出库。**v2 重写 API**：逻辑列（Schema）与物理版式（Layout）分离，链式调用，泛型 `Read[T]` / `Write[T]`。

## 安装

```bash
go get github.com/cyclonevox/excelizex/v2
```

要求 Go 1.22+（泛型）。

## Schema 与 Layout

- **Schema**：由结构体 `excel` / `conv` / `style` / `time` 等 tag 解析出的**逻辑列**（列名、字段、转换器）。`validate` tag 仅作为**不透明元数据**存入 `schema.Column.Validate`，库本身**不解读、不执行**。
- **Layout**：工作表**物理版式**——提示行、表头占几行、数据从哪行开始、如何把表头行收成逻辑列名。内置 `layout.NoticeHeaderData`（提示 + 单行表头 + 数据）与 `layout.HeaderData`（无提示行）。

## 校验（可插拔）

业务 DTO 通常已有 `validate:"..."` tag，由项目自己的 validator（如 playground/validator、validatorx）解释。excelizex **只负责在绑定后、回调前调用你注入的校验器**：

```go
type Validator interface {
    Validate(row any) error
}
```

`ReadBuilder.Validate` 可链式注册多个 `Validator`。Collect/Each 对每行绑定完成后调用 `v.Validate(&row)`（传指针，便于 `Struct()` 类校验器写回字段）。

```go
// 示例：接入项目 validator
type rowValidator struct{ v *validator.Validate }

func (r rowValidator) Validate(row any) error {
    return r.v.Struct(row)
}

rows, res, err := excelizex.Read[Row](wb.Sheet("导入").WithLayout(layout.NoticeHeaderData{})).
    Validate(rowValidator{v: validatorx.New()}).
    Collect(ctx)
```

DTO 上的 `validate:"required"` 等 tag **属于你的 validator**，不是 excelizex 内置行为。

## 读取

```go
type Row struct {
    Name string `excel:"姓名" validate:"required"` // tag 给业务 validator 用
    Age  int    `excel:"年龄"`
}

f, err := os.Open("import.xlsx")
if err != nil {
    return err
}
defer f.Close() // Open 只读入内容，不会关闭传入的 Reader

wb, err := excelizex.Open(f)
if err != nil {
    return err
}
defer wb.Close()

rows, res, err := excelizex.Read[Row](wb.Sheet("导入").WithLayout(layout.NoticeHeaderData{})).
    Validate(myValidator).
    Collect(ctx)
// rows：成功行；res.Errors()：行级错误
```

并发逐行处理用 `Each` / `EachMap`，可选 `Concurrency(n)`。

## 写入

```go
wb := excelizex.New()
defer wb.Close()
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

保存：`wb.Save(w)` 或 `wb.File()` 访问底层 `*excelize.File`。调用方负责 `wb.Close()`；`Open` 会把 `io.Reader` 全部读入内存且**不会**关闭传入的 Reader（例如 `os.File` 需自行 `Close`）。

## 并发安全

`Workbook` 的库方法（`Save`、`Close`、`RegisterStyle`、`Apply`、`Collect`/`Each` 中的读表、`WriteErrors` 等）通过内部互斥锁串行化对底层 excelize 文件的访问，因此可在同一 `Workbook` 上并发调用这些 API（例如 `Each` 与 `Save` 并行）而不会触发数据竞争。

`Each` 仅在 `GetRows` 等读表阶段持锁；行绑定与回调在内存数据上并行执行，可用 `Concurrency(n)` 控制并发度。

`File()` 是**逃逸口**：返回的 `*excelize.File` 不受库锁保护。若直接调用 excelize API，调用方须自行同步，且不得与并发的库方法混用同一文件句柄。

## 更多示例

见 [`example_test.go`](example_test.go)（`go test -run Example`）。

## E2E 测试

业务向端到端场景在 [`e2e/`](e2e/) 包：每个文件对应一类真实导入/导出流程（批量导入、部分失败回写、乱表头、模板下发、并发导入等）。测试数据**在运行时生成**（`t.TempDir()` / `bytes.Buffer`），通过 `excelizex.Write` 做库 round-trip，或用 `excelize` 构造「用户弄坏的 Excel」。公共 fixture 在 `e2e/fixture/`（`build.go` 造表、`open.go` 打开/保存、`validate.go` 测试用校验桩）。

```bash
go test ./e2e/... -count=1
go test ./e2e/... -count=1 -race
```

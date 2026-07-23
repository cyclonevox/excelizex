# Excelizex v2 示例

可独立运行的 `package main` 示例，演示典型导入/导出链式调用。在对应子目录下执行 `go run .`（默认读取 `../../e2e/testdata/` 夹具）。

| 目录 | 说明 |
|------|------|
| [`import_collect/`](import_collect/) | `Read` → `Collect` + `Validate`；可选 `WriteErrors` 回写 |
| [`import_each/`](import_each/) | `Each` 并发调用业务 `Create`，`Concurrency(n)` |
| [`export_template/`](export_template/) | `Write` → `Template` + `Dropdown` + `Protect` |
| [`export_rows/`](export_rows/) | `Write` → `Rows` 导出已有数据 |

```bash
cd v2

# 编译全部示例
go build ./examples/...

# 导入 Collect（默认空姓名夹具，含 1 条校验失败）
cd examples/import_collect && go run .

# 导入 Each（默认 students_notice_ok.xlsx）
cd ../import_each && go run .

# 导出模板 / 数据
cd ../export_template && go run . -out /tmp/考生模板.xlsx
cd ../export_rows && go run . -out /tmp/考生导出.xlsx
```

Godoc 短示例见仓库根 [`example_test.go`](../example_test.go)；业务向 E2E 见 [`e2e/`](../e2e/)。

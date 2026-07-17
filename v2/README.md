# Excelizex v2

> 根目录是 **v1**（功能冻结）。本目录是 **v2** 模块：`github.com/cyclonevox/excelizex/v2`。

Excelizex 不是又一个「把单元格读写包一层」的库。它要解决的是业务里真正烦人的那一段：

**Excel 永远吐给你 `[][]string`，而你手里已经有一套 Service DTO（还常带校验 tag、转换规则、嵌套字段）。**  
以前为了一张导入表，controller / service 外面经常要写一两百行胶水；这个库存在的意义，就是把那段脏活收成**十几二十行链式调用**。

v1 把这条路走通了（绑定、转换、校验钩子、并发读、错误回写）。v2 在同一目标下做了一次 breaking 重写：API 更干净、类型在编译期可见、版式不再写死，并补齐测试。

要求 **Go 1.22+**（泛型）。底层仍基于 [excelize](https://github.com/qax-os/excelize)，不自己实现 OOXML。

---

## 目录

- [为什么要做 v2](#为什么要做-v2)
- [相对 v1：改了什么、提升了什么](#相对-v1改了什么提升了什么)
- [功能与路线](#功能与路线)
- [安装](#安装)
- [从 v1 迁过来](#从-v1-迁过来)
- [Schema 与 Layout](#schema-与-layout)
- [读取](#读取)
- [写入](#写入)
- [校验（可插拔）](#校验可插拔)
- [字段转换](#字段转换)
- [并发与资源](#并发与资源)
- [示例与测试](#示例与测试)

---

## 为什么要做 v2

v1 在真实业务里证明过价值，但时间一长也暴露出结构性问题：

1. **回调类型是 `any`**：`Run(func(a any) error)` 里到处 `a.(*T)`，编译期帮不上忙。
2. **版式写死在 meta 里**：几乎只能是「提示行 → 表头 → 数据」。要无提示、或以后双表头，只能继续往同一套 meta 里堆特例。
3. **列模型、版式、excelize 细节揉在一起**：`meta*` 一改就牵全身，演进成本高。
4. **错误路径习惯 panic**：库内常规控制流不该靠 panic。
5. **打开已有文件读导入**：曾经依赖写路径缓存一类状态，读路径不够「只认文件内容」。

校验这件事 v1 已经做对了：只留 `Validate` 接口，tag 和扩展都归项目里的 validator。v2 **沿用同一思路**，不在库内再解释 `validate` tag。

v2 的原则很简单：**业务代码仍然短；底层分层清楚；类型在编译期可见。**

---

## 相对 v1：改了什么、提升了什么

| 维度 | v1 | v2 |
|------|----|----|
| 模块 | `github.com/cyclonevox/excelizex` | `github.com/cyclonevox/excelizex/v2`（独立 go.mod） |
| 行类型 | 反射 + `any` 回调 | 泛型 `Read[T]` / `Write[T]`，回调参数就是 `T` |
| 列 vs 版式 | 揉在 meta / tag（`header\|姓名`） | **Schema**（逻辑列）与 **Layout**（物理行号）分离 |
| 默认版式 | 固定 notice-header-data | 仍是默认（`NoticeHeaderData`），但可换成 `HeaderData` 或自定义 Layout |
| Tag | `excel:"header\|姓名"`、`excel-conv` | `excel:"姓名"`、`conv:"grade"`；`validate` 只给业务 validator 用 |
| 嵌套 DTO | 能力有限 | `,inline` flatten；复杂值走命名 `conv` |
| 校验 | 薄接口 + 项目自己的 validator（已做对） | 同一思路：`Validator` + `Validate(&row)`；库仍不解释 tag |
| 错误 | 常 panic | 常规路径返回 `error` |
| 并发读 | ants 池 + `Run` | `Each` / `EachMap` + `Concurrency(n)`；Workbook 内对 excelize 串行化 |
| 失败行 | 支持回写 | `Result` + `WriteErrors` 保留产品能力 |
| 测试 | 偏演示、偏薄 | 子包单测 + `e2e/` 业务场景（运行时造表，不靠提交 xlsx） |

**体感上你最该注意到的提升：**

- 导入回调里不用再断言类型，IDE / 编译器直接认识 `StudentRow`。
- 同一套结构体 tag 既能生成模板，又能读回数据；版式要变时改 Layout，不用重写绑定。
- 读已填好的用户文件时，只依赖 Layout + 文件内容，不再绑写路径的内部缓存。
- 校验继续像 v1：DTO 上已有的 tag，塞进项目里的 validator 即可。

**刻意不做 / 尚未做的**（避免再东一块西一块）：公式、图表、透视、图片；多级下拉；兼容 v1 旧 API；在库内「魔法绑定」任意深度业务聚合根。

---

## 功能与路线

读取：

- [x] 结构体 tag → Schema，按列名绑定（列顺序可变）
- [x] 内置 string / 数字 / bool / time（`time` tag）转换
- [x] 命名 `conv` 转换器
- [x] `,inline` 嵌套展平
- [x] 可插拔 `Validator`（对接 playground / validatorx 等）
- [x] `Collect` 聚合成功行 + `Result` 行级错误
- [x] `Each` / `EachMap` 并发逐行处理业务
- [x] FailFast
- [x] 失败行 `WriteErrors` 回写错误列
- [ ] 更复杂的多行表头 Layout（接口已留，实现按需加）

写入：

- [x] `Write[T]` 导出数据
- [x] `Template()` 只出模板（提示 + 表头）
- [x] 单级下拉、工作表保护
- [x] 样式名组合（header / body / notice 等）
- [ ] 流式大批量写（P2）
- [ ] 动态列 / 原 `extend` 类能力（P2）
- [ ] 多级下拉（P2）

工程：

- [x] 全链路 `error` 返回，无库内常规 panic
- [x] Workbook 级互斥，避免 `Each` 与 `Save`/`Apply` 踩同一 excelize.File
- [x] `example_test.go` + `e2e/` 场景测试
- [ ] 打 `v2.0.0` tag、英文文档与 v1 对齐程度再定

---

## 安装

```bash
go get github.com/cyclonevox/excelizex/v2
```

```go
import (
    excelizex "github.com/cyclonevox/excelizex/v2"
    "github.com/cyclonevox/excelizex/v2/layout"
)
```

---

## 从 v1 迁过来

v2 **不提供** API shim，建议新代码直接用 v2；老项目按表迁移。

| v1 | v2 |
|----|----|
| `excel:"header\|姓名"` | `excel:"姓名"` |
| `excel:"notice"` + 字段存文案 | `WithNotice("...")`；字段上可选 `excel:"notice"` |
| `excel-conv:"grade"` | `conv:"grade"` |
| `Read(ptr, sheet).SetConvert(...).SetValidates(...).Run(fn, n)` | `Read[T](sheet).Convert(...).Validate(...).Each/Collect` |
| `AddSheet` + `NewOptions` | `Write[T](...).Dropdown(...).Template()/Rows(...).Apply()` |
| 回调 `func(any) error` + 断言 | `func(ctx Context, row T) error` |

校验习惯不变：实现一个适配器，内部 `v.Struct(row)`，挂到 `Validate(...)` 上即可。

---

## Schema 与 Layout

- **Schema**：逻辑列——列名、字段路径、转换器名、样式名等。**不管**第几行是表头。`validate` tag 只作为不透明字符串存着，库不执行。
- **Layout**：物理版式——提示在哪一行、表头占几行、数据从哪行开始、如何把表头收成逻辑列名。

内置：

- `layout.NoticeHeaderData`：第 1 行提示，第 2 行表头，第 3 行起数据（默认体验，对齐 v1）
- `layout.HeaderData`：无提示行

以后要「双表头 / 数据从第 N 行」：加新 Layout，不改 Schema、不改绑定核心。

---

## 读取

```go
type Row struct {
    Name string `excel:"姓名" validate:"required"` // tag 给业务 validator
    Age  int    `excel:"年龄"`
    Grade int   `excel:"等级" conv:"grade"`
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

rows, res, err := excelizex.Read[Row](wb.Sheet("导入").
    WithLayout(layout.NoticeHeaderData{}).
    WithNotice("请按模板填写")). // 若需要核对提示文案
    Convert("grade", gradeImport).
    Validate(myValidator).
    Collect(ctx)
// rows：绑定+校验都成功的行
// res.Errors()：行号 + 原因；可用 wb.WriteErrors(res) 回写给用户改
```

逐行进业务（可并发）：

```go
_, err = excelizex.Read[Row](wb.Sheet("导入").WithLayout(layout.NoticeHeaderData{})).
    Convert("grade", gradeImport).
    Validate(myValidator).
    Each(ctx, func(ctx excelizex.Context, row Row) error {
        return svc.Create(ctx, row)
    }, excelizex.Concurrency(8))
```

Excel 行结构 ≠ Service 命令时，用 `EachMap` 做一层很薄的映射即可，不必把 tag 硬贴到大聚合根上。

---

## 写入

```go
wb := excelizex.New()
defer wb.Close()

err := excelizex.Write[Row](wb.Sheet("导入").
    WithLayout(layout.NoticeHeaderData{}).
    WithNotice("请填写信息")).
    Convert("grade", gradeExport).
    Dropdown("等级", []string{"A", "B"}).
    Protect("secret"). // 可选
    Rows(rows...).
    Apply()
```

只发模板（无数据行）：

```go
excelizex.Write[Row](wb.Sheet("导入").
    WithLayout(layout.NoticeHeaderData{}).
    WithNotice("请填写信息")).
    Template().
    Apply()
```

保存：`wb.Save(w)`；需要逃逸时用 `wb.File()` 拿底层 `*excelize.File`（见下方并发说明）。

---

## 校验（可插拔）

和 v1 一样：**库只留钩子，不替你选校验框架。**

```go
type Validator interface {
    Validate(row any) error
}
```

绑定完成后、业务回调前，对每行调用 `Validate(&row)`（传指针，方便 `Struct()` 写回）。  
DTO 上的 `validate:"required,id_card"` 等，由你项目里的 playground / validatorx 解释。

```go
type rowValidator struct{ v *validator.Validate }

func (r rowValidator) Validate(row any) error {
    return r.v.Struct(row)
}

// Read[T](...).Validate(rowValidator{v: validatorx.New()}).Collect(ctx)
```

---

## 字段转换

单元格永远是 string。库处理常见类型；业务枚举、特殊格式用命名转换器：

```go
// 结构体
Grade int `excel:"等级" conv:"grade"`

// 读
Convert("grade", func(raw string) (any, error) { ... })

// 写（导出时 int → 展示文案）
Convert("grade", func(v any) (string, error) { ... })
```

嵌套：

```go
type Address struct {
    City   string `excel:"城市"`
    Street string `excel:"街道"`
}
type Row struct {
    Name string  `excel:"姓名"`
    Addr Address `excel:",inline"`
}
```

不能展平的复杂值：用 `conv` 自己解析，不要指望库硬绑整棵业务树。

---

## 并发与资源

- `Workbook` 的库方法（`Save` / `Close` / `Apply` / 读表 / `WriteErrors` 等）内部互斥，可与 `Each` 并行调用而不数据竞争。
- `Each` 只在取行阶段持锁；绑定与业务回调在内存数据上并行，`Concurrency(n)` 控制并发度。
- `File()` 是逃逸口：返回的 `*excelize.File` **不受**库锁保护，不要和并发中的库 API 混用同一句柄。
- `Open(r)` 会读完 `r` 且**不**关闭 `r`；`*os.File` 请自行 `Close`。记得 `wb.Close()`。

---

## 示例与测试

- 短示例：[`example_test.go`](example_test.go)（`go test -run Example`）
- 业务向 E2E：[`e2e/`](e2e/)——批量导入、部分失败回写、乱表头、模板下发、取消、FailFast 等；表在测试里现场生成，不提交二进制 xlsx

```bash
cd v2
go test ./... -count=1
go test ./... -count=1 -race
```

---

非常欢迎针对真实导入场景提需求或 PR。v2 会优先把「短代码导入闭环」做稳，再往流式写、动态列、多级下拉开。

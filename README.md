## Excelizex 只是一个简单的 excel 库

中文| [English](README_ENG.md)

模块：`github.com/cyclonevox/excelizex`。需要 Go 1.22+（用到了泛型）。底层还是基于 [excelize](https://github.com/qax-os/excelize)。

Excelizex 的目标：用 golang 做 excel 导入导出时，让开发者写得更方便一点。

实际业务里最烦的往往不是读写单元格，而是：Excel 给你的是一堆 string，你手里却已经有一套带 tag、带校验、还可能嵌套的 DTO。以前为了一张导入表，controller / service 外面经常要写很多重复代码，并且无法使用已有的验证。这个库就是想把这部分省略。

现在这套 API 是泛型的 `Read[T]` / `Write[T]`：Schema 和 Layout 分开，DTO 上用 `ExcelField` / `Excel{字段}` 做转换，校验只留可插拔接口。

校验：库里只留接口，tag 和具体校验交给项目里的 validator，不会在库里去解析 `validate` tag。

鉴于时间和精力，非常希望大家能针对真实导入场景提建议，或者直接提 PR。

[目前计划或者已提供的功能](#目前计划或者已提供的功能)

[安装](#安装)

[使用说明](#使用说明)

[开始生成你的 excel](#开始生成你的-excel)

[读取数据并绑业务](#读取数据并绑业务)

[字段转换](#字段转换)

[字段验证器](#字段验证器)

[并发和资源](#并发和资源)

[示例](#示例)

[旧版](#旧版)

---

### 目前计划或者已提供的功能

读取方面：

- [x] 通过结构体 tag 生成 Schema，按列名绑定（列顺序乱了也能对上）
- [x] 内置 string / 数字 / bool / time 转换；DTO 转换：`ExcelField` 接口（直调）或 `Excel{字段}`（图省事，反射）
- [x] `,inline` 把嵌套结构展平到列上
- [x] 可扩展的数据验证，实现 `Validator` 接口即可（playground / validatorx 都行）
- [x] `Collect` 聚合成功行，`Result` 保留行级错误
- [x] `Each` / `EachMap` 绑定业务函数，支持并发
- [x] FailFast
- [x] 失败行可以通过 `WriteErrors` 回写错误列给用户改
- [ ] 更复杂的多行表头 Layout（接口先留着，后面按需加）

写入方面：

- [x] `Write[T]` 导出数据
- [x] `Template()` 只出模板（提示 + 表头）
- [x] 单级下拉、工作表保护
- [x] 样式名组合（header / body / notice 这些）
- [ ] 流式大批量写
- [ ] 动态列 / 原来那种 extend 能力
- [ ] 多级下拉

其他：

- [x] 常规路径都返回 `error`，不再靠 panic 走控制流
- [x] Workbook 里对 excelize 做了互斥，避免 `Each` 和 `Save`/`Apply` 踩同一把
- [x] `example_test.go` + `e2e/` 业务场景测试
- [x] Schema 和 Layout 拆开了：列模型不管物理行号，版式可以换
- [ ] 后面想到再说吧。。。

---

### 安装

```bash
go get github.com/cyclonevox/excelizex
```

```go
import (
    "github.com/cyclonevox/excelizex"
    "github.com/cyclonevox/excelizex/layout"
)
```

---

### 使用说明

一个 excel 文件由多个 sheet 组成。excelizex 也是围着 Sheet 转的。列模型和物理版式拆成两块：

- **Schema**：逻辑上有哪些列——列名、字段路径、样式名。它不管第几行是表头。`validate` tag 只当字符串存着，库自己不执行。
- **Layout**：物理版式——提示在哪一行、表头占几行、数据从哪开始、怎么把表头收成列名。

内置了两种常用的 Layout：

- `layout.NoticeHeaderData`：第 1 行提示，第 2 行表头，第 3 行起数据（默认）
- `layout.HeaderData`：没有提示行

以后要「双表头 / 数据从第 N 行」之类的，可以实现一个 Layout，也不用动绑定核心。

### 需要特别注意的是：

1. Sheet 名称还是必要的，不然不好定位你要操作哪张表。
2. `New()` 创建空白文件时，excelize 默认会带一张 `Sheet1`；写出其他表名时可能会被清掉。`Open()` 打开已有文件时**不会**自动删任何 Sheet。
3. `Open(r)` 只把内容交给 excelize 读，**不会**帮你关 `r`。`*os.File` 请自己 `Close`。Workbook 用完记得 `wb.Close()`。

常用 tag：

| 点 | 写法 |
| --- | --- |
| 表头列 | `excel:"姓名"` |
| 提示行核对 | 多数情况 `WithNotice("...")`；需要读时核对提示文案时再留 `excel:"notice"` |
| 转换 | DTO：`ExcelFieldImporter` / `ExcelFieldExporter`，或 `ExcelGrade` / `ExcelExportGrade` |
| 校验 tag | `validate:"..."`（或你自己 validator 的 tag），库不解析 |
| 忽略字段 | `excel:"-"` |
| 嵌套展平 | `excel:",inline"` |

内置样式名可以组合，例如 `style:"header-red,body"`。自定义样式用 `wb.RegisterStyle(...)`。

| 样式名 | 用途 |
| --- | --- |
| `notice` | 提示行 |
| `header` / `header-red` | 表头 |
| `body` / `body-blue` | 数据列 |
| `locked` / `unlocked` | 保护相关 |

---

### 开始生成你的 excel

通常情况下，无论你是要生成一个待填写的模板，还是一份已经装好数据的表，先定一个类型还是第一步。

大多情况下用 `excel` tag 就够了，内容就是表头名。特殊格式在 DTO 上写转换方法（见 [字段转换](#字段转换)）；校验 tag 留给你自己的 validator。

#### 例如现在有个 Row 类型，并且已经加好 tag

```go
type Row struct {
    Name  string `excel:"姓名" validate:"required"`
    Age   int    `excel:"年龄"`
    Grade int    `excel:"等级"`
}

// 图省事：按字段写 ExcelGrade（反射调用）
func (r *Row) ExcelGrade(raw string) error {
    g, err := grade.Parse(raw)
    if err != nil {
        return err
    }
    r.Grade = int(g)
    return nil
}

func (r *Row) ExcelExportGrade() (string, error) {
    return grade.Format(r.Grade)
}

// 要性能也可以不写上面两个，改实现 ExcelField / ExcelExportField，内部 switch 表头
```

然后 `New()` 出一个 workbook，链式写出去：

```go
wb := excelizex.New()
defer wb.Close()

err := excelizex.Write[Row](wb.Sheet("导入").
    WithLayout(layout.NoticeHeaderData{}).
    WithNotice("请填写信息")).
    Dropdown("等级", []string{"A", "B"}).
    Protect("secret"). // 可选
    Rows(rows...).
    Apply()
```

只发模板、不要数据行的话：

```go
excelizex.Write[Row](wb.Sheet("导入").
    WithLayout(layout.NoticeHeaderData{}).
    WithNotice("请填写信息")).
    Template().
    Apply()
```

`Apply()` 是整表重建：会先清掉目标 Sheet 上旧的数据行、数据验证，以及本库写选项时生成的辅助表，再写出这次的 notice / 表头 / 数据。所以同一张表第二次 `Apply`（包括再出一遍 `Template()`）不会残留上次的数据行。

保存用 `wb.Save(w)`；实在要逃逸可以 `wb.File()` 拿底层 `*excelize.File`（但注意下面并发那节）。工作表保护走 `Protect(password)`。

嵌套字段可以 `,inline` 展平：

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

展不平的复杂值，在 DTO 上写 `Excel*` / `ExcelField`，或者在 `Each` 业务里自己处理。别指望库去硬绑一整棵业务树。

---

## 读取数据并绑业务

依旧以 Row 为例。你可以用实现了 `io.Reader` 的东西喂给 `Open`。

```go
f, err := os.Open("import.xlsx")
if err != nil {
    return err
}
defer f.Close()

wb, err := excelizex.Open(f)
if err != nil {
    return err
}
defer wb.Close()

rows, res, err := excelizex.Read[Row](wb.Sheet("导入").
    WithLayout(layout.NoticeHeaderData{}).
    WithNotice("请按模板填写")). // 需要核对提示文案时再用
    Validate(myValidator).
    Collect(ctx)
// rows：绑定+校验都过了的行
// res.Errors()：行号 + 原因；可以用 wb.WriteErrors(res) 回写给用户改
```

如果更想逐行进业务（可以并发）：

```go
_, err = excelizex.Read[Row](wb.Sheet("导入").WithLayout(layout.NoticeHeaderData{})).
    Validate(myValidator).
    Each(ctx, func(ctx excelizex.Context, row Row) error {
        return svc.Create(ctx, row)
    }, excelizex.Concurrency(8))
```

Excel 行结构和 Service 命令对不上时，用 `EachMap` 做一层很薄的映射就行，没必要硬把 tag 贴到大聚合根上。

---

### 字段转换

单元格读上来永远是 string。常见类型（string / 数字 / bool / time）库会帮你转。业务枚举、特殊格式挂在 **DTO 方法**上。

库导出了两个兜底接口（实现对应方法即可，不必显式写 `var _ ExcelFieldImporter = ...`）：

```go
type ExcelFieldImporter interface {
    ExcelField(header, raw string) (handled bool, err error)
}

type ExcelFieldExporter interface {
    ExcelExportField(header string) (val string, handled bool, err error)
}
```

#### 写法一：要性能 / 列比较多 —— 实现上面两个接口，自己 switch

```go
func (r *Row) ExcelField(header, raw string) (bool, error) {
    switch header {
    case "等级":
        g, err := grade.Parse(raw) // 枚举逻辑仍在枚举包，不必改领域类型
        if err != nil {
            return true, err
        }
        r.Grade = int(g)
        return true, nil
    default:
        return false, nil // 这列交给库内置转换
    }
}

func (r *Row) ExcelExportField(header string) (string, bool, error) {
    if header != "等级" {
        return "", false, nil
    }
    s, err := grade.Format(r.Grade)
    return s, true, err
}
```

固定方法名，库断言成接口后**直接调用**（不走反射 `Call`）。

#### 写法二：图省事 —— 按字段名写 `Excel{字段}` / `ExcelExport{字段}`

```go
func (r *Row) ExcelGrade(raw string) error { ... }         // 读：字段名 Grade → ExcelGrade
func (r *Row) ExcelExportGrade() (string, error) { ... } // 写（Go 不能重载，所以导出方法加 Export）
```

写起来短，但方法名是动态拼的，只能反射调用，比写法一慢一点。少量列、不在意这点开销时够用。

优先级（读）：`Excel{字段}` → `ExcelField` → `TextUnmarshaler` → 内置。  
优先级（写）：`ExcelExport{字段}` → `ExcelExportField` → `TextMarshaler` → 内置。

想看两者实际差距可以跑（5 个需转换列的对照）：

```bash
go test ./bind/ -bench='BenchmarkBindRow_|BenchmarkExportRow_' -benchmem
```

一次性、只在业务里用的逻辑，也可以继续写在 `Each` / `Collect` 之后，不必挂 DTO。

---

### 字段验证器

处理 http 业务时，验证大多放在中间件或者 controller 那一层。excelizex 这边也一样：只要实现 Validate 接口，就可以在业务函数执行前挂上去。

**库只留钩子，不替你选校验框架。**

```go
type Validator interface {
    Validate(row any) error
}
```

绑定完成后、业务回调前，会对每行调用 `Validate(&row)`（传指针，方便 `Struct()` 写回）。  
DTO 上的 `validate:"required,id_card"` 之类，仍然由你项目里的 playground / validatorx 解释。

例如：

```go
type rowValidator struct{ v *validator.Validate }

func (r rowValidator) Validate(row any) error {
    return r.v.Struct(row)
}

// Read[T](...).Validate(rowValidator{v: validatorx.New()}).Collect(ctx)
```

---

### 并发和资源

- Workbook 上的库方法（`Save` / `Close` / `Apply` / 读表 / `WriteErrors` 等）内部有互斥，可以和 `Each` 一起用，不至于数据竞争。
- `Each` 默认并发是 **1**（串行）。`Concurrency(n)` / `SetConcurrency(n)` 控制 worker 上限。
- `File()` 是逃逸口：返回的 `*excelize.File` **不受**库锁保护，不要跟正在并发跑的库 API 混用同一句柄。
- 大文件可以看 `WithUnzipSizeLimit` / `WithUnzipXMLSizeLimit`。

依赖上锁定 Go 1.22+。底层 excelize 先停在 `v2.9.0`：上游 `v2.9.1+` / `v2.10+` 会抬高 Go 版本要求，go 版本升级后面会逐步考虑。

---

### 示例

[`examples/enrollment`](examples/enrollment) 是一条完整的考务导入：下发模板 → 班主任交表（含填错）→ 校验入库 → 错误列退回 → 改完再导 → 导出名单。

```bash
go run ./examples/enrollment
```

Godoc 短示例见 [`example_test.go`](example_test.go)。

---

### 旧版

旧版说明见 [README_v1.md](README_v1.md)。代码在 git tag [`v1.5.7`](https://github.com/cyclonevox/excelizex/tree/v1.5.7)：

```bash
go get github.com/cyclonevox/excelizex@v1.5.7
```

本仓库主线不再维护那套接口，也不提供兼容层。

欢迎针对真实导入场景提需求或 PR。

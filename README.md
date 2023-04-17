## Excelizex 只是一个简单的excel库

## This project has not done

中文| [English](README_ENG.md)

Excelizex目标就是在使用golang导入excel时，使开发者调用更加方便,更简单。

目前该库时是基于qax-os/excelize的封装。
由此提供了提供了一些数据绑定，验证等功能，来更方便的支持业务编写。
由于实际开发过程中本人接触的导入操作更多，目前该库也支持并发读取，重心暂时放在在读取方面。
之后会完善更多功能，当然也包含代码优化，API简单化，以及尽可能的性能优化。

鉴于时间和精力以及个人能力，目前只提供了 [提示信息行 - 表头行 - 数据行] excel模板，
非常希望大家能够提出改进建议或者是直接提交代码。

[现状及目前计划](#目前计划或者已提供的功能有)

[说明](#使用说明)

[开始生成你的excel](#开始生成你的excel)

[为Header和notice增加样式选项](#为Header和notice增加样式选项)

[添加某一列的下拉选项](#添加某一列的下拉选项)

[Excel数据读取业务支持](#Excel数据读取业务支持)

[字段转换器](#字段转换器)

[字段验证器](#字段验证器)


****

### 目前计划或者已提供的功能有

读取方面：

- [x] 基本的数据绑定，可以通过结构体的tag来提供生成excel表格，或读取excel表格数据
- [x] 读取excel表格时可扩展的数据验证支持，仅需要实现excelizex.Validate接口。
- [x] 在读取excel数据时，通过设置转换器，来完成业务的需求下对excel中数据的转换。是方便有重复的转换场景，便于减少代码。
- [x] 在数据绑定的结构体下 支持并发读取表行的功能，绑定业务函数进行业务编写。
- [x] 实现单级下拉选项菜单
- [ ] 提供更便利的多级下拉菜单的生成方法

写入方面：

- [x] 通过流式写入功能支持对读取表的验证和业务操作后产生的结果生成excel文件
- [x] 简单的数据写入依旧可以通过数据绑定的形式来生成数据。当然意味着略多的消耗。
- [ ] 通过实现excelizex提供的迭代器接口来配合 qax-os/excelize的流式写入方法 对Excel文件批量写入


- [x] excel加密限制，目前使用qax-os/excelize提供的excel内置加密保护选项。
- [ ] 更多的单元测试保证代码实现的正确性
- [ ] 后面的想到再说吧。。。。233

****

// 使用说明正在施工中...

### 使用说明：

众所周知，一个excel文件由多个sheet来组成，所以符合直觉的是，不管是qax-os/excelize,还是excelizex也都是这也做的。

不同的是excelizex基于反射去获取对象的tag来生成sheet，仅需要简单的tag就可以完成sheet的生成等工作，目前tag也仅包含了 excel,
style ,validate
无论是设置表头、提示的样式还是数据。当然一旦这样做，sheet也就仅内置好了三种类型，notice-抬头提示 header-表头 以及data-数据行

### 需要特别注意的是：

由于目前excelizex是基于qax-os/excelize做的，so

1. Sheet的名称是必要的。否则excelizex不方便找到你所需要操作的表是什么
2. Sheet的名称不能使用Sheet1名称，因为 qax-os/excelize 默认最后会删除名称为Sheet1的表

****

### 开始生成你的excel：

通常情况下，无论你是要生成一个待填写的模板，还是说一份已经装载好数据的表单，设置一个类型是实现业务的第一步。
使用大多情况下，你仅仅需要使用`excel` tag即可，tag填写`type|content`，
其中type是该字段生命该列是作为notice还是还是header，而content仅在header下生效，以为其对应的表头content即是指该列表头的具体内容。

#### 例如现在有testStruct类型，并且为其已经加好tag

```go
type Test struct {
Notice  string `excel:"notice"`
IdCard  string `excel:"header|身份证号" `
Name    string `excel:"header|姓名" `
Grade   string `excel:"header|年级" `
Class   string `excel:"header|班级" `
// 当然如果你不需要notice作为sheet中的提醒，当你结构体中没有notice tag时则不会生成notice行。
// 以本例来说 notice 会直接将Test.Notice中值作为notice内容。
}
```

紧接着就可以调用excelize.New()创建excel文件，再链式的调用AddSheet()方法

```go
ee := new(dto.Test)
ee.Notice = "*表格中标红字段为必填项，请按要求进行填写，否则可能会导致数据导入失败"
es := excelizex.New().AddSheet("导入导入", ee)

// 紧接着，你可以使用bufffer和SaveAs来帮助你生成文件(当然该方法直接支持excel加密)
es.buffer()
es.SaveAs("文件名称")
```

当然如果你想生成或者导出数据，你可以直接将一个类型的切片传入到AddSheet方法中，excelizex会生成一张表
还是以本例中的Test struct为例:

```go
ees:= make([]*dto.Test)
ees = append(ees, &Test{
Notice: "*表格中标红字段为必填项，请按要求进行填写，否则可能会导致数据导入失败",
IdCard: 123123123123,
Name:"测试人员",
Grade:"一年级",
Class:"二班"
})
es := excelizex.New().AddSheet("导入导入", ees)
```

### 为Header和notice增加样式选项

首先excelizex提供了几种内置的style：

1. default-notice
2. default-header-red
3. default-header
4. default-all
5. red-font
6. alignment
7. numFmtText
8. default-locked
9. default-no-locked

```go
type Test struct {
Notice  string `excel:"notice" style:"default-notice"`
IdCard  string `excel:"header|身份证号" style:"default-header-red"`
Name    string `excel:"header|姓名" style:"default-header-red" `
Grade   string `excel:"header|年级" style:"default-header-red" `
Class   string `excel:"header|班级" style:"default-header-red" `
}
```

当然你完全可以在tag层面相互组合。例如`style:"red-font,alignment"` execlizex检测到后会自动组合两种style，
但值得注意的是，excelizex的style还是基于excelize.Style来做的，所以相同字段会被覆盖，所以无论是从这点来看excelizex还实现了style的添加功能
你可以实现excelizex.Style接口，你也可以直接使用excelizex.style.DefaultStyle

```go
excelizex.New().AddStyles("custom-style", NewDefaultStyle(&excelize.Style{NumFmt: 49}))
```

### 添加某一列的下拉选项

```go
excelizex.New().AddSheet("考生导入", ee,
NewOptions("学生姓名", []string{"tom", "jerry"}),
NewOptions("学生号码", []string{"13380039232", "13823021932", "17889032312"}),
NewOptions("学生编号", []string{"1", "2", "3"}),
)
```

excelizex.NewOptions 中第一个入参请填写那一列的header名称，而后面的数组即是选项。

****

## Excel数据读取业务支持

依旧以Test类型为例

```go
type Test struct {
Notice  string `excel:"notice"`
IdCard  string `excel:"header|身份证号" `
Name    string `excel:"header|姓名" `
Grade   string `excel:"header|年级" `
Class   string `excel:"header|班级" `
```

你可以使用实现了io.Reader接口的任何对象来作为excelizex.New的入参。excelizex内置了一些从http
body获取或者multipart提取excel文件的方法。

```go
goNum:= 1
f, err := os.Open("./xlsx/batch_data.xlsx")
if err != nil {
panic(err)
}
r := excelizex.New(f).Read(new(Test), "Sheet1").Run(func (a any) (err error) {
b := a.(*Test)
time.Sleep(1 * time.Second)
fmt.Println(b)

return
}, goNum)

```

该示例中，goNum设置为1以上时，则为并发读取操作。Run 中传入的func(a any)(err error)为具体的业务函数。

excel每行读取后，则会执行匿名函数的内容。

### 字段转换器

本类型中，需要注意的是，我们不仅使用`excel`tag 还使用了`excel-conv`的tag，

```go
type Test2 struct {
Id   int64  `excel:"header|埃低"`
Name string `excel:"header|名称"`
List []struct {
Id int64
} `excel:"header|列表" excel-conv:"list"`
}
```

往往，在业务中，你的很多操作并非是类型完全一致的，并且可能你还可能会对excel提取上来的值做某些的处理。
例如枚举值等等，用户可能填的是 是、否，而你实际业务中需要的是 true false。

该方法有点类似于中间件一样，在excel中表格的值映射至每个对象中时，会先执行此函数，以避免excelize读取出的string类型与你的业务中使用的结构体完全不同
。

当然这样的简单的东西你也完全可以写道业务代码中去处理。但最后我还是保留了这个方法以减少重复的业务代码。
当你使用与业务同样的类型来做excel操作时，也许他会有点作用。如何使用转换器、是否使用转换器都取决于你自己。

在该例子中他代表会使用名称为list的转换器将该表头下的数据进行转换。

```go
func listConvert(rawData string) (any, error) {
i, err := strconv.ParseInt(rawData, 10, 64)
if err != nil {
return nil, err
}

return []struct{ Id int64 }{{i}}, nil
}

var (
sList []readTestStruct
s = new(readTestStruct)
)

file.SetConvert("list", listConvert)

```

### 字段验证器

在处理http web业务时，验证操作大多都放在了中间件或者说controller等层来做数据绑定并验证。

那么在exceliezx中，数据表的验证实际上可以使用和go-playground/validator/v10同样的数据验证。只要实现Validate接口的的方法都可以被
添加到excelizex的添加方法中。只要在业务执行函数之前调用SetValidates传入验证器的实现即可使用验证器。
当然不要忘记在具体使用的类型上加上tag，具体的tag当然取决于你的验证器具体实现。

例如我使用在基于go-playground/validator/v10封装的cyclonevox/validatorx 则使用的tag为`validation`

```go

type Test struct {
Notice  string `excel:"notice" style:"default-notice"`
IdCard  string `excel:"header|身份证号" style:"default-header-red"  validate:"required,id_card"`
Name    string `excel:"header|姓名" style:"default-header-red"  validate:"required"`
Grade   string `excel:"header|年级" style:"default-header-red" `
Class   string `excel:"header|班级" style:"default-header-red" `
}

```

```go
if results, err = excel.Read(new(Test), "考生导入").
SetValidates(e.validator).Run(func (any any) (err error) {
ee := any.(*Test)
...
}
```

// 未完待续


## Excelizex 只是一个简单的excel库

中文| [English](README_ENG.md)

Excelizex目标就是在使用golang导入excel时，使开发者调用更加方便,更简单。

目前该库时对qax-os/excelize 的简单封装。
本库提供了提供了一些数据绑定，验证等功能，来更方便的支持业务编写。
由于实际开发过程中本人接触的导入操作更多，目前该库也支持并发读取，重心暂时放在在读取方面。
之后会完善更多功能。

鉴于时间和精力以及个人能力，目前只提供了 [提示信息行 - 表头行 - 数据行] excel模板，
以及有更好的建议等等，非常感谢提出建议。

****
目前计划或者已提供的功能有

- [x] 基本的数据绑定
- [x] 通过实现excelizex提供的迭代器接口来配合 qax-os/excelize的流式写入方法 对Excel文件批量写入
- [x] 通过数据绑定的结构体声明变量 来生成包含数据的Sheet(表)
- [x] 在读取时，通过设置转换器，来完成业务的需求下对excel中数据的转换
- [x] 在数据绑定的结构体下 支持并发读取表行的功能，绑定业务函数进行业务编写。
- [x] 通过流式写入功能支持对读取表的验证和业务操作后产生的结果生成excel文件
- [x] 实现单级下拉选项菜单
- [x] excel加密限制，目前使用qax-os/excelize
- [x] 可扩展的数据验证支持，仅需要实现excelizex.Validate接口，并在excelizex.Read()进行测试
- [ ] 提供更便利的多级下拉菜单的生成方法
- [ ] 后面的想到再说吧。。。。233
****

// 使用说明正在施工中...

#### 使用说明：

excelizex抽象了一个Sheet类型，而一个excel是由多个Sheet组成的，则可以将
Sheet类型作为生成excel的参数，Sheet类型中包含表名表头等

```go
type Sheet struct {
// 表名
Name string `json:"name"`
// 顶栏提示
Notice string `json:"notice"`
// 表头
Header []string `json:"header"`
// 数据
Data [][]any `json:"data"`
....
}
```

可以通过 **excelizex.NewSheet()** 的方法来创建Sheet(表)，并经由
**excelizex.New()** 创建excel文件类型，并调用其AddSheets方法来加入创建好的表，
而excelizex 用到的qax-os/excelize库创建excel 文件时则会默认生成一个Sheet1表，
excelizex目前采取的方案是到最后生成excel的os.File或者 bytes时会删除掉该表。
(其实就是懒得判断懒得给默认生成的Sheet1改名

### 所以当然至少目前需要特别注意的是：

1. Sheet的名称是必要的。否则excelizex不方便找到你所需要操作的表是什么
2. Sheet的名称不能使用Sheet1名称，因为最后会删除名称为Sheet1的表
****

### 写入：

使用`excel` tag即可，tag中的内容则为表的表头。

##### 例如现在有testStruct类型，并且为其已经加好tag

```go
type testStruct struct {
Name       string `excel:"名称" json:"sheet"`
Sex        string `excel:"性别" json:"sex"`
HelloWorld string `excel:"测试" json:"helloWorld"`
}

```

```go
// 你可以使用 SetHeaderByStruct方法来生成
// 例如：
s:= excelizex.NewSheet(excelizex.SetHeaderByStruct(&testStcut{})).SetName("test")
excelFile :=excelizex.New().AddSheet(s)

// 或者
excelFile := excelizex.New().AddSimpleSheet(&testStcut{}, excelizex.SetName("test"))

// 如果是已经有一个创建好的slice，并且也想将他的数据写入表中
ttt := []testStruct{
{"123", "男", "456"},
{"456", "女", "213"},
}

s := NewDataSheet(&ttt, excelizex.SetName("test"))
excelFile := excelizex.New().AddSheet(s)
// 或者

excelFile:= excelizex.New().AddDataSheet(&testStcut{}, excelizex.SetName("test"))
```
excelizex 同样支持了流式迭代器写入的方式，StreamWriteIn方法使用了qax-os/excelize中的流式迭代器，通过实现StreamWritable接口来创建表并写入。

****
### excelizex 也支持了对excel读取并绑定业务函数的功能

```go
// 首先通过 excelizex.New() 来读取文件,以http的multipart传输的文件为例。
var fileHeader *multipart.FileHeader
if fileHeader, err = ctx.FormFile("file"); err != nil {
return
}

filename := fileHeader.Filename
isXlsx := strings.HasSuffix(filename, ".xlsx")
if !isXlsx {
errors.New("support xlsx excel type only")
}

var file multipart.File
if file, err = fileHeader.Open(); err != nil {
return
}
defer file.Close()

excel:=excelizex.New(file)
```
以readTestStruct为例

```go
type readTestStruct struct {
	Id   int64  `excel:"埃低"`
	Name string `excel:"名称"`
	List []struct {
		Id int64
	} `excel:"列表" excel-conv:"list"`
}
```
本类型中，需要注意的是，我们不仅使用`excel`tag 还使用了`excel-conv`的tag，
在该例子中他代表会使用名称为list的转换器将该表头下的数据进行转换。
**并且传入READ()中的绑定结构体变量 需要使用其变量的指针而不是值**
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

file.SetConvert("list", listConvert).Read("test", s, func() error {
sList = append(sList, *s)

return nil
})

```



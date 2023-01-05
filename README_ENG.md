## excelizex is a simple Excel library.

excelizex is a simple package of qax-os/excelize. It's purpose is to provide a certain degree of ease of use when importing and exporting Excel files to meet business development needs (of course, it still a piece of shit. It is only because excelize does not directly provide data binding and other functions to facilitate their own use. In view of time, energy and personal capabilities, only the [Notice Information Line - Header Line - Data Line] excel template is currently provided.
****
The functions currently planned or provided are:

- [x] Basic data binding
- [x] 通过实现excelizex提供的迭代器接口来配合 qax-os/excelize的流式写入方法 对Excel文件批量写入
- [x] 通过数据绑定的结构体声明变量 来生成包含数据的Sheet(表)
- [x] Set the converter to complete the data conversion in Excel according to the business requirements.
- [x] The data binding structure supports the function of reading table rows and binding business functions.
- [x] 通过流式写入功能支持对读取表的验证和业务操作后产生的结果生成excel文件
- [ ] Provide built-in data validation, support its extension, and support translation.
- [ ] Provides a more convenient method for generating multi-level pull-down menus.
- [ ] more..

****

#### Instructions:

Excelizex abstracts a Sheet type, and an excel is composed of multiple sheets. The Sheet type can be used as a parameter to generate excel. The Sheet type includes table name, header, etc

```go
type Sheet struct {
// Sheet Name
Name string `json:"name"`
// Notice Information Line
Notice string `json:"notice"`
// Header Line
Header []string `json:"header"`
// Data Lines
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

1.Sheet的名称是必要的。否则excelizex不方便找到你所需要操作的表是什么
2.Sheet的名称不能使用Sheet1名称，因为最后会删除名称为Sheet1的名称
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
```go
func listConvert(rawData string) (any, error) {
	i, err := strconv.ParseInt(rawData, 10, 64)
	if err != nil {
		return nil, err
	}

	return []struct{ Id int64 }{{i}}, nil
}

var sList []readTestStruct

file.SetConvert("list", listConvert).Read("test", s, func() error {
sList = append(sList, *s)

return nil
})

```



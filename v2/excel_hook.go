package excelizex

import "github.com/cyclonevox/excelizex/v2/bind"

// ExcelFieldImporter 是读侧兜底转换接口（接口断言后直接调用，不走反射 Call）。
// 在 DTO 上实现 ExcelField；热路径、列多时建议用这个，内部自己 switch 表头。
type ExcelFieldImporter = bind.ExcelFieldImporter

// ExcelFieldExporter 是写侧兜底转换接口（同样直调）。
// 在 DTO 上实现 ExcelExportField。
type ExcelFieldExporter = bind.ExcelFieldExporter

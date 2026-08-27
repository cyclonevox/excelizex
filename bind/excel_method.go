package bind

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
)

// ExcelFieldImporter is the catch-all import hook.
// Implementing it on a DTO enables a direct method call (no reflect.Call).
// Prefer this on hot paths; switch on header inside ExcelField.
type ExcelFieldImporter interface {
	ExcelField(header, raw string) (handled bool, err error)
}

// ExcelFieldExporter is the catch-all export hook (direct call).
type ExcelFieldExporter interface {
	ExcelExportField(header string) (val string, handled bool, err error)
}

// Convenience per-field hooks (ExcelGrade / ExcelExportGrade, …) are still
// discovered by name and invoked via reflect — easier to write, slower.

type excelImportPlan struct {
	fieldIdx   map[string]int // leaf field name -> method index on *T (reflect path)
	hasCatchAll bool          // type has ExcelField with the right signature
}

type excelExportPlan struct {
	fieldIdx    map[string]int
	hasCatchAll bool
}

var (
	importPlanCache sync.Map // reflect.Type (struct) -> *excelImportPlan
	exportPlanCache sync.Map // reflect.Type (struct) -> *excelExportPlan
	errorType       = reflect.TypeOf((*error)(nil)).Elem()
)

func leafFieldName(path string) string {
	parts := strings.Split(path, ".")
	for i := len(parts) - 1; i >= 0; i-- {
		if parts[i] != "" {
			return parts[i]
		}
	}

	return ""
}

func structTypeOf(v reflect.Value) reflect.Type {
	t := v.Type()
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return nil
	}

	return t
}

func rowPointer(row reflect.Value) reflect.Value {
	if row.Kind() == reflect.Ptr {
		if row.IsNil() {
			return reflect.Value{}
		}

		return row
	}
	if row.CanAddr() {
		return row.Addr()
	}
	cp := reflect.New(row.Type())
	cp.Elem().Set(row)

	return cp
}

func importPlanFor(structType reflect.Type) *excelImportPlan {
	if structType == nil {
		return &excelImportPlan{}
	}
	if v, ok := importPlanCache.Load(structType); ok {
		return v.(*excelImportPlan)
	}
	plan := buildImportPlan(structType)
	actual, _ := importPlanCache.LoadOrStore(structType, plan)

	return actual.(*excelImportPlan)
}

func exportPlanFor(structType reflect.Type) *excelExportPlan {
	if structType == nil {
		return &excelExportPlan{}
	}
	if v, ok := exportPlanCache.Load(structType); ok {
		return v.(*excelExportPlan)
	}
	plan := buildExportPlan(structType)
	actual, _ := exportPlanCache.LoadOrStore(structType, plan)

	return actual.(*excelExportPlan)
}

func buildImportPlan(structType reflect.Type) *excelImportPlan {
	plan := &excelImportPlan{fieldIdx: make(map[string]int)}
	ptrType := reflect.PointerTo(structType)
	if ptrType.Implements(reflect.TypeOf((*ExcelFieldImporter)(nil)).Elem()) {
		plan.hasCatchAll = true
	}
	for i := 0; i < ptrType.NumMethod(); i++ {
		m := ptrType.Method(i)
		switch {
		case m.Name == "ExcelField", strings.HasPrefix(m.Name, "ExcelExport"):
			continue
		case strings.HasPrefix(m.Name, "Excel"):
			field := strings.TrimPrefix(m.Name, "Excel")
			if field != "" && isImportFieldSig(m.Type) {
				plan.fieldIdx[field] = m.Index
			}
		}
	}

	return plan
}

func buildExportPlan(structType reflect.Type) *excelExportPlan {
	plan := &excelExportPlan{fieldIdx: make(map[string]int)}
	ptrType := reflect.PointerTo(structType)
	if ptrType.Implements(reflect.TypeOf((*ExcelFieldExporter)(nil)).Elem()) {
		plan.hasCatchAll = true
	}
	for i := 0; i < ptrType.NumMethod(); i++ {
		m := ptrType.Method(i)
		switch {
		case m.Name == "ExcelExportField":
			continue
		case strings.HasPrefix(m.Name, "ExcelExport"):
			field := strings.TrimPrefix(m.Name, "ExcelExport")
			if field != "" && isExportFieldSig(m.Type) {
				plan.fieldIdx[field] = m.Index
			}
		}
	}

	return plan
}

// Method types from Type.Method include receiver as In(0).
func isImportFieldSig(t reflect.Type) bool {
	return t.NumIn() == 2 && t.NumOut() == 1 &&
		t.In(1).Kind() == reflect.String &&
		t.Out(0) == errorType
}

func isExportFieldSig(t reflect.Type) bool {
	return t.NumIn() == 1 && t.NumOut() == 2 &&
		t.Out(0).Kind() == reflect.String && t.Out(1) == errorType
}

func tryExcelImport(row reflect.Value, fieldPath, header, raw string, catchAll ExcelFieldImporter) (handled bool, err error) {
	ptr := rowPointer(row)
	if !ptr.IsValid() || ptr.IsNil() {
		return false, nil
	}
	plan := importPlanFor(structTypeOf(row))
	if name := leafFieldName(fieldPath); name != "" {
		if idx, ok := plan.fieldIdx[name]; ok {
			return true, invokeImportField(ptr.Method(idx), raw)
		}
	}
	if catchAll != nil {
		return catchAll.ExcelField(header, raw)
	}

	return false, nil
}

func invokeImportField(m reflect.Value, raw string) error {
	out := m.Call([]reflect.Value{reflect.ValueOf(raw)})
	if e := out[0].Interface(); e != nil {
		return e.(error)
	}

	return nil
}

func tryExcelExport(row reflect.Value, fieldPath, header string, catchAll ExcelFieldExporter) (val string, handled bool, err error) {
	ptr := rowPointer(row)
	if !ptr.IsValid() || ptr.IsNil() {
		return "", false, nil
	}
	plan := exportPlanFor(structTypeOf(row))
	if name := leafFieldName(fieldPath); name != "" {
		if idx, ok := plan.fieldIdx[name]; ok {
			s, err := invokeExportField(ptr.Method(idx))

			return s, true, err
		}
	}
	if catchAll != nil {
		return catchAll.ExcelExportField(header)
	}

	return "", false, nil
}

func invokeExportField(m reflect.Value) (string, error) {
	out := m.Call(nil)
	if e := out[1].Interface(); e != nil {
		return out[0].String(), e.(error)
	}

	return out[0].String(), nil
}

func fieldAssignError(header string, err error) error {
	return fmt.Errorf("%s: %w", header, err)
}

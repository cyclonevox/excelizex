package bind

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
)

// pathStep is one segment of a schema FieldPath on a concrete struct type.
// index is StructField.Index from FieldByName at that level (may be multi for embedding).
type pathStep struct {
	index []int
}

type fieldPathPlan struct {
	steps []pathStep
}

type fieldPathCacheKey struct {
	typ  reflect.Type
	path string
}

var fieldPathCache sync.Map // fieldPathCacheKey -> *fieldPathPlan

func fieldByPath(v reflect.Value, path string) (reflect.Value, error) {
	root := structTypeOf(v)
	if root == nil {
		return reflect.Value{}, fmt.Errorf("bind: invalid path %q", path)
	}
	plan, err := fieldPathPlanFor(root, path)
	if err != nil {
		return reflect.Value{}, err
	}

	return fieldByPlan(v, plan)
}

func fieldPathPlanFor(root reflect.Type, path string) (*fieldPathPlan, error) {
	key := fieldPathCacheKey{typ: root, path: path}
	if v, ok := fieldPathCache.Load(key); ok {
		return v.(*fieldPathPlan), nil
	}
	plan, err := buildFieldPathPlan(root, path)
	if err != nil {
		return nil, err
	}
	actual, _ := fieldPathCache.LoadOrStore(key, plan)

	return actual.(*fieldPathPlan), nil
}

func buildFieldPathPlan(root reflect.Type, path string) (*fieldPathPlan, error) {
	parts := strings.Split(path, ".")
	t := root
	steps := make([]pathStep, 0, len(parts))
	for _, p := range parts {
		if p == "" {
			continue
		}
		for t.Kind() == reflect.Ptr {
			t = t.Elem()
		}
		if t.Kind() != reflect.Struct {
			return nil, fmt.Errorf("bind: invalid path %q", path)
		}
		f, ok := t.FieldByName(p)
		if !ok {
			return nil, fmt.Errorf("bind: unknown field %q in path %q", p, path)
		}
		idx := append([]int(nil), f.Index...)
		steps = append(steps, pathStep{index: idx})
		t = f.Type
	}
	if len(steps) == 0 {
		return nil, fmt.Errorf("bind: invalid path %q", path)
	}

	return &fieldPathPlan{steps: steps}, nil
}

func fieldByPlan(v reflect.Value, plan *fieldPathPlan) (reflect.Value, error) {
	cur := v
	for _, st := range plan.steps {
		for cur.Kind() == reflect.Ptr {
			if cur.IsNil() {
				if !cur.CanSet() {
					return reflect.Value{}, fmt.Errorf("bind: cannot allocate nil pointer in path")
				}
				cur.Set(reflect.New(cur.Type().Elem()))
			}
			cur = cur.Elem()
		}
		if cur.Kind() != reflect.Struct {
			return reflect.Value{}, fmt.Errorf("bind: invalid field path walk")
		}
		cur = cur.FieldByIndex(st.index)
		if !cur.IsValid() {
			return reflect.Value{}, fmt.Errorf("bind: invalid field index %v", st.index)
		}
	}

	return cur, nil
}

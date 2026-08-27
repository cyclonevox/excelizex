package main

import (
	"context"
	"fmt"
	"sync"
)

// enrollment 模拟报名库：按身份证去重。同一人重复上报视为已录取。
type enrollment struct {
	mu      sync.Mutex
	byID    map[string]student
	created []student
}

func newEnrollment() *enrollment {
	return &enrollment{byID: make(map[string]student)}
}

func (e *enrollment) Create(_ context.Context, row student) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if existing, ok := e.byID[row.IDCard]; ok {
		if existing.Name == row.Name {
			return nil
		}
		return fmt.Errorf("身份证已存在: %s（%s）", row.IDCard, existing.Name)
	}
	e.byID[row.IDCard] = row
	e.created = append(e.created, row)
	return nil
}

func (e *enrollment) List() []student {
	e.mu.Lock()
	defer e.mu.Unlock()
	out := make([]student, len(e.created))
	copy(out, e.created)
	return out
}

func (e *enrollment) Count() int {
	e.mu.Lock()
	defer e.mu.Unlock()
	return len(e.created)
}

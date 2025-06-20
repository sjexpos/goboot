package concurrent

import (
	"sync"
)

type GoRoutineLocal[T any] struct {
	values sync.Map
}

func NewGoRoutineLocal[T any]() *GoRoutineLocal[T] {
	return &GoRoutineLocal[T]{}
}

func (t *GoRoutineLocal[T]) Get() *T {
	goId := GoroutineID()
	v, found := t.values.Load(goId)
	if !found {
		return nil
	}
	return v.(*T)
}

func (t *GoRoutineLocal[T]) Set(value *T) {
	goId := GoroutineID()
	t.values.Store(goId, value)
}

func (t *GoRoutineLocal[T]) Clear() {
	goId := GoroutineID()
	t.values.Delete(goId)
}

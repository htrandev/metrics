package pool

import "sync"

type Resettable interface {
	Reset()
}

type ResetPool[T Resettable] struct {
	pool sync.Pool
}

func New[T Resettable](newFunc func() T) *ResetPool[T] {
	return &ResetPool[T]{
		pool: sync.Pool{
			New: func() any {
				return newFunc()
			},
		},
	}
}

func (p *ResetPool[T]) Get() T {
	return p.pool.Get().(T)
}

func (p *ResetPool[T]) Put(t T) {
	t.Reset()
	p.pool.Put(t)
}

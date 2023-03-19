package util

import (
	"go.uber.org/atomic"
	"io"
)

type RefCount[T io.Closer] struct {
	refCount *atomic.Int32
	object   T
	release  func(r *RefCount[T]) error
}

func NewRefCount[T io.Closer](object T, release func(r *RefCount[T]) error) *RefCount[T] {
	return &RefCount[T]{
		refCount: atomic.NewInt32(0),
		object:   object,
		release:  release,
	}
}

// DecRef
// Decrements the reference counting of this object.
// When reference counting hits 0, calls release().
func (r *RefCount[T]) DecRef() error {
	rc := r.refCount.Dec()
	if rc == 0 {
		if err := r.release(r); err != nil {
			r.refCount.Inc()
			return err
		}
	}
	return nil
}

func (r *RefCount[T]) Get() T {
	return r.object
}

// GetRefCount
// Returns the current reference count.
func (r *RefCount[T]) GetRefCount() int {
	return int(r.refCount.Load())
}

// IncRef
// Increments the reference count. Calls to this method must be matched with calls to decRef().
func (r *RefCount[T]) IncRef() {
	r.refCount.Inc()
}

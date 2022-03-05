package executors

import (
	"context"
	"fmt"
	"sync/atomic"
)

type ErrInvalidState struct {
	state uint32
}

func (e ErrInvalidState) Error() string {
	switch e.state {
	case _StateNew:
		return "state:new"
	case _StateCompleting:
		return "state:completing"
	case _StateNormal:
		return "state:normal"
	case _StateCanceled:
		return "state:canceled"
	case _StateError:
		return "state:error"
	default:
		return fmt.Sprintf("unkown state %d", e.state)
	}
}

const (
	_StateNew uint32 = iota
	_StateCompleting
	_StateNormal
	_StateCanceled
	_StateError
)

type FutureTask[T any] struct {
	val      T
	err      error
	callable Callable[T]
	closeCh  chan struct{}
	state    uint32
}

func NewFutureTask[T any](callable Callable[T]) *FutureTask[T] {
	return &FutureTask[T]{
		callable: callable,
		closeCh:  make(chan struct{}),
		state:    _StateNew,
	}
}

// Run implement runnable
// will set f.err if Call return error
func (f *FutureTask[T]) Run(ctx context.Context) {
	if f.state != _StateNew {
		return
	}

	val, err := f.callable.Call(ctx)
	if err != nil {
		f.setError(err)
		return
	}
	f.set(val)
	return
}

func (f *FutureTask[T]) Get(ctx context.Context) (T, error) {
	if f.Completed() {
		return f.report(atomic.LoadUint32(&f.state))
	}
	select {
	case <-ctx.Done():
		return f.val, ctx.Err()
	case <-f.closeCh:
		return f.report(atomic.LoadUint32(&f.state))
	}
}

func (f *FutureTask[T]) report(state uint32) (T, error) {
	switch state {
	case _StateNormal:
		return f.val, nil
	case _StateCanceled:
		return f.val, ErrFutureCanceled
	case _StateError:
		return f.val, f.err
	}
	panic(ErrInvalidState{state: f.state})
}

func (f *FutureTask[T]) set(val T) {
	state := &f.state
	if atomic.CompareAndSwapUint32(state, _StateNew, _StateCompleting) {
		f.val = val
		atomic.StoreUint32(state, _StateNormal)
		close(f.closeCh)
	}
}

func (f *FutureTask[T]) setError(err error) {
	state := &f.state
	if atomic.CompareAndSwapUint32(state, _StateNew, _StateCompleting) {
		f.err = err
		atomic.StoreUint32(state, _StateError)
		close(f.closeCh)
	}
}

func (f *FutureTask[T]) Cancel() bool {
	if f.state != _StateNew {
		return false
	}
	ok := atomic.CompareAndSwapUint32(&f.state, _StateNew, _StateCanceled)
	if ok {
		close(f.closeCh)
		return true
	}
	return false
}

func (f *FutureTask[T]) Canceled() bool {
	state := atomic.LoadUint32(&f.state)
	return state == _StateCanceled
}

func (f *FutureTask[T]) Completed() bool {
	state := atomic.LoadUint32(&f.state)
	return state >= _StateNormal
}

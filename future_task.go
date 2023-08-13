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
		return fmt.Sprintf("unknown state %d", e.state)
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
	val        T
	err        error
	callable   Callable[T]
	closeCh    chan struct{}
	state      uint32
	cancelFunc context.CancelFunc
	thenFunc   ThenFunction[T]
	catchFunc  CatchFunction
}

func NewFutureTask[T any](callable Callable[T]) *FutureTask[T] {
	return &FutureTask[T]{
		callable: callable,
		closeCh:  make(chan struct{}),
		state:    _StateNew,
	}
}

// Run implement runnable
func (f *FutureTask[T]) Run(ctx context.Context) {
	if f.state != _StateNew {
		return
	}

	ctx, f.cancelFunc = context.WithCancel(ctx)
	val, err := f.callable.Call(ctx)
	if err != nil {
		f.completeError(err)
		return
	}
	f.completeValue(val)
}

func (f *FutureTask[T]) Get(ctx context.Context) (T, error) {
	if f.Completed() {
		return f.report(atomic.LoadUint32(&f.state))
	}
	select {
	case <-ctx.Done():
		f.cancelCallable()
		return f.val, ctx.Err()
	case <-f.closeCh:
		return f.report(atomic.LoadUint32(&f.state))
	}
}

func (f *FutureTask[T]) Then(thenFunc ThenFunction[T]) Future[T] {
	f.thenFunc = thenFunc
	if f.Completed() {
		f.postComplete()
	}
	return f
}

func (f *FutureTask[T]) Catch(catchFunc CatchFunction) Future[T] {
	f.catchFunc = catchFunc
	if f.Completed() {
		f.postComplete()
	}
	return f
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

func (f *FutureTask[T]) completeValue(val T) {
	state := &f.state
	if atomic.CompareAndSwapUint32(state, _StateNew, _StateCompleting) {
		f.val = val
		atomic.StoreUint32(state, _StateNormal)
		close(f.closeCh)
		f.postComplete()
	}
}

func (f *FutureTask[T]) completeError(err error) {
	state := &f.state
	if atomic.CompareAndSwapUint32(state, _StateNew, _StateCompleting) {
		f.err = err
		atomic.StoreUint32(state, _StateError)
		close(f.closeCh)
		f.postComplete()
	}
}

func (f *FutureTask[T]) postComplete() {
	if f.CompletedError() {
		if f.catchFunc != nil {
			f.catchFunc(f.err)
		}
	} else {
		if f.thenFunc != nil {
			f.thenFunc(f.val)
		}
	}
}

func (f *FutureTask[T]) Cancel() bool {
	if f.state != _StateNew {
		return false
	}
	ok := atomic.CompareAndSwapUint32(&f.state, _StateNew, _StateCanceled)
	if ok {
		f.cancelCallable()
		f.err = ErrFutureCanceled
		close(f.closeCh)
		f.postComplete()
		return true
	}
	return false
}

func (f *FutureTask[T]) cancelCallable() {
	if f.cancelFunc != nil {
		f.cancelFunc()
	}
}

func (f *FutureTask[T]) Canceled() bool {
	state := atomic.LoadUint32(&f.state)
	return state == _StateCanceled
}

func (f *FutureTask[T]) Completed() bool {
	state := atomic.LoadUint32(&f.state)
	return state >= _StateNormal
}

func (f *FutureTask[T]) CompletedError() bool {
	state := atomic.LoadUint32(&f.state)
	return state != _StateNormal
}

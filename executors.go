package executors

import (
	"context"
	"errors"
)

var (
	ErrRejectedExecution = errors.New("rejected execution")
	ErrClosed            = errors.New("closed")
)

type Runnable interface {
	Run(ctx context.Context) error
}

type Callable[T any] interface {
	Call(ctx context.Context) (T, error)
}

type RunnableFunc func(ctx context.Context) error

func (r RunnableFunc) Run(ctx context.Context) error {
	return r(ctx)
}

type CallableFunc[T any] func(ctx context.Context) (T, error)

func (c CallableFunc[T]) Call(ctx context.Context) (T, error) {
	return c(ctx)
}

type Executor interface {
	Execute(Runnable) error
	Shutdown(ctx context.Context) error
}

type ExecutorService[T any] interface {
	Executor
	Submit(callable Callable[T]) (Future[T], error)
}

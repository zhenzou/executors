package executors

import (
	"context"
	"errors"
	"fmt"
	"time"
)

var (
	ErrRejectedExecution = errors.New("rejected execution")
	ErrClosed            = errors.New("closed")
)

type CancelFunc = func()

type ErrPanic struct {
	Cause interface{}
}

func (e ErrPanic) Error() string {
	return fmt.Sprintf("%v", e.Cause)
}

type Runnable interface {
	Run(ctx context.Context)
}

type Callable[T any] interface {
	Call(ctx context.Context) (T, error)
}

type RunnableFunc func(ctx context.Context)

func (r RunnableFunc) Run(ctx context.Context) {
	r(ctx)
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

type ScheduledExecutor interface {
	Executor
	Schedule(r Runnable, delay time.Duration) (CancelFunc, error)
	ScheduleAtFixRate(r Runnable, period time.Duration) (CancelFunc, error)
}

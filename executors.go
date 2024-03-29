package executors

import (
	"context"
	"errors"
	"fmt"
	"time"
)

var (
	ErrRejectedExecution   = errors.New("rejected execution")
	ErrShutdown            = errors.New("shutdown")
	ErrInvalidCronExpr     = errors.New("invalid corn expr")
	ErrInvalidCronTimezone = errors.New("invalid corn timezone")
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
	// Execute execute a task in background
	// Will return ErrShutdown if shutdown already
	// Will return ErrRejectedExecution if task out of cap
	Execute(Runnable) error

	// Shutdown shutdown the executor
	// Will wait the queued task to be finish
	Shutdown(ctx context.Context) error
}

type ExecutorService[T any] interface {
	Executor

	// Submit execute a task with result async, and can get the task result via get
	Submit(callable Callable[T]) (Future[T], error)
}

type CRONRule struct {
	// Expr cron expr
	Expr string `json:"expr,omitempty"`

	// Timezone default UTC
	Timezone string `json:"timezone,omitempty"`
}

type ScheduledExecutor interface {
	Executor
	// Schedule run a one time task after delay duration
	Schedule(r Runnable, delay time.Duration) (CancelFunc, error)

	// ScheduleAtFixRate schedule a periodic task in fixed rate
	ScheduleAtFixRate(r Runnable, period time.Duration) (CancelFunc, error)

	// ScheduleAtCronRate schedule at periodic cron task
	ScheduleAtCronRate(r Runnable, rule CRONRule) (CancelFunc, error)
}

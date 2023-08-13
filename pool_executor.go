package executors

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/panjf2000/ants/v2"
)

type _PoolExecutorOption func(opts *poolExecutorOptions)

type poolExecutorOptions struct {
	MaxConcurrent    int
	MaxBlockingTasks int
	ExecuteTimeout   time.Duration
	ErrorHandler     ExceptionHandler
	RejectionHandler RejectionHandler
	Logger           *slog.Logger
}

var _DefaultPoolExecutorOptions = poolExecutorOptions{
	MaxConcurrent:    10,
	ExecuteTimeout:   0,
	ErrorHandler:     NoopErrorHandler{},
	RejectionHandler: NoopRejectionPolicy{},
	Logger:           slog.Default(),
}

func WithMaxConcurrent(concurrent int) _PoolExecutorOption {
	return func(opts *poolExecutorOptions) {
		opts.MaxConcurrent = concurrent
	}
}

func WithMaxBlockingTasks(max int) _PoolExecutorOption {
	return func(opts *poolExecutorOptions) {
		opts.MaxBlockingTasks = max
	}
}

func WithExecuteTimeout(ts time.Duration) _PoolExecutorOption {
	return func(opts *poolExecutorOptions) {
		opts.ExecuteTimeout = ts
	}
}

func WithRejectionHandler(handler RejectionHandler) _PoolExecutorOption {
	return func(opts *poolExecutorOptions) {
		opts.RejectionHandler = handler
	}
}

func WithErrorHandler(handler ExceptionHandler) _PoolExecutorOption {
	return func(opts *poolExecutorOptions) {
		opts.ErrorHandler = handler
	}
}

func WithLogger(logger *slog.Logger) _PoolExecutorOption {
	return func(opts *poolExecutorOptions) {
		opts.Logger = logger
	}
}

type RejectionHandler interface {
	RejectExecution(runnable Runnable, e Executor) error
}

type ExceptionHandler interface {
	CatchException(runnable Runnable, e error)
}

type ErrorHandlerFunc func(runnable Runnable, e error)

func (f ErrorHandlerFunc) CatchException(runnable Runnable, e error) {
	f(runnable, e)
}

func NewPoolExecutor(opts ..._PoolExecutorOption) Executor {
	return NewPoolExecutorService[any](opts...)
}

func NewPoolExecutorService[T any](opts ..._PoolExecutorOption) ExecutorService[T] {
	return _NewPoolExecutorService[T](opts...)
}

func _NewPoolExecutorService[T any](opts ..._PoolExecutorOption) *PoolExecutor[T] {
	var opt = _DefaultPoolExecutorOptions
	for _, o := range opts {
		o(&opt)
	}
	pool, err := ants.NewPool(opt.MaxConcurrent,
		ants.WithMaxBlockingTasks(opt.MaxBlockingTasks),
		// do nothing, will handle by ExceptionHandler
		ants.WithPanicHandler(func(cause interface{}) {}))
	if err != nil {
		panic(err)
	}
	return &PoolExecutor[T]{
		opts: opt,
		pool: pool,
	}
}

type PoolExecutor[T any] struct {
	opts poolExecutorOptions
	pool *ants.Pool
}

func (p *PoolExecutor[T]) Execute(r Runnable) error {
	err := p.pool.Submit(func() {
		ctx, cancelFunc := p.newContext()
		defer cancelFunc()
		defer func() {
			if cause := recover(); cause != nil {
				p.opts.ErrorHandler.CatchException(r, ErrPanic{Cause: cause})
			}
		}()
		r.Run(ctx)
	})

	if err == nil {
		return nil
	}

	switch {
	case errors.Is(err, ants.ErrPoolClosed):
		return ErrShutdown
	case errors.Is(err, ants.ErrPoolOverload):
		return p.opts.RejectionHandler.RejectExecution(r, p)
	default:
		return err
	}
}

func (p *PoolExecutor[T]) Submit(callable Callable[T]) (Future[T], error) {
	f := NewFutureTask[T](callable)
	err := p.Execute(f)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func (p *PoolExecutor[T]) Shutdown(ctx context.Context) error {
	ch := make(chan struct{})
	go func() {
		p.pool.Release()
		ch <- struct{}{}
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-ch:
		return nil
	}
}

func (p *PoolExecutor[T]) newContext() (context.Context, context.CancelFunc) {
	if p.opts.ExecuteTimeout == 0 {
		return context.WithCancel(context.Background())
	}
	return context.WithTimeout(context.Background(), p.opts.ExecuteTimeout)
}

type NoopErrorHandler struct {
}

func (d NoopErrorHandler) CatchException(runnable Runnable, e error) {
	panic(e)
}

type DiscardErrorHandler struct {
}

func (d DiscardErrorHandler) CatchException(runnable Runnable, e error) {
}

type NoopRejectionPolicy struct {
}

func (d NoopRejectionPolicy) RejectExecution(runnable Runnable, e Executor) error {
	return ErrRejectedExecution
}

type DiscardRejectionPolicy struct {
}

func (d DiscardRejectionPolicy) RejectExecution(runnable Runnable, e Executor) error {
	return nil
}

type CallerRunsRejectionPolicy struct {
}

func (d CallerRunsRejectionPolicy) RejectExecution(runnable Runnable, e Executor) error {
	runnable.Run(context.Background())
	return nil
}

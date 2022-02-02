package executors

import (
	"context"
	"errors"
	"time"

	"github.com/panjf2000/ants/v2"
)

type poolExecutorOption func(opts *poolExecutorOptions)

type poolExecutorOptions struct {
	MaxConcurrent    int
	MaxBlockingTasks int
	ExecuteTimeout   time.Duration
	ErrorHandler     ErrorHandler
	RejectionHandler RejectionHandler
}

var defaultPoolExecutorOptions = poolExecutorOptions{
	MaxConcurrent:    10,
	ExecuteTimeout:   0,
	ErrorHandler:     DiscardErrorHandler{},
	RejectionHandler: DiscardRejectionPolicy{},
}

func WithMaxConcurrent(concurrent int) poolExecutorOption {
	return func(opts *poolExecutorOptions) {
		opts.MaxConcurrent = concurrent
	}
}

func WithMaxBlockingTasks(max int) poolExecutorOption {
	return func(opts *poolExecutorOptions) {
		opts.MaxBlockingTasks = max
	}
}

func WithExecuteTimeout(ts time.Duration) poolExecutorOption {
	return func(opts *poolExecutorOptions) {
		opts.ExecuteTimeout = ts
	}
}

func WithRejectionHandler(handler RejectionHandler) poolExecutorOption {
	return func(opts *poolExecutorOptions) {
		opts.RejectionHandler = handler
	}
}

type RejectionHandler interface {
	RejectExecution(runnable Runnable, e Executor) error
}

type ErrorHandler interface {
	HandlerError(runnable Runnable, e error)
}

func NewPoolExecutor(opts ...poolExecutorOption) Executor {
	var opt = defaultPoolExecutorOptions
	for _, o := range opts {
		o(&opt)
	}
	pool, err := ants.NewPool(opt.MaxConcurrent, ants.WithMaxBlockingTasks(opt.MaxBlockingTasks))
	if err != nil {
		panic(err)
	}
	return &PoolExecutor[any]{
		opts: opt,
		pool: pool,
	}
}

func NewPoolExecutorService[T any](opts ...poolExecutorOption) ExecutorService[T] {
	var opt = defaultPoolExecutorOptions
	for _, o := range opts {
		o(&opt)
	}
	pool, err := ants.NewPool(opt.MaxConcurrent)
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
		err := r.Run(ctx)
		if err != nil {
			p.opts.ErrorHandler.HandlerError(r, err)
		}
	})

	if err == nil {
		return nil
	}

	switch {
	case errors.Is(err, ants.ErrPoolClosed):
		return ErrClosed
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

type DiscardErrorHandler struct {
}

func (d DiscardErrorHandler) HandlerError(runnable Runnable, e error) {
}

type DiscardRejectionPolicy struct {
}

func (d DiscardRejectionPolicy) RejectExecution(runnable Runnable, e Executor) error {
	return nil
}

type CallerRunsRejectionPolicy struct {
}

func (d CallerRunsRejectionPolicy) RejectExecution(runnable Runnable, e Executor) error {
	return runnable.Run(context.Background())
}

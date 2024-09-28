package executors

import (
	"context"
	"errors"

	"github.com/panjf2000/ants/v2"
)

func NewPoolExecutor(opts ..._PoolExecutorOption) Executor {
	return NewPoolExecutorService[any](opts...)
}

func NewPoolExecutorService[T any](opts ..._PoolExecutorOption) ExecutorService[T] {
	return internalNewPoolExecutorService[T](opts...)
}

func internalNewPoolExecutorService[T any](opts ..._PoolExecutorOption) *PoolExecutor[T] {
	var opt = _DefaultPoolExecutorOptions
	for _, o := range opts {
		o(&opt)
	}
	pool, err := ants.NewPool(opt.MaxConcurrent,
		ants.WithMaxBlockingTasks(opt.MaxBlockingTasks),
		// do nothing, will handle by ErrorHandler
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
				p.opts.Logger.Debug("failed to execute task")
				p.opts.ErrorHandler.CatchError(r, ErrPanic{Cause: cause})
			}
		}()
		r.Run(ctx)
	})

	if err == nil {
		p.opts.Logger.Debug("submitted a new task")
		return nil
	}

	p.opts.Logger.Debug("failed to submit task")

	switch {
	case errors.Is(err, ants.ErrPoolClosed):
		return ErrShutdown
	case errors.Is(err, ants.ErrPoolOverload):
		return p.opts.RejectionHandler.RejectExecution(r, p)
	default:
		return err
	}
}

func (p *PoolExecutor[T]) ExecuteFunc(fn func(ctx context.Context)) error {
	return p.Execute(RunnableFunc(fn))
}

func (p *PoolExecutor[T]) Submit(callable Callable[T]) (Future[T], error) {
	f := NewFutureTask[T](callable)
	err := p.Execute(f)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func (p *PoolExecutor[T]) SubmitFunc(fn func(ctx context.Context) (T, error)) (Future[T], error) {
	return p.Submit(CallableFunc[T](fn))
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

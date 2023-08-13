package executors

import (
	"log/slog"
	"time"
)

type RejectionHandler interface {
	RejectExecution(runnable Runnable, e Executor) error
}

type ErrorHandler interface {
	CatchError(runnable Runnable, e error)
}

type ErrorHandlerFunc func(runnable Runnable, e error)

func (f ErrorHandlerFunc) CatchError(runnable Runnable, e error) {
	f(runnable, e)
}

type _PoolExecutorOption func(opts *poolExecutorOptions)

type poolExecutorOptions struct {
	MaxConcurrent    int
	MaxBlockingTasks int
	ExecuteTimeout   time.Duration
	ErrorHandler     ErrorHandler
	RejectionHandler RejectionHandler
	Logger           *slog.Logger
}

var _DefaultPoolExecutorOptions = poolExecutorOptions{
	MaxConcurrent:    10,
	ExecuteTimeout:   0,
	ErrorHandler:     LogErrorHandler{},
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

func WithErrorHandler(handler ErrorHandler) _PoolExecutorOption {
	return func(opts *poolExecutorOptions) {
		opts.ErrorHandler = handler
	}
}

func WithLogger(logger *slog.Logger) _PoolExecutorOption {
	return func(opts *poolExecutorOptions) {
		opts.Logger = logger
	}
}

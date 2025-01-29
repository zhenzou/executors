package executors

import (
	"context"
	"errors"
)

var (
	ErrFutureCanceled = errors.New("future canceled")
)

type ThenFunction[T any] func(val T) error
type CatchFunction func(err error)

type Future[T any] interface {
	Get(ctx context.Context) (T, error)
	Then(thenFunc ThenFunction[T]) NotThenableFuture[T]
	Catch(catchFunc CatchFunction) NotChainableFuture[T]
	Cancel() bool
	Canceled() bool
	Completed() bool
	CompletedError() bool
}

type NotThenableFuture[T any] interface {
	Get(ctx context.Context) (T, error)
	Catch(catchFunc CatchFunction) NotChainableFuture[T]
	Cancel() bool
	Canceled() bool
	Completed() bool
	CompletedError() bool
}

type NotChainableFuture[T any] interface {
	Get(ctx context.Context) (T, error)
	Cancel() bool
	Canceled() bool
	Completed() bool
	CompletedError() bool
}

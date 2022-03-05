package executors

import (
	"context"
	"errors"
)

var (
	ErrFutureCanceled = errors.New("future canceled")
)

type ThenFunction[T any] func(val T)
type CatchFunction func(err error)

type Future[T any] interface {
	Get(ctx context.Context) (T, error)
	Then(thenFunc ThenFunction[T])
	Catch(catchFunc CatchFunction)
	Cancel() bool
	Canceled() bool
	Completed() bool
	CompletedError() bool
}

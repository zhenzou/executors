package executors

import (
	"context"
	"errors"
)

var (
	ErrFutureCanceled = errors.New("future canceled")
)

type Future[T any] interface {
	Get(ctx context.Context) (T, error)
	Cancel() bool
	Canceled() bool
	Completed() bool
}

package executors

import (
	"context"
	"log/slog"
)

type NoopErrorHandler struct {
}

func (d NoopErrorHandler) CatchError(runnable Runnable, e error) {
	panic(e)
}

type LogErrorHandler struct {
}

func (d LogErrorHandler) CatchError(runnable Runnable, e error) {
	slog.Error("catch error", slog.Any("cause", e))
}

type DiscardErrorHandler struct {
}

func (d DiscardErrorHandler) CatchError(runnable Runnable, e error) {
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

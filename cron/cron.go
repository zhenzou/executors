package cron

import (
	"sync/atomic"
	"time"

	"github.com/aptible/supercronic/cronexpr"
)

type task[T any] struct {
	ID          int32
	Expr        *cronexpr.Expression
	Task        T
	LastRunTime time.Time
	NextRunTime time.Time
	Location    *time.Location

	nowFn func() time.Time
}

var (
	taskIdGenerator atomic.Int32
)

func nextTaskID() int32 {
	return taskIdGenerator.Add(1)
}

func newTask[T any](r T, expr *cronexpr.Expression, location *time.Location, nowFn func() time.Time) *task[T] {
	task := &task[T]{
		ID:          nextTaskID(),
		Expr:        expr,
		Task:        r,
		NextRunTime: nowFn(),
		nowFn:       nowFn,
		Location:    location,
	}

	task.scheduleNextRun()

	return task
}

func (t *task[T]) now() time.Time {
	if t.nowFn == nil {
		return time.Now()
	}
	return t.nowFn()
}

func (t *task[T]) ready() bool {
	return !t.NextRunTime.After(t.now().In(t.Location))
}

func (t *task[T]) scheduleNextRun() {
	t.LastRunTime = t.NextRunTime
	t.NextRunTime = t.Expr.Next(t.LastRunTime.In(t.Location)).Truncate(time.Second)
}

func (t *task[T]) untilNextRun() time.Duration {
	if t.ready() {
		return 0
	}
	return t.NextRunTime.Sub(t.now()).Truncate(time.Second)
}

type Dispatcher[T any] interface {
	// AddTask return func to remove task
	AddTask(r T, expr *cronexpr.Expression, location *time.Location) func()

	Shutdown()

	// GetReadyTask get ready task chain
	GetReadyTask() <-chan T
}

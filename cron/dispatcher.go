package cron

import (
	"log/slog"
	"sync"
	"time"

	"github.com/aptible/supercronic/cronexpr"
	"github.com/zyedidia/generic/heap"

	"github.com/zhenzou/executors/routine"
	"github.com/zhenzou/executors/sleeper"
)

func NewDispatcher[T any](logger *slog.Logger) Dispatcher[T] {
	return &dispatcher[T]{
		heap:      heap.New[*task[T]](taskLessThan[T]),
		locker:    &sync.Mutex{},
		readyChan: make(chan T, 1),
		close:     make(chan struct{}, 1),
		logger:    logger,
		sleeper:   sleeper.NewSleeper(),
	}
}

func taskLessThan[T any](a, b *task[T]) bool {
	return a.NextRunTime.Before(b.NextRunTime)
}

type dispatcher[T any] struct {
	heap      *heap.Heap[*task[T]]
	locker    sync.Locker
	readyChan chan T
	close     chan struct{}
	closed    bool
	loopOnce  sync.Once
	logger    *slog.Logger
	sleeper   sleeper.Sleeper
}

func (d *dispatcher[T]) AddTask(r T, expr *cronexpr.Expression, location *time.Location) func() {
	t := newTask(r, expr, location)

	d.locker.Lock()
	defer d.locker.Unlock()

	d.heap.Push(t)

	d.sleeper.Wakeup()

	return d.getRemoveFunc(t)
}

func (d *dispatcher[T]) Close() {
	close(d.close)
	d.sleeper.Wakeup()
	close(d.readyChan)
}

func (d *dispatcher[T]) getRemoveFunc(t *task[T]) func() {
	return func() { d.removeTask(t) }
}

func (d *dispatcher[T]) removeTask(t *task[T]) {
	d.locker.Lock()
	defer d.locker.Unlock()

	var tasks []*task[T]

	for {
		task, ok := d.heap.Pop()
		if !ok {
			break
		}
		if task.ID == t.ID {
			continue
		}
		tasks = append(tasks, task)
	}

	d.heap = heap.FromSlice[*task[T]](taskLessThan[T], tasks)
}

func (d *dispatcher[T]) GetReadyTask() <-chan T {
	d.loopOnce.Do(d.loopTakeReadyTask)
	return d.readyChan
}

func (d *dispatcher[T]) loopTakeReadyTask() {
	routine.GoWithRecovery(d.logger, func() {
		for {
			select {
			case <-d.close:
				d.logger.Info("dispatcher closed")
				d.closed = true
				return
			default:
				duration, take := d.takeReadyTask()
				if !take {
					d.sleeper.Sleep(duration)
				}
			}
		}
	}, d.loopTakeReadyTask)
}

const (
	maxYieldDuration = 10 * time.Minute
)

func getYieldDuration[T any](t *task[T]) time.Duration {
	return min(maxYieldDuration, t.untilNextRun())
}

// will return yield time if false
// will return 0 if true
func (d *dispatcher[T]) takeReadyTask() (time.Duration, bool) {
	d.locker.Lock()
	defer d.locker.Unlock()

	t, ok := d.heap.Peek()

	if !ok {
		return maxYieldDuration, false
	}
	if !t.ready() {
		return getYieldDuration(t), false
	}

	_, _ = d.heap.Pop()
	t.scheduleNextRun()
	d.heap.Push(t)

	d.readyChan <- t.Task
	return 0, true
}

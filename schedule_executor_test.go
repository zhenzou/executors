package executors

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPoolScheduleExecutor_ScheduleAtFixRate(t *testing.T) {

	scheduleExecutor := NewPoolScheduleExecutor(WithMaxConcurrent(10))

	var i int64 = 10
	_, _ = scheduleExecutor.ScheduleAtFixRate(RunnableFunc(func(ctx context.Context) {
		atomic.AddInt64(&i, 10)
	},
	), 100*time.Millisecond)

	time.AfterFunc(1*time.Second, func() {
		assert.GreaterOrEqual(t, atomic.LoadInt64(&i), int64(100))
	})

	time.Sleep(2 * time.Second)
	_ = scheduleExecutor.Shutdown(context.Background())
}

func TestPoolScheduleExecutor_Schedule(t *testing.T) {

	scheduleExecutor := NewPoolScheduleExecutor(WithMaxConcurrent(10))

	var i int64 = 10
	_, _ = scheduleExecutor.Schedule(RunnableFunc(func(ctx context.Context) {
		atomic.AddInt64(&i, 10)
	},
	), 500*time.Millisecond)

	time.AfterFunc(1*time.Second, func() {
		assert.Equal(t, int64(20), atomic.LoadInt64(&i))
	})

	time.Sleep(2 * time.Second)
	_ = scheduleExecutor.Shutdown(context.Background())
}

func TestPoolScheduleExecutor_ScheduleAtCronRate(t *testing.T) {
	scheduleExecutor := NewPoolScheduleExecutor(WithMaxConcurrent(10))

	var i int64 = 10
	_, _ = scheduleExecutor.ScheduleAtCronRate(RunnableFunc(func(ctx context.Context) {
		atomic.AddInt64(&i, 10)
	},
	), CRONRule{
		Expr:     "*/2 * * * * * *",
		Timezone: "",
	})

	time.AfterFunc(3*time.Second, func() {
		assert.LessOrEqual(t, int64(20), atomic.LoadInt64(&i))
	})

	time.Sleep(3 * time.Second)
	_ = scheduleExecutor.Shutdown(context.Background())
}

package executors

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestPoolScheduleExecutor_ScheduleAtFixRate(t *testing.T) {

	scheduleExecutor := NewPoolScheduleExecutor(WithMaxConcurrent(10))

	ch1 := make(chan struct{})
	ch2 := make(chan struct{})

	t.Run("schedule", func(t *testing.T) {
		go func() {
			var i = 10
			_, _ = scheduleExecutor.Schedule(RunnableFunc(func(ctx context.Context) { i += 10 }), 500*time.Millisecond)
			time.AfterFunc(1*time.Second, func() {
				require.Equal(t, 20, i)
			})
			println(i)
			ch1 <- struct{}{}
		}()
	})

	t.Run("fix rate", func(t *testing.T) {
		go func() {
			var i = 10
			_, _ = scheduleExecutor.ScheduleAtFixRate(RunnableFunc(func(ctx context.Context) { i += 10 }), 100*time.Millisecond)

			time.AfterFunc(1*time.Second, func() {
				require.Equal(t, 100, i)
			})
			println(i)
			ch2 <- struct{}{}
		}()
	})

	<-ch1
	<-ch2
	_ = scheduleExecutor.Shutdown(context.Background())
}

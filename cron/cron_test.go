package cron

import (
	"testing"
	"time"

	"github.com/aptible/supercronic/cronexpr"
	"github.com/stretchr/testify/require"
)

func Test_task_ready(t1 *testing.T) {
	type testCase[T any] struct {
		name string
		t    task[T]
		want bool
	}
	tests := []testCase[any]{
		{
			name: "given run at tomorrow, will got false",
			t: task[any]{
				NextRunTime: time.Now().AddDate(0, 0, 1),
				Location:    time.UTC,
			},
			want: false,
		},
		{
			name: "given run at now, will got true",
			t: task[any]{
				NextRunTime: time.Now(),
				Location:    time.UTC,
			},
			want: true,
		},
		{
			name: "given run at yesterday, will got true",
			t: task[any]{
				NextRunTime: time.Now().AddDate(0, 0, -1),
				Location:    time.UTC,
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			if got := tt.t.ready(); got != tt.want {
				t1.Errorf("ready() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_task_scheduleNextRun(t *testing.T) {

	t.Run("given run at 30m of  hour", func(t *testing.T) {
		ts := time.Date(2023, 8, 12, 0, 0, 0, 0, time.UTC)

		task := task[any]{
			Expr:        cronexpr.MustParse("30 * * * *"),
			NextRunTime: ts,
			Location:    time.UTC,
		}
		task.scheduleNextRun()

		require.Equal(t, ts, task.LastRunTime)
		require.Equal(t, ts.Add(30*time.Minute), task.NextRunTime)
	})

	t.Run("given run every 5 min", func(t *testing.T) {
		ts := time.Date(2023, 8, 12, 0, 0, 0, 0, time.UTC)

		task := task[any]{
			Expr:        cronexpr.MustParse("*/5 * * * *"),
			NextRunTime: ts,
			Location:    time.UTC,
		}
		task.scheduleNextRun()

		require.Equal(t, ts, task.LastRunTime)
		require.Equal(t, ts.Add(5*time.Minute), task.NextRunTime)

		task.scheduleNextRun()
		require.Equal(t, ts.Add(5*time.Minute), task.LastRunTime)
		require.Equal(t, ts.Add(10*time.Minute), task.NextRunTime)
	})

}

func Test_task_untilNextRun(t1 *testing.T) {
	type testCase[T any] struct {
		name string
		t    task[T]
		want time.Duration
	}
	now := time.Now()
	tests := []testCase[any]{
		{
			name: "given run after 30min, will got 30 min",
			t: task[any]{
				NextRunTime: now.Add(30 * time.Minute),
				Location:    time.UTC,
				nowFn: func() time.Time {
					return now
				},
			},
			want: 30 * time.Minute,
		},
		{
			name: "given run at now, will got zero",
			t: task[any]{
				NextRunTime: now,
				nowFn: func() time.Time {
					return now
				},
				Location: time.UTC,
			},
			want: 0,
		},
		{
			name: "given run at yesterday, will got 0",
			t: task[any]{
				NextRunTime: now.AddDate(0, 0, -1),
				Location:    time.UTC,
			},
			want: 0,
		},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			if got := tt.t.untilNextRun(); got != tt.want {
				t1.Errorf("untilNextRun() = %v, want %v", got, tt.want)
			}
		})
	}
}

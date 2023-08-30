package cron

import (
	"log/slog"
	"testing"
	"time"

	"github.com/aptible/supercronic/cronexpr"
	"github.com/stretchr/testify/require"
)

func Test_getYieldDuration(t *testing.T) {
	type args[T any] struct {
		t *task[T]
	}
	type testCase[T any] struct {
		name string
		args args[T]
		want time.Duration
	}
	now := time.Now()
	tests := []testCase[any]{
		{
			name: "given run after 30 min, should got 10min",
			args: args[any]{
				t: &task[any]{
					NextRunTime: now.Add(30 * time.Minute),
					Location:    time.UTC,
					nowFn: func() time.Time {
						return now
					},
				},
			},
			want: maxYieldDuration,
		},
		{
			name: "given run after 30 secs, should got 30 secs",
			args: args[any]{
				t: &task[any]{
					NextRunTime: now.Add(30 * time.Second),
					Location:    time.UTC,
					nowFn: func() time.Time {
						return now
					},
				},
			},
			want: 30 * time.Second,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getYieldDuration(tt.args.t); got != tt.want {
				t.Errorf("getYieldDuration() = %v, want %v", got, tt.want)
			}
		})
	}
}

type F func()

func Test_dispatcher_GetReadyTask(t *testing.T) {
	dispatcher := NewDispatcher[F](slog.Default())
	counter := 0
	f := func() {
		counter++
		// println(counter)
	}

	dispatcher.AddTask(f, cronexpr.MustParse("*/1 * * * * * *"), time.UTC)

	startAt := time.Now()
	ch := dispatcher.GetReadyTask()

	for i := 0; i < 6; i++ {
		next := <-ch
		next()
		require.Equal(t, i+1, counter)
	}

	require.True(t, time.Since(startAt) > 5*time.Second)
}

func Test_dispatcher_Close(t *testing.T) {
	dispatcher := NewDispatcher[F](slog.Default()).(*dispatcher[F])
	counter := 0
	f := func() {
		counter++
		// println(counter)
	}

	dispatcher.AddTask(f, cronexpr.MustParse("*/1 * * * * * *"), time.UTC)

	ch := dispatcher.GetReadyTask()

	dispatcher.Shutdown()

	for i := 0; i < 6; i++ {
		next, ok := <-ch
		require.False(t, ok)
		require.Nil(t, next)
	}

	for f := range ch {
		// not run
		require.Nil(t, f)
	}

	dispatcher.sleeper.Wakeup()
	time.Sleep(1 * time.Millisecond)
	require.True(t, dispatcher.closed)
}

type Person struct {
	Name string
}

func Test_dispatcher_takeReadyTask(t *testing.T) {

	dispatcher := NewDispatcher[Person](slog.Default()).(*dispatcher[Person])

	// 11 sec
	now := time.Date(2023, 8, 13, 12, 0, 11, 0, time.UTC)
	dispatcher.nowFn = func() time.Time {
		return now
	}

	p1 := Person{Name: "p1"}
	p2 := Person{Name: "p2"}

	dispatcher.AddTask(p1, cronexpr.MustParse("*/2 * * * * * *"), time.UTC)
	dispatcher.AddTask(p2, cronexpr.MustParse("*/5 * * * * * *"), time.UTC)

	go func() {
		ch := dispatcher.GetReadyTask()

		for p := range ch {
			require.NotEmpty(t, p.Name)
		}
	}()

	duration, ok := dispatcher.takeReadyTask()
	require.False(t, ok)
	require.Equal(t, 1*time.Second, duration)
	peek, ok := dispatcher.heap.Peek()
	require.True(t, ok)
	require.Equal(t, p1, peek.Task)

	// 12 sec
	now = now.Add(1 * time.Second)

	_, ok = dispatcher.takeReadyTask()
	require.True(t, ok)
	require.Equal(t, 2, dispatcher.heap.Size())
	peek, ok = dispatcher.heap.Peek()
	require.True(t, ok)
	require.Equal(t, p1, peek.Task)

	// 14 sec
	now = now.Add(2 * time.Second)
	_, ok = dispatcher.takeReadyTask()
	require.True(t, ok)
	peek, ok = dispatcher.heap.Peek()
	require.True(t, ok)
	require.Equal(t, p2, peek.Task)

	// 15 sec
	now = now.Add(1 * time.Second)
	_, ok = dispatcher.takeReadyTask()
	require.True(t, ok)
	peek, ok = dispatcher.heap.Peek()
	require.True(t, ok)
	require.Equal(t, p1, peek.Task)
}

func Test_dispatcher_removeTask(t *testing.T) {
	dispatcher := NewDispatcher[Person](slog.Default()).(*dispatcher[Person])

	p1 := Person{Name: "p1"}
	p2 := Person{Name: "p2"}

	removeTaskP1 := dispatcher.AddTask(p1, cronexpr.MustParse("*/2 * * * * * *"), time.UTC)
	dispatcher.AddTask(p2, cronexpr.MustParse("*/3 * * * * * *"), time.UTC)

	removeTaskP1()
	require.Equal(t, 1, dispatcher.heap.Size())

	peek, ok := dispatcher.heap.Peek()
	require.True(t, ok)
	require.Equal(t, p2, peek.Task)
}

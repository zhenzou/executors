package sleeper

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func Test_sleeper_Wakeup(t *testing.T) {
	s := NewSleeper()

	var i int

	t.Cleanup(func() {
		require.Equal(t, 1, i)
	})

	// should not block if no waiting
	s.Wakeup()

	s.Wakeup()

	s.Wakeup()

	i++
}

func Test_sleeper_Sleep(t *testing.T) {
	s := NewSleeper()

	var counter atomic.Int32
	t.Cleanup(func() {
		require.Equal(t, int32(10), counter.Load())
	})

	go func() {
		start := time.Now()

		s.Sleep(10 * time.Second)

		counter.Add(10)
		require.True(t, time.Since(start) < 4*time.Second)
	}()

	s.Wakeup()

	time.Sleep(2 * time.Second)

	s.Wakeup()

}

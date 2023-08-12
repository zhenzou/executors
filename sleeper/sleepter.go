package sleeper

import (
	"time"
)

// Sleeper not concurrency safe
type Sleeper interface {
	// Wakeup wakeup the sleeper
	Wakeup()

	// Sleep sleep duration time unless wakeup
	// NOTE: can not sleep multiple times
	Sleep(duration time.Duration)
}

func NewSleeper() Sleeper {
	return &sleeper{
		ch: make(chan struct{}, 1),
	}
}

type sleeper struct {
	ch chan struct{}
}

func (s *sleeper) Wakeup() {
	// do nothing if no waiting goroutine
	select {
	case s.ch <- struct{}{}:
	default:
	}
}

func (s *sleeper) Sleep(duration time.Duration) {
	select {
	case <-time.After(duration):
		// timed out
	case <-s.ch:
		// Wait returned
	}
}

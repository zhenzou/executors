package executors

import (
	"context"
	"time"

	gxtime "github.com/dubbogo/timer"
)

func NewPoolScheduleExecutor(opts ..._PoolExecutorOption) ScheduledExecutor {
	return &PoolScheduleExecutor{
		PoolExecutor: _NewPoolExecutorService[any](opts...),
		tw:           gxtime.NewTimerWheel(),
	}
}

type PoolScheduleExecutor struct {
	*PoolExecutor[any]
	tw *gxtime.TimerWheel
}

func (p *PoolScheduleExecutor) Schedule(r Runnable, delay time.Duration) (CancelFunc, error) {
	timer := p.tw.AfterFunc(delay, func() {
		err := p.PoolExecutor.Execute(r)
		if err != nil {
			p.opts.ErrorHandler.CatchException(r, err)
		}
	})
	return timer.Stop, nil
}

func (p *PoolScheduleExecutor) ScheduleAtFixRate(r Runnable, period time.Duration) (CancelFunc, error) {
	ticker := p.tw.TickFunc(period, func() {
		err := p.PoolExecutor.Execute(r)
		if err != nil {
			p.opts.ErrorHandler.CatchException(r, err)
		}
	})
	return ticker.Stop, nil
}

func (p *PoolScheduleExecutor) Shutdown(ctx context.Context) error {
	defer func() {
		p.tw.Close()
	}()

	return p.PoolExecutor.Shutdown(ctx)
}

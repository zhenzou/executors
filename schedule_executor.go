package executors

import (
	"context"
	"sync"
	"time"

	"github.com/aptible/supercronic/cronexpr"
	gxtime "github.com/dubbogo/timer"

	"github.com/zhenzou/executors/cron"
	"github.com/zhenzou/executors/routine"
)

func NewPoolScheduleExecutor(opts ..._PoolExecutorOption) ScheduledExecutor {
	executor := _NewPoolExecutorService[any](opts...)
	return &PoolScheduleExecutor{
		PoolExecutor: executor,
		tw:           gxtime.NewTimerWheel(),
		dispatcher:   cron.NewDispatcher[Runnable](executor.opts.Logger),
	}
}

type PoolScheduleExecutor struct {
	*PoolExecutor[any]
	tw               *gxtime.TimerWheel
	cronScheduleOnce sync.Once
	dispatcher       cron.Dispatcher[Runnable]
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

func (p *PoolScheduleExecutor) ScheduleAtCronRate(r Runnable, rule CRONRule) (CancelFunc, error) {
	expr, err := cronexpr.ParseStrict(rule.Expr)
	if err != nil {
		return nil, ErrInvalidCronExpr
	}
	location, err := time.LoadLocation(rule.Timezone)
	if err != nil {
		return nil, ErrInvalidCronTimezone
	}

	removeFunc := p.dispatcher.AddTask(r, expr, location)

	p.cronScheduleOnce.Do(p.dispatchCRON)

	return removeFunc, nil
}

func (p *PoolScheduleExecutor) dispatchCRON() {
	routine.GoWithRecovery(p.opts.Logger, func() {
		ch := p.dispatcher.GetReadyTask()
		for r := range ch {
			err := p.Execute(r)
			if err != nil {
				p.opts.ErrorHandler.CatchException(r, err)
			}
		}
	}, p.dispatchCRON)
}

func (p *PoolScheduleExecutor) Shutdown(ctx context.Context) error {
	defer func() {
		p.tw.Close()
	}()

	return p.PoolExecutor.Shutdown(ctx)
}

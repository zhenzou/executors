package executors

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/aptible/supercronic/cronexpr"
	gxtime "github.com/dubbogo/timer"

	"github.com/zhenzou/executors/cron"
	"github.com/zhenzou/executors/routine"
)

func NewPoolScheduleExecutor(opts ..._PoolExecutorOption) ScheduledExecutor {
	executor := internalNewPoolExecutorService[any](opts...)
	scheduleExecutor := PoolScheduleExecutor{
		PoolExecutor: executor,
		dispatcher:   cron.NewDispatcher[Runnable](executor.opts.Logger),
	}
	scheduleExecutor.initTimerWheelOnce = sync.OnceFunc(scheduleExecutor.initTimerWheel)
	return &scheduleExecutor
}

type PoolScheduleExecutor struct {
	*PoolExecutor[any]
	tw                 *gxtime.TimerWheel
	initTimerWheelOnce func()
	cronScheduleOnce   sync.Once
	dispatcher         cron.Dispatcher[Runnable]
}

func (p *PoolScheduleExecutor) initTimerWheel() {
	p.tw = gxtime.NewTimerWheel()
}

func (p *PoolScheduleExecutor) Schedule(r Runnable, delay time.Duration) (CancelFunc, error) {
	p.opts.Logger.Debug("start to schedule new task", slog.Duration("delay", delay))

	p.initTimerWheelOnce()

	timer := p.tw.AfterFunc(delay, func() {
		err := p.PoolExecutor.Execute(r)
		if err != nil {
			if errors.Is(err, ErrShutdown) {
				return
			}
			p.opts.ErrorHandler.CatchError(r, err)
		}
	})
	return timer.Stop, nil
}

func (p *PoolScheduleExecutor) ScheduleFunc(fn func(ctx context.Context), delay time.Duration) (CancelFunc, error) {
	return p.Schedule(RunnableFunc(fn), delay)
}

func (p *PoolScheduleExecutor) ScheduleAtFixRate(r Runnable, period time.Duration) (CancelFunc, error) {
	p.opts.Logger.Debug("start to schedule new task at fix rate", slog.Duration("period", period))

	p.initTimerWheelOnce()

	ticker := p.tw.TickFunc(period, func() {
		p.opts.Logger.Debug("start to execute task at fix rate", slog.Duration("period", period))

		err := p.PoolExecutor.Execute(r)
		if err != nil {
			if errors.Is(err, ErrShutdown) {
				return
			}
			p.opts.ErrorHandler.CatchError(r, err)
		}
	})
	return ticker.Stop, nil
}

func (p *PoolScheduleExecutor) ScheduleFuncAtFixRate(fn func(ctx context.Context), delay time.Duration) (CancelFunc, error) {
	return p.ScheduleAtFixRate(RunnableFunc(fn), delay)
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

	p.opts.Logger.Debug("start to schedule new task at cron rate", slog.Any("rule", rule))

	removeFunc := p.dispatcher.AddTask(r, expr, location)

	p.cronScheduleOnce.Do(p.dispatchCRON)

	return removeFunc, nil
}

func (p *PoolScheduleExecutor) ScheduleFuncAtCronRate(fn func(ctx context.Context), rule CRONRule) (CancelFunc, error) {
	return p.ScheduleAtCronRate(RunnableFunc(fn), rule)
}

func (p *PoolScheduleExecutor) dispatchCRON() {
	routine.GoWithRecovery(p.opts.Logger, func() {
		p.opts.Logger.Debug("start to dispatch cron tasks")
		ch := p.dispatcher.GetReadyTask()
		for r := range ch {
			p.opts.Logger.Debug("start to execute cron task")

			err := p.Execute(r)
			if err != nil {
				if errors.Is(err, ErrShutdown) {
					return
				}
				p.opts.Logger.Debug("fail to execute cron task")

				p.opts.ErrorHandler.CatchError(r, err)
			}
		}
	}, p.dispatchCRON)
}

func (p *PoolScheduleExecutor) Shutdown(ctx context.Context) error {
	defer func() {
		if p.tw != nil {
			// wakeup tw
			p.tw.Tick(1 * time.Millisecond)
			p.tw.Close()
		}
	}()

	defer func() {
		p.dispatcher.Shutdown()
	}()

	return p.PoolExecutor.Shutdown(ctx)
}

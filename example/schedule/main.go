package main

import (
	"context"
	"time"

	"github.com/zhenzou/executors"
)

func main() {

	executor := executors.NewPoolScheduleExecutor(executors.WithMaxConcurrent(10))

	callable := executors.RunnableFunc(func(ctx context.Context) {
		println(time.Now().Format(time.RFC3339))
	})

	_, err := executor.ScheduleAtCronRate(callable, executors.CRONRule{
		Expr:     "*/1 * * * * * *",
		Timezone: "Local",
	})
	if err != nil {
		panic(err)
	}

	time.Sleep(24 * time.Hour)
}

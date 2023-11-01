package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/zhenzou/executors"
)

func main() {

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	executor := executors.NewPoolScheduleExecutor(
		executors.WithMaxConcurrent(10),
		executors.WithLogger(logger))

	callable1 := executors.RunnableFunc(func(ctx context.Context) {
		println("1:", time.Now().Format(time.RFC3339))
	})

	callable2 := executors.RunnableFunc(func(ctx context.Context) {
		println("2:", time.Now().Format(time.RFC3339))
	})

	callable3 := executors.RunnableFunc(func(ctx context.Context) {
		println("3:", time.Now().Format(time.RFC3339))
	})

	_, err := executor.ScheduleAtCronRate(callable1, executors.CRONRule{
		// Expr: "0/1 * * * *",
		// Expr: "*/1 * * * *",
		Expr:     "*/5 * * * * * *",
		Timezone: "Local",
	})
	if err != nil {
		panic(err)
	}

	_, err = executor.ScheduleAtCronRate(callable2, executors.CRONRule{
		// Expr: "0 4 * * *",
		// Expr: "*/1 * * * *",
		Expr:     "*/5 * * * * * *",
		Timezone: "Local",
	})
	if err != nil {
		panic(err)
	}

	_, err = executor.ScheduleAtCronRate(callable3, executors.CRONRule{
		// Expr: "0 4 * * *",
		// Expr: "*/1 * * * *",
		Expr:     "*/5 * * * * * *",
		Timezone: "Local",
	})
	if err != nil {
		panic(err)
	}

	time.Sleep(24 * time.Hour)
}

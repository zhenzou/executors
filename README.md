## executors

A package like Java ThreadPoolExecutor for Go based on generic.

To bring better async task processing experience.

[![Go Reference](https://pkg.go.dev/badge/image)](https://pkg.go.dev/github.com/zhenzou/executors)
[![Go Report](https://goreportcard.com/badge/github.com/zhenzou/executors)](https://goreportcard.com/report/github.com/zhenzou/executors)
![Go Lint](https://github.com/zhenzou/executors/actions/workflows/lint.yml/badge.svg)
![Go Test](https://github.com/zhenzou/executors/actions/workflows/test.yml/badge.svg)
[![Go Coverage](https://github.com/zhenzou/executors/wiki/coverage.svg)](https://github.com/zhenzou/executors/wiki/Test-coverage-report)


https://github.com/zhenzou/executors/actions/workflows/test/badge.svg

## Example

```go
package main

import (
	"context"
	"time"

	"github.com/zhenzou/executors"
)

type Person struct {
	Name string
}

func main() {

	executor := executors.NewPoolExecutorService[Person](executors.WithMaxConcurrent(10))

	callable := executors.CallableFunc[Person](func(ctx context.Context) (Person, error) {
		time.Sleep(1 * time.Second)
		return Person{
			Name: "future",
		}, nil
	})

	f1, err := executor.Submit(callable)
	if err != nil {
		panic(err)
	}
	// get, block until async task completed
	got, err := f1.Get(context.Background())
	if err != nil {
		panic(err)
	}
	println(got.Name)

	f2, _ := executor.Submit(callable)
	// then, add callback when call succeed
	f2.Then(func(val Person) {
		println(val.Name)
	})

	f3, _ := executor.Submit(callable)
	// catch, add callback when call failed
	f3.Catch(func(err error) {
		println(err.Error())
	})
}

```


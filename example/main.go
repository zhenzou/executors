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

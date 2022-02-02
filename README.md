## executors

A package like Java ThreadPoolExecutor for Go based on generic.

To bring better async task processing experience.

## Example

```go

package main

import (
	"context"
)

type Person struct {
	Name string
}

func main() {

	service := NewPoolExecutorService[Person](WithMaxConcurrent(10))

	callable := CallableFunc[Person](func(ctx context.Context) (Person, error) {
		return Person{
			Name: "future",
		}, nil
	})
	f, err := service.Submit(callable)

	if err != nil {
		panic(err)
	}

	for i := 0; i < 100; i++ {
		got, err := f.Get(context.Background())
		if err != nil {
			panic(err)
		}

		println(got.Name)
	}
}


```


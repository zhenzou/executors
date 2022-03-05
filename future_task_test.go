package executors

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFutureTask_Then(t *testing.T) {
	service := NewPoolExecutorService[Person](WithMaxConcurrent(10))

	thenName1 := ""
	ch1 := make(chan struct{})

	thenName2 := ""
	ch2 := make(chan struct{})

	t.Run("success then", func(t *testing.T) {
		callable := CallableFunc[Person](func(ctx context.Context) (Person, error) {
			return Person{
				Name: "future1",
			}, nil
		})
		f, err := service.Submit(callable)

		require.NoError(t, err)

		f.Then(func(val Person) {
			thenName1 = val.Name
			ch1 <- struct{}{}
		})
	})

	t.Run("error then", func(t *testing.T) {
		callable := CallableFunc[Person](func(ctx context.Context) (Person, error) {
			if true {
				ch2 <- struct{}{}
				return Person{}, errors.New("text")
			}
			return Person{
				Name: "future2",
			}, nil
		})
		f, err := service.Submit(callable)

		require.NoError(t, err)

		f.Then(func(val Person) {
			thenName2 = val.Name
			ch2 <- struct{}{}
		})
	})

	<-ch1
	<-ch2

	require.Equal(t, "future1", thenName1)
	require.Equal(t, "", thenName2)
}

func TestFutureTask_Catch(t *testing.T) {
	service := NewPoolExecutorService[Person](WithMaxConcurrent(10))

	thenName1 := ""
	ch1 := make(chan struct{})

	thenName2 := ""
	ch2 := make(chan struct{})

	t.Run("success then", func(t *testing.T) {
		callable := CallableFunc[Person](func(ctx context.Context) (Person, error) {
			return Person{
				Name: "future1",
			}, nil
		})
		f, err := service.Submit(callable)

		require.NoError(t, err)

		f.Then(func(val Person) {
			ch1 <- struct{}{}
		})

		f.Catch(func(err error) {
			thenName1 = "future1"
			ch1 <- struct{}{}
		})
	})

	t.Run("error then", func(t *testing.T) {
		callable := CallableFunc[Person](func(ctx context.Context) (Person, error) {
			if true {
				return Person{}, errors.New("future2")
			}
			return Person{
				Name: "future2",
			}, nil
		})
		f, err := service.Submit(callable)

		require.NoError(t, err)

		f.Catch(func(err error) {
			thenName2 = err.Error()
			ch2 <- struct{}{}
		})
	})

	<-ch1
	<-ch2

	require.Equal(t, "", thenName1)
	require.Equal(t, "future2", thenName2)
}

package executors

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFutureTask_Then(t *testing.T) {
	service := NewPoolExecutorService[Person](WithMaxConcurrent(10))

	name1 := ""
	ch1 := make(chan struct{})

	name2 := ""
	ch2 := make(chan struct{})

	name3 := ""
	ch3 := make(chan struct{})

	t.Run("success then", func(t *testing.T) {
		callable := CallableFunc[Person](func(ctx context.Context) (Person, error) {
			return Person{
				Name: "future1",
			}, nil
		})
		f, err := service.Submit(callable)

		require.NoError(t, err)

		f.Then(func(val Person) error {
			name1 = val.Name
			ch1 <- struct{}{}
			return nil
		})
	})

	t.Run("catch call error", func(t *testing.T) {
		callable := CallableFunc[Person](func(ctx context.Context) (Person, error) {
			if true {
				return Person{}, errors.New("call error")
			}
			return Person{
				Name: "future2",
			}, nil
		})
		f, err := service.Submit(callable)

		require.NoError(t, err)

		f.Then(func(val Person) error {
			name2 = "test"
			ch2 <- struct{}{}
			return err
		}).Catch(func(err error) {
			name2 = err.Error()
			ch2 <- struct{}{}
		})
	})

	t.Run("catch then error", func(t *testing.T) {
		callable := CallableFunc[Person](func(ctx context.Context) (Person, error) {
			return Person{
				Name: "future2",
			}, nil
		})
		f, err := service.Submit(callable)

		require.NoError(t, err)

		f.Then(func(val Person) error {
			return fmt.Errorf("then error %s", val.Name)
		}).Catch(func(err error) {
			name3 = err.Error()
			ch3 <- struct{}{}
		})
	})

	<-ch1
	<-ch2
	<-ch3

	require.Equal(t, "future1", name1)
	require.Equal(t, "call error", name2)
	require.Equal(t, "then error future2", name3)
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

		f.Then(func(val Person) error {
			ch1 <- struct{}{}
			return nil
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

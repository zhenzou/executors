package executors

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestPoolExecutor_Submit(t *testing.T) {
	type Person struct {
		Name string
	}
	service := NewPoolExecutorService[Person](WithMaxConcurrent(10))

	t.Run("one task success", func(t *testing.T) {
		callable := CallableFunc[Person](func(ctx context.Context) (Person, error) {
			return Person{
				Name: "future",
			}, nil
		})
		f, err := service.Submit(callable)

		require.NoError(t, err)

		var wg sync.WaitGroup

		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				got, err := f.Get(context.Background())
				require.NoError(t, err)
				require.Equal(t, "future", got.Name)
			}()
		}
		wg.Wait()
	})

	t.Run("one task error", func(t *testing.T) {
		targetErr := errors.New("error")
		callable := CallableFunc[Person](func(ctx context.Context) (Person, error) {
			return Person{}, targetErr
		})
		f, err := service.Submit(callable)

		require.NoError(t, err)

		var wg sync.WaitGroup

		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, err := f.Get(context.Background())
				require.True(t, errors.Is(err, targetErr))
			}()
		}
		wg.Wait()
	})

	t.Run("one task canceled", func(t *testing.T) {
		callable := CallableFunc[Person](func(ctx context.Context) (Person, error) {
			time.Sleep(1 * time.Second)
			return Person{
				Name: "future",
			}, nil
		})
		f, err := service.Submit(callable)
		require.NoError(t, err)

		_ = time.AfterFunc(50*time.Millisecond, func() {
			cancel := f.Cancel()
			require.True(t, cancel)
		})

		var wg sync.WaitGroup

		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, err := f.Get(context.Background())
				require.True(t, errors.Is(err, ErrFutureCanceled))
			}()
		}
		wg.Wait()

	})
}

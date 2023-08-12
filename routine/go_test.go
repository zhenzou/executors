package routine

import (
	"errors"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWithRecovery(t *testing.T) {
	t.Run("run f", func(t *testing.T) {
		counter := 1
		t.Cleanup(func() {
			require.Equal(t, 2, counter)
		})

		WithRecovery(slog.Default(), func() {
			counter++
		}, nil)
	})

	t.Run("run cleanup", func(t *testing.T) {
		counter := 1
		t.Cleanup(func() {
			require.Equal(t, 3, counter)
		})

		WithRecovery(slog.Default(), func() {
			panic(errors.New("test"))
		}, func() {
			counter += 2
		})
	})
}

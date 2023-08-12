package routine

import (
	"log/slog"
)

// WithRecovery run fn with recovery
// will log error and cleanup if error happened
func WithRecovery(logger *slog.Logger, fn, cleanup func()) {
	defer func() {
		if cause := recover(); cause != nil {
			logError(logger, cause)
			cleanup()
		}
	}()

	fn()
}

// GoWithRecovery go run fn with recovery
// will log error and cleanup if error happened
func GoWithRecovery(logger *slog.Logger, fn, cleanup func()) {
	go WithRecovery(logger, fn, cleanup)
}

func logError(logger *slog.Logger, cause any) {
	logger.Error("panic happened, will restart loop", slog.Any("cause", cause))
}

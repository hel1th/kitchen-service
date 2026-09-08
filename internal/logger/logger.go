package logger

import (
	"log/slog"
	"os"
	"strings"
)

// Setup creates and configures a new slog.Logger based on the given level
func Setup(level string) *slog.Logger {
	var l slog.Level
	switch strings.ToLower(level) {
	case "debug":
		l = slog.LevelDebug
	case "info":
		l = slog.LevelInfo
	case "warn":
		l = slog.LevelWarn
	case "error":
		l = slog.LevelError
	default:
		l = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: l,
	}

	handler := slog.NewJSONHandler(os.Stdout, opts)
	logger := slog.New(handler)

	// Set it as default for the standard log package (if it uses slog)
	slog.SetDefault(logger)

	return logger
}

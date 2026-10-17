package logger

import (
	"log/slog"
	"os"
	"strings"
)

func NewLogger() *slog.Logger {
	format := strings.ToLower(os.Getenv("LOG_FORMAT"))
	levelStr := strings.ToLower(os.Getenv("LOG_LEVEL"))

	var level slog.Level
	switch levelStr {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: level,
	}

	var handler slog.Handler
	if format == "json" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	return slog.New(handler)
}

// Global helper for simple initialization
var Log = NewLogger()

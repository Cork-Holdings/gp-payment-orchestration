package utils

import (
	"context"
	"io"
	"log/slog"
	"os"
)

var logger *slog.Logger

func InitLogger(logFilePath string) error {

	file, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)

	if err != nil {
		return err
	}

	// w := io.MultiWriter(os.Stderr, file)

	w := io.Writer(file)
	handlerOptions := &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}

	logger = slog.New(slog.NewJSONHandler(w, handlerOptions))

	return nil
}

func Log(level slog.Level, msg string, args ...interface{}) {

	ctx := context.Background()

	if logger != nil {
		logger.Log(ctx, level, msg, args...)
	} else {
		// Fallback to stderr if logger is not initialized
		slog.Log(ctx, level, msg, args...)
	}
}

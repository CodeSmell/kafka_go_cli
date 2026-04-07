package logging

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
)

func New(levelName string) (*slog.Logger, error) {
	level, err := parseLevel(levelName)
	if err != nil {
		return nil, err
	}

	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	})
	return slog.New(handler), nil
}

func parseLevel(levelName string) (slog.Level, error) {
	switch strings.ToLower(levelName) {
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn", "warning":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return 0, fmt.Errorf("invalid log level %q (use debug, info, warn, error)", levelName)
	}
}

package app

import (
	"log/slog"
	"os"
)

func setLogger(level string) *slog.Logger {
	var log *slog.Logger
	switch level {
	case "local":
		log = slog.New(
			slog.NewTextHandler(
				os.Stdout, &slog.HandlerOptions{
					Level: slog.LevelDebug,
				},
			),
		)
	default:
		log = slog.New(
			slog.NewJSONHandler(
				os.Stdout, &slog.HandlerOptions{
					Level: slog.LevelInfo,
				},
			),
		)
	}

	return log
}

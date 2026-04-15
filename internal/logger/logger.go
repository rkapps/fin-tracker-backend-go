package logger

import (
	"log/slog"
	"os"
)

func SetLogger() {

	programLevel := new(slog.LevelVar)

	switch os.Getenv("LOG_LEVEL") {
	case "DEBUG":
		programLevel.Set(slog.LevelDebug)
	case "WARN":
		programLevel.Set(slog.LevelWarn)
	case "ERROR":
		programLevel.Set(slog.LevelError)
	default:
		programLevel.Set(slog.LevelInfo)
	}

	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: programLevel,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.LevelKey {
				a.Key = "severity"
			}
			if a.Key == slog.MessageKey {
				a.Key = "message"
			}
			return a
		},
	})

	slog.SetDefault(slog.New(handler))
}

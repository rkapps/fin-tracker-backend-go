package logger

import (
	"log/slog"
	"os"
)

func SetLogger() {

	programLevel := new(slog.LevelVar)

	// 2. Set it to Debug immediately
	programLevel.Set(slog.LevelDebug)

	// 3. IMPORTANT: Pass the LevelVar to the handler
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: programLevel,
	})

	// 4. IMPORTANT: Set this logger as the global default
	logger := slog.New(handler)
	slog.SetDefault(logger)
}

package logger

import (
	"context"
	"log/slog"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	charmlog "github.com/charmbracelet/log"
)

const (
	LevelTrace = slog.Level(-8)  // below Debug (-4)
	LevelDebug = slog.LevelDebug // -4
	LevelInfo  = slog.LevelInfo  //  0
	LevelWarn  = slog.LevelWarn  //  4
	LevelError = slog.LevelError //  8
)

// Config holds the log level configuration per component.
// Parsed from LOG_LEVELS env var — e.g. "portfolio=debug,pipeline=info,storage=warn"
type Config struct {
	levels       map[string]slog.Level
	defaultLevel slog.Level
}

// func (c *Config) WithContext(param any, plog *Logger) {
// 	panic("unimplemented")
// }

// NewLogger parses the LOG_LEVELS env var and returns a Config.
func New() *Config {
	c := &Config{
		levels:       make(map[string]slog.Level),
		defaultLevel: slog.LevelInfo,
	}

	env := strings.TrimSpace(os.Getenv("LOG_LEVELS"))
	if env == "" {
		return c
	}

	for _, part := range strings.Split(env, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		kv := strings.SplitN(part, "=", 2)

		if len(kv) == 1 {
			// no "=" found — treat as global default level
			// LOG_LEVELS=trace
			c.defaultLevel = parseLevel(kv[0])
			continue
		}

		// component=level format
		// LOG_LEVELS=pipeline=trace
		component := strings.TrimSpace(kv[0])
		level := parseLevel(kv[1])
		c.levels[component] = level
	}

	return c
}

func (c *Config) newLogger(component string, level slog.Level) *Logger {
	handler := newHandler(level)
	return &Logger{
		slog.New(handler).With("component", component),
	}
}

func (c *Config) For(component string) *Logger {
	// exact match first
	if level, ok := c.levels[component]; ok {
		return c.newLogger(component, level)
	}

	// prefix match — "processor" matches "processor.acquisition"
	parts := strings.Split(component, ".")
	for i := len(parts) - 1; i > 0; i-- {
		prefix := strings.Join(parts[:i], ".")
		if level, ok := c.levels[prefix]; ok {
			return c.newLogger(component, level)
		}
	}

	return c.newLogger(component, c.defaultLevel)

}

// Default returns a logger with the default level and no component.
func (c *Config) Default() *Logger {
	return &Logger{slog.New(newHandler(c.defaultLevel))}
}

// newHandler builds the correct handler based on environment.
// JSON for production (Cloud Run), Text for local development.
func newHandler(level slog.Level) slog.Handler {
	if isLocal() {
		charm := charmlog.NewWithOptions(os.Stdout, charmlog.Options{
			Level:           charmlog.Level(level),
			ReportTimestamp: true,
			TimeFormat:      "15:04:05",
			ReportCaller:    false,
		})

		styles := charmlog.DefaultStyles()

		styles.Levels[charmlog.InfoLevel] = lipgloss.NewStyle().
			SetString("INFO ").Bold(true).Foreground(lipgloss.Color("46")) // ← trailing space

		styles.Levels[charmlog.Level(LevelTrace)] = lipgloss.NewStyle().
			SetString("TRACE").
			Bold(true).
			Foreground(lipgloss.Color("99"))

		// explicitly set DEBUG so we know what it is
		styles.Levels[charmlog.DebugLevel] = lipgloss.NewStyle().
			SetString("DEBUG").
			Bold(true).
			Foreground(lipgloss.Color("39")) // cyan

		styles.Levels[charmlog.ErrorLevel] = lipgloss.NewStyle().
			SetString("ERROR").
			Bold(true).
			Foreground(lipgloss.Color("200")) // cyan

		charm.SetStyles(styles)

		// wrap with fixed width formatter
		fixed := newFixedWidthHandler(charm,
			15, // key width   — "component      "
			20, // value width — "portfolio           "
		)

		return &traceHandler{Handler: fixed, level: level}
		// // wrap with traceHandler to bypass slog's level filtering
		// return &traceHandler{
		// 	Handler: handler,
		// 	level:   level,
		// }
	}

	return slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	})

}

// isLocal checks if we are running in a local environment.
func isLocal() bool {
	env := os.Getenv("ENV")
	return env == "" || env == "local" || env == "development"
}
func parseLevel(level string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "trace":
		return LevelTrace // -8
	case "debug":
		return slog.LevelDebug // -4
	case "info":
		return slog.LevelInfo // 0
	case "warn":
		return slog.LevelWarn // 4
	case "error":
		return slog.LevelError // 8
	default:
		return slog.LevelInfo
	}
}

// Logger wraps slog.Logger and adds custom levels.
type Logger struct {
	*slog.Logger
}

// Trace logs at TRACE level.
func (l *Logger) Trace(msg string, args ...any) {
	// fmt.Printf("Trace called, handler enabled: %v level: %d\n",
	// 	l.Handler().Enabled(context.Background(), LevelTrace),
	// 	LevelTrace,
	// )
	l.Log(context.Background(), LevelTrace, msg, args...)
}

// TraceCtx logs at TRACE level with context.
func (l *Logger) TraceCtx(ctx context.Context, msg string, args ...any) {
	l.Log(ctx, LevelTrace, msg, args...)
}

// With returns a new Logger with pre-attached fields.
// Overrides slog.Logger.With to return our custom type.
func (l *Logger) With(args ...any) *Logger {
	return &Logger{l.Logger.With(args...)}
}

// traceHandler wraps a slog.Handler and enables levels below Debug.
type traceHandler struct {
	slog.Handler
	level slog.Level
}

// Enabled overrides slog's default level check.
func (h *traceHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level
}

// WithAttrs returns a new traceHandler with the given attributes.
func (h *traceHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &traceHandler{
		Handler: h.Handler.WithAttrs(attrs),
		level:   h.level,
	}
}

// WithGroup returns a new traceHandler with the given group.
func (h *traceHandler) WithGroup(name string) slog.Handler {
	return &traceHandler{
		Handler: h.Handler.WithGroup(name),
		level:   h.level,
	}
}

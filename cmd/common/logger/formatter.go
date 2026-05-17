package logger

import (
	"context"
	"fmt"
	"log/slog"
	"runtime"
	"strings"
)

type fixedWidthHandler struct {
	slog.Handler
	keyWidth  int
	valWidth  int
	component string       // ← store component here
	posWidths []fieldWidth // ← width per position

}

type fieldWidth struct {
	key int
	val int
}

func newFixedWidthHandler(inner slog.Handler, keyWidth, valWidth int) *fixedWidthHandler {
	return &fixedWidthHandler{
		Handler:  inner,
		keyWidth: keyWidth,
		valWidth: valWidth,
		posWidths: []fieldWidth{
			{key: 20, val: 50}, // 1st arg — e.g. UID, AccountID (long values)
			{key: 10, val: 20}, // 2nd arg — e.g. Symbol, Type (short values)
			{key: 6, val: 20},  // 3rd arg — e.g. Qty, Amount
			{key: 6, val: 10},  // 4th arg
		},
	}
}

func (h *fixedWidthHandler) Handle(ctx context.Context, r slog.Record) error {
	_, file, line, ok := runtime.Caller(6)
	caller := ""
	if ok {
		c := fmt.Sprintf("%s:%d", file, line)
		if len(c) > 30 {
			c = "..." + c[len(c)-27:]
		}
		caller = fmt.Sprintf("%-30s", c)
	}
	// component is stored on the handler itself via With()
	// access it from the pre-stored field
	// component := fmt.Sprintf("%-25s", h.component)
	// msg := fmt.Sprintf("%s %s %-25s", caller, component, r.Message)
	// newRecord := slog.NewRecord(r.Time, r.Level, msg, r.PC)

	pos := 0
	var sb strings.Builder

	r.Attrs(func(a slog.Attr) bool {
		if a.Key == "component" {
			return true
		}
		// get width for this position
		kw := h.keyWidth
		vw := h.valWidth
		if pos < len(h.posWidths) {
			kw = h.posWidths[pos].key
			vw = h.posWidths[pos].val
		}

		// paddedKey := fmt.Sprintf("%-*s", kw, a.Key)
		// paddedVal := fmt.Sprintf("%-*s", vw, strings.Trim(a.Value.String(), `"`))
		val := fmt.Sprintf("%v", a.Value.Any())

		paddedKey := fmt.Sprintf("%-*s", kw, truncate(a.Key, kw)) // ← truncate key
		paddedVal := fmt.Sprintf("%-*s", vw, truncate(val, vw))   // ← truncate val

		sb.WriteString(paddedKey)
		sb.WriteString(" ")
		sb.WriteString(paddedVal)
		sb.WriteString(" ")

		pos++
		return true
	})

	// build full message — caller + component + message + attrs
	fullMsg := fmt.Sprintf("%s %-25s %-20s %s",
		caller,
		h.component,
		r.Message,
		sb.String(),
	)

	// pass to charmbracelet as a single message with NO attrs
	newRecord := slog.NewRecord(r.Time, r.Level, fullMsg, r.PC)
	// no AddAttrs — charmbracelet renders nothing extra

	return h.Handler.Handle(ctx, newRecord)
}

func (h *fixedWidthHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	component := h.component

	// filter out component before passing to inner handler
	filtered := make([]slog.Attr, 0, len(attrs))
	for _, a := range attrs {
		if a.Key == "component" {
			component = a.Value.String()
			continue // ← don't pass to inner handler
		}
		filtered = append(filtered, a)
	}

	return &fixedWidthHandler{
		Handler:   h.Handler.WithAttrs(filtered), // ← filtered, no component
		keyWidth:  h.keyWidth,
		valWidth:  h.valWidth,
		component: component,
		posWidths: h.posWidths, // ← copy
	}
}

func (h *fixedWidthHandler) WithGroup(name string) slog.Handler {
	return &fixedWidthHandler{
		Handler:   h.Handler.WithGroup(name),
		keyWidth:  h.keyWidth,
		valWidth:  h.valWidth,
		posWidths: h.posWidths, // ← copy
	}
}

// truncate string to max length, add "..." if truncated
func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	if max <= 3 {
		return s[:max]
	}
	return s[:max-3] + "..."
}

// rawValue is a custom type that renders without quotes
type rawValue string

func (r rawValue) String() string {
	return string(r)
}

package logger

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"
)

var pacific *time.Location

func init() {
	var err error
	pacific, err = time.LoadLocation("America/Los_Angeles")
	if err != nil {
		pacific = time.FixedZone("PST", -8*3600)
	}
}

func dualTime(t time.Time) string {
	utc := t.UTC().Format(time.RFC3339)
	local := t.In(pacific).Format("2006-01-02T15:04:05 MST")
	return fmt.Sprintf("%s (%s)", utc, local)
}

func handlerOptions(level slog.Level) *slog.HandlerOptions {
	return &slog.HandlerOptions{
		Level:     level,
		AddSource: true,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey && len(groups) == 0 {
				return slog.String("time", dualTime(a.Value.Time()))
			}
			return a
		},
	}
}

// fanout sends each record to every handler that accepts its level
type fanout struct{ handlers []slog.Handler }

func (f fanout) Enabled(ctx context.Context, level slog.Level) bool {
	for _, h := range f.handlers {
		if h.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (f fanout) Handle(ctx context.Context, r slog.Record) error {
	for _, h := range f.handlers {
		if h.Enabled(ctx, r.Level) {
			_ = h.Handle(ctx, r.Clone())
		}
	}
	return nil
}

func (f fanout) WithAttrs(attrs []slog.Attr) slog.Handler {
	out := make([]slog.Handler, len(f.handlers))
	for i, h := range f.handlers {
		out[i] = h.WithAttrs(attrs)
	}
	return fanout{out}
}

func (f fanout) WithGroup(name string) slog.Handler {
	out := make([]slog.Handler, len(f.handlers))
	for i, h := range f.handlers {
		out[i] = h.WithGroup(name)
	}
	return fanout{out}
}

func Setup(logPath string) error {
	if err := os.MkdirAll(filepath.Dir(logPath), 0o755); err != nil {
		return err
	}

	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}

	console := slog.NewTextHandler(os.Stderr, handlerOptions(slog.LevelInfo))
	errorsFile := slog.NewJSONHandler(file, handlerOptions(slog.LevelInfo))

	logger := slog.New(fanout{[]slog.Handler{console, errorsFile}})
	slog.SetDefault(logger)
	return nil
}

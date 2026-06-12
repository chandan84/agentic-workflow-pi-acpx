// Package obs provides logging, trace/span context helpers, and OTel init stubs.
//
// Logging is slog JSON with run/trace/span correlation. OTel init is a stub
// that records the configured endpoint and returns a no-op shutdown — real
// otel SDK wiring lands in a follow-up iteration once the dep is approved.
package obs

import (
	"context"
	"io"
	"log/slog"
	"os"
)

// Logger constructs a configured slog.Logger.
func Logger(service, level, format string) *slog.Logger {
	var lvl slog.Level
	if err := lvl.UnmarshalText([]byte(level)); err != nil {
		lvl = slog.LevelInfo
	}
	var h slog.Handler
	w := io.Writer(os.Stdout)
	opts := &slog.HandlerOptions{Level: lvl}
	if format == "text" {
		h = slog.NewTextHandler(w, opts)
	} else {
		h = slog.NewJSONHandler(w, opts)
	}
	return slog.New(h).With("service", service)
}

type ctxKey struct{ k string }

var (
	keyRun   = ctxKey{"run"}
	keyTrace = ctxKey{"trace"}
	keySpan  = ctxKey{"span"}
)

// WithRunID stamps a run id on the context for log correlation.
func WithRunID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, keyRun, id)
}

// FromContext returns a logger whose attributes are pre-stamped with the
// run/trace/span ids found in ctx.
func FromContext(ctx context.Context, base *slog.Logger) *slog.Logger {
	l := base
	if v, ok := ctx.Value(keyRun).(string); ok && v != "" {
		l = l.With("run_id", v)
	}
	if v, ok := ctx.Value(keyTrace).(string); ok && v != "" {
		l = l.With("trace_id", v)
	}
	if v, ok := ctx.Value(keySpan).(string); ok && v != "" {
		l = l.With("span_id", v)
	}
	return l
}

// InitTracing is a placeholder that returns a no-op shutdown. Once we adopt
// the OTel SDK, this will configure the OTLP exporter and tracer provider.
func InitTracing(_ context.Context, service, endpoint string) (shutdown func(context.Context) error, err error) {
	_ = service
	_ = endpoint
	return func(context.Context) error { return nil }, nil
}

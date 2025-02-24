package logger

import (
	"context"
	"fmt"
	"io"
	"log"
	"log/slog"
	"path/filepath"
	"runtime"
	"time"
)

// represent a func to return the trace id from specified context
type TraceIDFn func(cxt context.Context) string

type Logger struct {
	handler   slog.Handler
	TraceIDFn TraceIDFn
}

// New constructs a new log for app use
func New(w io.Writer, minLevel Level, serviceName string, traceIDFn TraceIDFn) *Logger {
	return new(w, minLevel, serviceName, traceIDFn, Events{})
}

// NewWithEvents constructs a new log for application use with events.

func NewWithEvents(w io.Writer, minLevel Level, serviceName string, traceIDFn TraceIDFn, events Events) *Logger {
	return new(w, minLevel, serviceName, traceIDFn, events)
}

func NewWithHandler(h slog.Handler) *Logger {
	return &Logger{
		handler: h,
	}
}

func NewStdLogger(logger *Logger, level Level) *log.Logger {
	return slog.NewLogLogger(logger.handler, slog.Level(level))
}

func new(w io.Writer, minLevel Level, serviceName string, traceIDFn TraceIDFn, events Events) *Logger {
	f := func(groups []string, attributes slog.Attr) slog.Attr {
		if attributes.Key == slog.SourceKey {
			if source, ok := attributes.Value.Any().(*slog.Source); ok {
				pathLine := fmt.Sprintf("%s:%d", filepath.Base(source.File), source.Line)
				return slog.Attr{Key: "file", Value: slog.StringValue(pathLine)}
			}
		}
		return attributes
	}

	handler := slog.Handler(slog.NewJSONHandler(w, &slog.HandlerOptions{AddSource: true, Level: slog.Level(minLevel), ReplaceAttr: f}))
	// If events are to be processed, wrap the JSON handler around the custom
	// log handler.
	if events.Debug != nil || events.Error != nil || events.Info != nil || events.Warn != nil {
		handler = newLogHandler(handler, events)
	}

	// we need to add additional attributes with each log
	otherAttributes := []slog.Attr{
		{Key: "service", Value: slog.StringValue(serviceName)},
	}
	handler = handler.WithAttrs(otherAttributes)
	return &Logger{
		handler:   handler,
		TraceIDFn: traceIDFn,
	}
}

// ////////////
func (log *Logger) write(cxt context.Context, level Level, caller int, msg string, args ...any) {
	slogLevel := slog.Level(level)
	if !log.handler.Enabled(cxt, slogLevel) {
		return
	}
	var pcs [1]uintptr
	runtime.Callers(caller, pcs[:])

	r := slog.NewRecord(time.Now(), slogLevel, msg, pcs[0])

	if log.TraceIDFn != nil {
		args = append(args, "trace_id", log.TraceIDFn(cxt))
	}
	r.Add(args...)
	log.handler.Handle(cxt, r)
}
func (log *Logger) Debug(cxt context.Context, msg string, args ...any) {
	log.write(cxt, LevelDebug, 3, msg, args...)
}

// Debugc logs the information at the specified call stack position.
func (log *Logger) Debugc(cxt context.Context, msg string, caller int, args ...any) {
	log.write(cxt, LevelDebug, caller, msg, args...)
}

// Info logs at LevelInfo with the given context.
func (log *Logger) Info(ctx context.Context, msg string, args ...any) {
	log.write(ctx, LevelInfo, 3, msg, args...)
}

// Infoc logs the information at the specified call stack position.
func (log *Logger) Infoc(ctx context.Context, caller int, msg string, args ...any) {
	log.write(ctx, LevelInfo, caller, msg, args...)
}

// Warn logs at LevelWarn with the given context.
func (log *Logger) Warn(ctx context.Context, msg string, args ...any) {
	log.write(ctx, LevelWarn, 3, msg, args...)
}

// Warnc logs the information at the specified call stack position.
func (log *Logger) Warnc(ctx context.Context, caller int, msg string, args ...any) {
	log.write(ctx, LevelWarn, caller, msg, args...)
}

// Error logs at LevelError with the given context.
func (log *Logger) Error(ctx context.Context, msg string, args ...any) {
	log.write(ctx, LevelError, 3, msg, args...)
}

// Errorc logs the information at the specified call stack position.
func (log *Logger) Errorc(ctx context.Context, caller int, msg string, args ...any) {
	log.write(ctx, LevelError, caller, msg, args...)
}

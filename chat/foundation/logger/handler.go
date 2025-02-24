package logger

import (
	"context"
	"log/slog"
)

// is a wrapper around slog handler to capture which log level is being logged for event handling
type logHandler struct {
	handler slog.Handler
	events  Events
}

func newLogHandler(handler slog.Handler, events Events) *logHandler {
	return &logHandler{
		handler: handler,
		events:  events,
	}
}

// because the handler igornoe record whose level is lower
// with this function we can enabled logging record at certain level
func (h *logHandler) Enabled(cxt context.Context, level slog.Level) bool {
	return h.handler.Enabled(cxt, level)
}

// WithAttrs returns a new Handler whose attributes consist of both the receiver's attributes and the arguments.
// The Handler owns the slice: it may retain, modify or discard it.
func (h *logHandler) WithAttrs(attrs []slog.Attr) slog.Handler {

	return &logHandler{
		handler: h.handler.WithAttrs(attrs),
		events:  h.events,
	}
}

func (h *logHandler) WithGroup(name string) slog.Handler {
	return &logHandler{
		handler: h.handler.WithGroup(name),
		events:  h.events,
	}
}

func (h *logHandler) Handle(cxt context.Context, r slog.Record) error {
	switch r.Level {
	case slog.LevelDebug:
		if h.events.Debug != nil {
			h.events.Debug(cxt, toRecord(r))
		}

	case slog.LevelError:
		if h.events.Error != nil {
			h.events.Error(cxt, toRecord(r))
		}

	case slog.LevelInfo:
		if h.events.Info != nil {
			h.events.Info(cxt, toRecord(r))
		}

	case slog.LevelWarn:
		if h.events.Warn != nil {
			h.events.Warn(cxt, toRecord(r))
		}

	}
	return h.handler.Handle(cxt, r)

}

package logger

import (
	"context"
	"log/slog"
	"time"
)

type Level slog.Level

// log levels
const (
	LevelDebug = Level(slog.LevelDebug)
	LevelInfo  = Level(slog.LevelInfo)
	LevelWarn  = Level(slog.LevelWarn)
	LevelError = Level(slog.LevelError)
)

// record represent data to be logged
type Record struct {
	Time       time.Time
	Message    string
	Level      Level
	Attributes map[string]any
}

// function to be excuted against log level
type EventFn func(cxt context.Context, r Record)

// event func to log level
type Events struct {
	Debug EventFn
	Info  EventFn
	Warn  EventFn
	Error EventFn
}

// convert default slog record to my Record struct
func toRecord(r slog.Record) Record {
	attributes := make(map[string]any, r.NumAttrs())
	fun := func(attr slog.Attr) bool {
		attributes[attr.Key] = attr.Value.Any()
		return true
	}
	r.Attrs(fun)
	return Record{
		Time:       r.Time,
		Message:    r.Message,
		Level:      Level(r.Level),
		Attributes: attributes,
	}
}

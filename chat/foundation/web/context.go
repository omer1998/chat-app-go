package web

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

const (
	writerKey = iota + 1
	traceIdKey
)

// set the writer of the reques
func SetWriter(cxt context.Context, w http.ResponseWriter) context.Context {
	return context.WithValue(cxt, writerKey, w)
}

// get the writer of the reques
func GetWriter(cxt context.Context) http.ResponseWriter {
	v, ok := cxt.Value(writerKey).(http.ResponseWriter)
	if !ok {
		return nil
	}
	return v
}

func SetTraceId(cxt context.Context, id uuid.UUID) context.Context {
	return context.WithValue(cxt, traceIdKey, id)
}

// GetTraceID returns the traceID for the request.
func GetTraceId(cxt context.Context) uuid.UUID {
	id, ok := cxt.Value(traceIdKey).(uuid.UUID)
	if !ok {
		return uuid.UUID{}
	}
	return id
}

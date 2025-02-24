package web

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"
)

type Encoder interface {
	Encode() (data []byte, contentType string, err error)
}

// represent a function that handle a http request
type HandlerFunc func(cxt context.Context, r *http.Request) Encoder

// represent function to add information to the logs
type Logger func(cxt context.Context, msg string, args ...any)

type App struct {
	log    Logger
	mux    *http.ServeMux
	mw     []MidFunc
	origin []string
}

func NewApp(log Logger, mw ...MidFunc) *App {

	return &App{
		log: log,
		mw:  mw,
		mux: http.NewServeMux(),
	}
}

// ServeHTTP implements the http.Handler interface. It's the entry point for
// all http traffic and allows the opentelemetry mux to run first to handle
// tracing. The opentelemetry mux then calls the application mux to handle
// application traffic. This was set up in the NewApp function.
func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.mux.ServeHTTP(w, r)
}

func (a *App) corsHandler(webHandler HandlerFunc) HandlerFunc {
	h := func(cxt context.Context, r *http.Request) Encoder {
		reqOrigin := r.Header.Get("Origin")
		w := GetWriter(cxt) // response writer
		for _, origin := range a.origin {
			if origin == "*" || origin == reqOrigin {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				break
			}
		}
		w.Header().Set("Access-Control-Allow-Methods", "POST, PATCH, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		w.Header().Set("Access-Control-Max-Age", "86400")
		return webHandler(cxt, r)
	}
	return h
}

// enable cors

func (a *App) EnableCORS(origin []string) {
	a.origin = origin

	handler := func(cxt context.Context, r *http.Request) Encoder {
		return nil
	}
	handler = WrapMiddleWare([]MidFunc{a.corsHandler}, handler)
}

// HandlerFuncNoMid sets a handler function for a given HTTP method and path
// pair to the application server mux. Does not include the application
// middleware or OTEL tracing.

func (a *App) HandlerFuncNoMid(method string, group string, path string, handerFunc HandlerFunc) {

	// the final point to reach is to handle the request on specific path
	// mean we need to reach this--> a.mux.Handle(path, func(w http.RepsonseWriter, r *http.Reques){
	// })
	// 1 so after we recieve the incoming request we need to save the response writer of request in the context
	h := func(w http.ResponseWriter, r *http.Request) {
		cxt := SetWriter(r.Context(), w)
		// cxt  set traceid

		resp := handerFunc(cxt, r)
		if err := Respond(cxt, w, resp); err != nil {
			a.log(cxt, "web-respond", "ERROR", err)
		}

	}
	finalPath := path
	if group != "" {
		finalPath = "/" + group + path
	}
	finalPath = fmt.Sprintf("%s %s", method, finalPath)
	a.mux.HandleFunc(finalPath, h)
}

func (a *App) HandlerFunc(method string, group string, path string, handlerFunc HandlerFunc, mw ...MidFunc) {
	handlerFunc = WrapMiddleWare(a.mw, handlerFunc)
	handlerFunc = WrapMiddleWare(mw, handlerFunc)

	if a.origin != nil {
		handlerFunc = WrapMiddleWare([]MidFunc{a.corsHandler}, handlerFunc)
	}
	h := func(w http.ResponseWriter, r *http.Request) {
		cxt := r.Context()
		cxt = SetWriter(cxt, w)
		cxt = SetTraceId(cxt, uuid.New())

		resp := handlerFunc(cxt, r)

		if err := Respond(cxt, w, resp); err != nil {
			// a.log(cxt, "Error - response", "err", err.Error())
			a.log(cxt, "web-respond", "ERROR", err)
			return
		}

	}

	finalPath := path
	if group != "" {
		finalPath = "/" + group + path
	}

	finalPath = fmt.Sprintf("%s %s", method, finalPath)
	a.mux.HandleFunc(finalPath, h)
}

func (a *App) RawHandlerFunc(method string, group string, path string, rawHandlerFunc http.HandlerFunc, mw ...MidFunc) {
	handlerFunc := func(cxt context.Context, r *http.Request) Encoder {
		r = r.WithContext(cxt)
		rawHandlerFunc(GetWriter(cxt), r)
		return nil
	}
	handlerFunc = WrapMiddleWare(a.mw, handlerFunc)
	handlerFunc = WrapMiddleWare(mw, handlerFunc)

	if a.origin != nil {
		handlerFunc = WrapMiddleWare([]MidFunc{a.corsHandler}, handlerFunc)
	}
	h := func(w http.ResponseWriter, r *http.Request) {
		cxt := r.Context()
		cxt = SetWriter(cxt, w)
		cxt = SetTraceId(cxt, uuid.New())

		resp := handlerFunc(cxt, r)

		if err := Respond(cxt, w, resp); err != nil {
			// a.log(cxt, "Error - response", "err", err.Error())
			a.log(cxt, "web-respond", "ERROR", err)
			return
		}

	}

	finalPath := path
	if group != "" {
		finalPath = "/" + group + path
	}

	finalPath = fmt.Sprintf("%s %s", method, finalPath)
	a.mux.HandleFunc(finalPath, h)
}

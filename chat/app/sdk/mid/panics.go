package mid

import (
	"context"
	"net/http"
	"runtime/debug"

	"github.com/omer1998/chat-app-go.git/chat/app/sdk/errs"
	"github.com/omer1998/chat-app-go.git/chat/foundation/web"
)

// Panic recover from panic and convert panic to error so it will be handled in error

func Panics() web.MidFunc {
	midFunc := func(next web.HandlerFunc) web.HandlerFunc {

		handleFunc := func(cxt context.Context, r *http.Request) (resp web.Encoder) {

			defer func() {
				// stops the panicking sequence by restoring normal execution and
				// retrieves the error value passed to the call of panic.
				if rec := recover(); rec != nil {
					trace := debug.Stack()
					resp = errs.Newf(errs.InternalOnlyLog, "PANIC [%v] Trace [%s] ", rec, trace)
				}
			}()

			return next(cxt, r)

		}
		return handleFunc
	}
	return midFunc
}

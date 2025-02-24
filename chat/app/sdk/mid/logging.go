package mid

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/omer1998/chat-app-go.git/chat/app/sdk/errs"
	"github.com/omer1998/chat-app-go.git/chat/foundation/logger"
	"github.com/omer1998/chat-app-go.git/chat/foundation/web"
)

// this logging middleware

func Logger(log *logger.Logger) web.MidFunc {
	midFunc := func(next web.HandlerFunc) web.HandlerFunc {
		h := func(cxt context.Context, r *http.Request) web.Encoder {
			start := time.Now()

			path := r.URL.Path
			if r.URL.RawQuery != "" { // to look if there is query parameters in the url
				path = fmt.Sprintf("%s?%s", path, r.URL.RawQuery)
			}
			log.Info(cxt, "request started", "method", r.Method, "path", path, "remote address", r.RemoteAddr)
			resp := next(cxt, r)
			err := IsError(resp)

			statusCode := errs.OK

			if err != nil { // mean there is actually an error
				statusCode = errs.Internal

				var appErr *errs.Error
				if errors.As(err, &appErr) {
					statusCode = appErr.Code
				}
			}

			log.Info(cxt, "request complete", "method", r.Method, "path", path, "remode address", r.RemoteAddr,
				"status code", statusCode, "since", time.Since(start).String())
			return resp
		}
		return h
	}
	return midFunc
}

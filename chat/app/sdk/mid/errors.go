// middle ware func to handle the error comming from calls
package mid

import (
	"context"
	"errors"
	"net/http"

	"github.com/omer1998/chat-app-go.git/chat/app/sdk/errs"
	"github.com/omer1998/chat-app-go.git/chat/foundation/logger"
	"github.com/omer1998/chat-app-go.git/chat/foundation/web"
)

func Errors(log *logger.Logger) web.MidFunc {
	midFunc := func(next web.HandlerFunc) web.HandlerFunc {
		h := func(cxt context.Context, r *http.Request) web.Encoder {
			resp := next(cxt, r)
			err := IsError(resp)
			if err == nil {
				return resp
			}
			var appErr *errs.Error
			if !errors.As(err, &appErr) {
				appErr = errs.Newf(errs.Internal, "Internal Server Error")
			}

			log.Error(cxt, "Handle error during request",
				"error", err,
				"source-file-err", appErr.FileName,
				"source-func-name", appErr.FuncName,
			)
			if appErr.Code == errs.InternalOnlyLog {
				appErr = errs.Newf(errs.Internal, "Internal Server Error")
			}

			return appErr

		}
		return h

	}
	return midFunc

}

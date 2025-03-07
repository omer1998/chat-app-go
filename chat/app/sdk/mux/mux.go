package mux

import (
	"context"
	"net/http"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/omer1998/chat-app-go.git/chat/app/domain/chatapp"
	"github.com/omer1998/chat-app-go.git/chat/app/sdk/mid"
	"github.com/omer1998/chat-app-go.git/chat/foundation/logger"
	"github.com/omer1998/chat-app-go.git/chat/foundation/web"
)

type Config struct {
	Log     *logger.Logger
	Js      jetstream.JetStream
	Subject string
	Stream  jetstream.Stream
	Api     *chatapp.App
}

func WebAPI(config Config) http.Handler {

	logger := func(cxt context.Context, msg string, args ...any) {
		config.Log.Info(cxt, msg, args...)
	}
	app := web.NewApp(logger,
		mid.Logger(config.Log),
		mid.Errors(config.Log),
		mid.Panics())

	// here we need to connect the routes of my app in order to be known to the mux and therefore become able to
	// perform task according to the path patterns (routes will be in the app/ domain level / chatapp.go)
	chatapp.Routes(app, config.Log, config.Js, config.Subject, config.Stream, config.Api)
	return app // this app is actually a http.Handler because it implement Handler interface
	//type Handler interface {
	// ServeHTTP(ResponseWriter, *Request)
	// }
}

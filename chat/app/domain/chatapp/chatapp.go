package chatapp

import (
	"context"
	"net/http"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/omer1998/chat-app-go.git/chat/app/sdk/chat"
	"github.com/omer1998/chat-app-go.git/chat/app/sdk/chat/users"
	"github.com/omer1998/chat-app-go.git/chat/app/sdk/errs"
	"github.com/omer1998/chat-app-go.git/chat/foundation/logger"
	"github.com/omer1998/chat-app-go.git/chat/foundation/web"
)

// remember these packages at the domain level of the app
// are responsible of recieving external input --> validating it  --> calling it to the bussiness layer --> and formulating a response or error to the middleware

type App struct {
	log  *logger.Logger
	chat *chat.Chat
}

func NewApp(log *logger.Logger, js jetstream.JetStream, subject string, stream jetstream.Stream) (*App, error) {
	users := users.NewUsers(log)
	chat, err := chat.NewChat(log, users, js, stream, subject)
	if err != nil {
		return nil, err
	}
	return &App{log: log, chat: chat}, nil
}
func (a App) Test(cxt context.Context, r *http.Request) web.Encoder {
	// w := web.GetWriter(cxt)
	// w.Write([]byte("hello from test"))
	return web.NewNoResponse()
}
func (a App) connect(cxt context.Context, r *http.Request) web.Encoder {
	// fmt.Println("request", r)

	user, err := a.chat.Handshake(cxt, web.GetWriter(cxt), r)
	if err != nil {
		return errs.Newf(errs.Internal, "error handshake %s: ", err.Error())
	}

	// now we need to listen for message on this connection
	// and direct sending message accordingly
	// cxtWithCancel, cancel := context.WithCancel(context.Background())
	// defer cancel()
	a.chat.Listen(cxt, user)
	a.chat.Listen(cxt, user)

	return web.NewNoResponse()
}

// func WriteMessage(){}

func Routes(app *web.App, log *logger.Logger, js jetstream.JetStream, subject string, stream jetstream.Stream, api *App) {
	app.HandlerFunc(http.MethodGet, "", "/test", api.Test)
	app.HandlerFunc(http.MethodGet, "", "/connect", api.connect)
}

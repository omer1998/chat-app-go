package chatapp

import (
	"context"
	"net/http"

	"github.com/omer1998/chat-app-go.git/chat/foundation/web"
)

// remember these packages at the domain level of the app
// are responsible of recieving external input --> validating it  --> calling it to the bussiness layer --> and formulating a response or error to the middleware

type app struct {
}

func NewApp() *app {
	return &app{}
}
func (a app) Test(cxt context.Context, r *http.Request) web.Encoder {
	// w := web.GetWriter(cxt)
	// w.Write([]byte("hello from test"))
	return web.NewNoResponse()
}
func Routes(app *web.App) {
	api := NewApp()
	app.HandlerFunc(http.MethodGet, "", "/", api.Test)

}

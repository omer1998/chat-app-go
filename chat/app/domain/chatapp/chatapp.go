package chatapp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/omer1998/chat-app-go.git/chat/app/sdk/errs"
	"github.com/omer1998/chat-app-go.git/chat/foundation/logger"
	"github.com/omer1998/chat-app-go.git/chat/foundation/web"
)

// remember these packages at the domain level of the app
// are responsible of recieving external input --> validating it  --> calling it to the bussiness layer --> and formulating a response or error to the middleware

type app struct {
	log *logger.Logger
}

func NewApp(log *logger.Logger) *app {
	return &app{log: log}
}
func (a app) Test(cxt context.Context, r *http.Request) web.Encoder {
	// w := web.GetWriter(cxt)
	// w.Write([]byte("hello from test"))
	return web.NewNoResponse()
}
func (a app) connect(cxt context.Context, r *http.Request) web.Encoder {
	// fmt.Println("request", r)
	upgrader := websocket.Upgrader{
		// CheckOrigin: func(r *http.Request) bool {
		// 	return true
		// },
	}

	conn, err := upgrader.Upgrade(web.GetWriter(cxt), r, nil)
	if err != nil {
		return errs.Newf(errs.Internal, "error upgrading server to websocket %s", err.Error())
	}
	user, err := a.handshake(conn)
	if err != nil {
		return errs.Newf(errs.Internal, "error handshake %s: ", err.Error())
	}

	a.log.Info(cxt, "handshake complete", "user", user)

	return web.NewNoResponse()
}

func (a app) handshake(conn *websocket.Conn) (User, error) {
	// after connection established we send HELLO to the client, after client recieve the HELLO we expect to hear from him id and name
	err := conn.WriteMessage(websocket.TextMessage, []byte("HELLO"))
	if err != nil {
		return User{}, errs.Newf(errs.Internal, "error write %s ", err.Error())
	}
	cxt := context.Background()
	cxt, cancel := context.WithTimeout(cxt, time.Second*10)
	defer cancel()

	msg, err := a.readMessages(cxt, conn)
	if err != nil {
		return User{}, errs.Newf(errs.Internal, "error reading message %s ", err.Error())
	}
	// here we expect the msg to be the id and name (which the User type)
	// the msg is of byte type we need to marshal it to the User type
	var user User
	err = json.Unmarshal(msg, &user)
	if err != nil {
		return User{}, errs.Newf(errs.Internal, "error unmarshal data %s: ", err.Error())
	}

	// after we recieve the user
	// we need to send WELCOME user.name
	helloName := fmt.Sprintf("WELCOME %s ", user.Name)
	err = conn.WriteMessage(websocket.TextMessage, []byte(helloName))
	if err != nil {
		return User{}, errs.Newf(errs.Internal, "error write %s ", err.Error())
	}
	return user, nil
}
func (a app) readMessages(cxt context.Context, conn *websocket.Conn) ([]byte, error) {
	type Response struct {
		data []byte
		err  error
	}
	respChan := make(chan Response)

	go func() {
		a.log.Info(cxt, "starting handshake read")
		defer a.log.Info(cxt, "end handshake read")
		_, data, err := conn.ReadMessage()
		if err != nil {
			respChan <- Response{err: err, data: nil}
		}
		respChan <- Response{
			data: data,
			err:  nil,
		}
	}()

	select {
	case msg := <-respChan:
		return msg.data, msg.err
	case <-cxt.Done():
		conn.Close()
		return nil, cxt.Err()
	}

}

// func WriteMessage(){}

func Routes(app *web.App, log *logger.Logger) {
	api := NewApp(log)
	app.HandlerFunc(http.MethodGet, "", "/test", api.Test)
	app.HandlerFunc(http.MethodGet, "", "/connect", api.connect)
}

package chat

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/omer1998/chat-app-go.git/chat/app/sdk/errs"
	"github.com/omer1998/chat-app-go.git/chat/foundation/logger"
	"github.com/omer1998/chat-app-go.git/chat/foundation/web"
)

var (
	ErrUserNotExist = fmt.Errorf("user not exist")
	ErrUserExist    = fmt.Errorf("user already exist")
)

type Users interface {
	AddUser(user User) error
	RemoveUser(cxt context.Context, id uuid.UUID) error
	RetrieveUser(id uuid.UUID) (User, error)
	Connections() map[uuid.UUID]Connection
	UpdateLastPingTime(usrId uuid.UUID) error
	UpdateLastPongTime(usrId uuid.UUID) error
}

type Chat struct {
	log      *logger.Logger
	subject  string
	js       jetstream.JetStream
	users    Users
	stream   jetstream.Stream
	Consumer jetstream.Consumer
}

func NewChat(log *logger.Logger, users Users, js jetstream.JetStream, stream jetstream.Stream, subject string) (*Chat, error) {

	// here we need to create a stream
	cons, err := stream.Consumer(context.Background(), "omerconsumer")
	if err != nil {
		return nil, err
	}
	cht := &Chat{
		log:      log,
		users:    users,
		subject:  subject,
		js:       js,
		stream:   stream,
		Consumer: cons,
	}

	maxWait := time.Second * 10
	cht.ping(maxWait)
	return cht, nil
}

func (cht *Chat) Handshake(cxt context.Context, w http.ResponseWriter, r *http.Request) (User, error) {
	cht.log.Info(cxt, "handshake", "status", "started")

	// after connection established we send HELLO to the client, after client recieve the HELLO we expect to hear from him id and name
	upgrader := websocket.Upgrader{
		// CheckOrigin: func(r *http.Request) bool {
		// 	return true
		// },
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		cht.log.Info(cxt, "handshake", "error upgrading connection", "error", err)

		return User{}, errs.Newf(errs.Internal, "error upgrading server to websocket %s", err.Error())
	}
	err = conn.WriteMessage(websocket.TextMessage, []byte("HELLO"))
	if err != nil {
		cht.log.Info(cxt, "handshake", "error writing hello from server", "error", err)

		return User{}, fmt.Errorf("error write %s", err.Error())
	}

	cxt, cancel := context.WithTimeout(cxt, time.Second*10)
	defer cancel()

	usr := User{
		Conn:     conn,
		LastPing: time.Now(),
		LastPong: time.Now(),
	}

	msg, err := cht.readMessages(cxt, usr)
	if err != nil {
		cht.log.Info(cxt, "handshake", "error reading user data from client", "error", err)

		return User{}, fmt.Errorf("error reading message %s ", err.Error())
	}
	// here we expect the msg to be the id and name (which the User type)
	// the msg is of byte type we need to marshal it to the User type

	err = json.Unmarshal(msg, &usr)
	if err != nil {
		cht.log.Info(cxt, "handshake", "error unmarshal user data", "error", err)

		return User{}, fmt.Errorf("error unmarshal data %s: ", err.Error())
	}

	// after we recieve the user
	// add this user to our map
	if err = cht.users.AddUser(usr); err != nil {
		err = conn.WriteMessage(websocket.TextMessage, []byte("already connected"))
		defer conn.Close()
		if err != nil {
			cht.log.Info(cxt, "handshake", "error adding user", "error", err)

			return User{}, fmt.Errorf("error writing: %w", err)
		}

		return User{}, fmt.Errorf("error adding user: %w", err)
	}
	// we need to send WELCOME user.name
	helloName := fmt.Sprintf("WELCOME %s \n id %s", usr.Name, usr.Id)
	err = conn.WriteMessage(websocket.TextMessage, []byte(helloName))
	if err != nil {
		cht.log.Info(cxt, "handshake", "error sending msg", "error", err)

		return User{}, fmt.Errorf("error write %s ", err.Error())
	}
	//here we set the pong handler all we do is telling the connection when you recieve pong from this connection
	// update the last pong time

	cht.log.Info(cxt, "handshake", "status", "complete", "user", usr)
	// here we set the pong handler
	// in which we specifiy what we need to do when we recieve a pong from the client
	// in our situation we update the lastPongTime
	conn.SetPongHandler(cht.pong(cxt, usr.Id))
	return usr, nil
}
func (cht *Chat) Listen(cxt context.Context, from User) {

	for {

		data, err := cht.readMessages(cxt, from) //in readmessgae we already handle the logic of removing user if connection closed
		// here we only need to decide if close error we need to return else continue
		if err != nil {
			if cht.isCriticalError(cxt, err) {
				return
			}
			continue

		}

		cht.log.Info(cxt, "message recieved", "id", from.Id)

		var msg InMessage
		err = json.Unmarshal(data, &msg)
		if err != nil {
			cht.log.Info(cxt, "listen msg unmarshal", "status", "failed", "err", err)
			continue
		}

		to, err := cht.users.RetrieveUser(msg.ToId)
		if err != nil {
			if errors.Is(err, ErrUserNotExist) {
				msgBus := InMessageBus{
					ToId: msg.ToId,
					From: from,
					Msg:  msg.Msg,
				}
				cht.publishMessage(cxt, msgBus)
				cht.log.Info(cxt, "message send to bus", "from", from.Id, "id", to.Id)

				continue
			}
			return
		}

		if err := cht.sendMessage(from, to, msg); err != nil {
			cht.log.Info(cxt, "listen send message ", "status", "failed", "err", err)

		}
		cht.log.Info(cxt, "message send", "from", from.Id, "id", to.Id)

	}

}

// publishMessage will publish messgae to the stream
func (cht *Chat) publishMessage(cxt context.Context, inMsg InMessageBus) error {
	// here we need to push this message on the nats strem
	cht.log.Info(cxt, "publish message bus", "status", "start")

	data, err := json.Marshal(inMsg)
	if err != nil {

		return fmt.Errorf("publish msg bus, error marshal %w", err)
	}
	ack, err := cht.js.Publish(cxt, cht.subject, data)
	if err != nil {
		return fmt.Errorf("error publishing message: %w", err)
	}
	cht.log.Info(cxt, "publish message bus", "status", "success", "stream", ack.Stream, "sequence", ack.Sequence)

	return nil
}

// listenBus listen for incomming messages and direct them to users
// i think it is app level not connection level
func (cht *Chat) ListenBus(cxt context.Context, consumer jetstream.Consumer) {
	cht.log.Info(cxt, "listen bus", "status", "started")
	defer cht.log.Info(cxt, "listen bus", "status", "completed")

	go func() {
		for {
			// we want to get messages from the stream
			// Next is used to retrieve the next message from the consumer.
			// This method will block until the message is retrieved or timeout is reached.

			msg, err := consumer.Next()
			if err != nil {
				cht.log.Info(cxt, "listen bus consume", "error", err.Error())
				continue
			}
			if err = msg.Ack(); err != nil {
				cht.log.Info(cxt, "listen bus ack", "error", err.Error())
				continue
			}
			// unmarshal msg
			var msgBus InMessageBus
			err = json.Unmarshal(msg.Data(), &msgBus)
			if err != nil {
				cht.log.Info(cxt, "listen bus unmarshal", "error", err.Error())

				continue
			}

			if err = cht.sendMessageFromBus(msgBus); err != nil {
				cht.log.Info(cxt, "listen bus send message", "error", err.Error())
				continue
			}

		}
	}()

}
func (cht *Chat) sendMessageFromBus(message InMessageBus) error {

	// we need to write message to the connection of the to user
	msg := OutMessage{
		From: message.From,
		Msg:  message.Msg,
	}
	if err := message.From.Conn.WriteJSON(msg); err != nil {
		// here also if we can't send message
		// mostprobably the connection is problematic or may be closed
		return fmt.Errorf("error send msg to user: %w ", err)
	}
	return nil

}

func (cht *Chat) sendMessage(from User, to User, message InMessage) error {

	// we need to write message to the connection of the to user
	msg := OutMessage{
		From: from,
		Msg:  message.Msg,
	}
	if err := to.Conn.WriteJSON(msg); err != nil {
		// here also if we can't send message
		// mostprobably the connection is problematic or may be closed
		return fmt.Errorf("error send msg to user: %w ", err)
	}
	return nil

}

//	func (cht *Chat) writeMessage(data []byte, conn *websocket.Conn, msgType int) error {
//		if err := conn.WriteMessage(msgType, data); err != nil {
//			return fmt.Errorf("error writing message: %w", err)
//		}
//		return nil
//	}
func (cht *Chat) readMessages(cxt context.Context, user User) ([]byte, error) {
	type Response struct {
		data []byte
		err  error
	}
	respChan := make(chan Response)

	go func() {
		cht.log.Info(cxt, "read message", "status", "start")
		defer cht.log.Info(cxt, "read message", "status", "complete")
		// 	How ReadMessage Works:
		// Blocking Call: ReadMessage is a blocking call, meaning it will wait until a message is received or an error occurs.
		// Message Type: It returns the type of the message (e.g., websocket.TextMessage, websocket.BinaryMessage, etc.).
		// Message Data: It returns the message data as a byte slice.
		// Error Handling: If an error occurs, it returns the error. This can include network errors, WebSocket protocol errors, or connection closure errors.
		if user.Conn == nil {
			cht.log.Info(cxt, "read message", "status", "failure", "error", "connection is nil")

		}
		_, data, errr := user.Conn.ReadMessage()
		if errr != nil {
			respChan <- Response{err: errr, data: nil}
		}
		respChan <- Response{
			data: data,
			err:  nil,
		}
	}()

	select {
	case msg := <-respChan:

		if msg.err != nil {
			cht.log.Info(cxt, "read message", "status", "client diconnected", "error", msg.err.Error())
			cht.users.RemoveUser(cxt, user.Id)
			return nil, msg.err
		}

		return msg.data, msg.err
	case <-cxt.Done():
		// when connection closed we need to remove this user
		cht.log.Info(cxt, "read message", "status", "client closed", "error", cxt.Err().Error())
		cht.users.RemoveUser(cxt, user.Id)
		return nil, cxt.Err()
	}

}

func (c *Chat) isCriticalError(ctx context.Context, err error) bool {
	switch err.(type) {
	case *websocket.CloseError:
		c.log.Info(ctx, "chat-isCriticalError", "status", "client disconnected - websocket close", "error", err)
		return true
	case *net.OpError:
		c.log.Info(ctx, "chat-isCriticalError", "status", "client disconnected - op error", "error", err)
		return true
	default:
		if errors.Is(err, context.Canceled) {
			c.log.Info(ctx, "chat-isCriticalError", "status", "client canceled")
			return true
		}

		c.log.Info(ctx, "chat-isCriticalError", "err", err)
		return false
	}

}

func (cht *Chat) pong(cxt context.Context, userId uuid.UUID) func(appData string) error {
	h := func(appData string) error {
		cht.log.Debug(cxt, " ** pong handler **", "status", "started")
		defer cht.log.Debug(cxt, " ** pong handler **", "status", "complete", "appData", appData)

		err := cht.users.UpdateLastPongTime(userId)
		if err != nil {
			cht.log.Info(cxt, "pong handler", "status", "failed", "error", err)
			return err
		}
		cht.log.Info(cxt, "pong handler", "status", "success", "msg", "last pong time updated successfuly")

		return nil
	}
	return h
}

// it will be an orfen go routine continously working to ping all connection in order to maintain only the
// available connecction
// failure to ping mean the connection is closed thus we need to remove this connenction from the users
// we will ping based on a new copy of our user:connection map
// this ping help us maintain only the available connection in our users
func (cht *Chat) ping(maxWait time.Duration) {

	ticker := time.NewTicker(maxWait)

	go func() {
		// the aim is pind
		ctx := web.SetTraceId(context.Background(), uuid.New())
		for {

			<-ticker.C // this is blocking code every 10 seconds this loop run
			cht.log.Debug(context.Background(), "Ping", "status", "started")

			for usrId, v := range cht.users.Connections() {

				// here we need to check for this connection; if the duration till the pong time excced certain duration
				// if so we need to remove this user
				sub := v.LastPong.Sub(v.LastPing)
				cht.log.Debug(context.Background(), "Ping", "ping/pong period", sub.String())

				if sub > maxWait {
					// here we need to remove the user
					cht.log.Info(ctx, "Ping", "status", "failed", "ping/pong period", sub.String())
					err := cht.users.RemoveUser(context.Background(), usrId)
					if err != nil {
						cht.log.Info(ctx, "Ping", "status", "failed", "error", err)
						return
					}
					continue
				}
				// // here we need to update the last ping time of this usrID in order to to update the connection
				// // in the next loop
				err := cht.users.UpdateLastPingTime(usrId)
				if err != nil {
					cht.log.Info(ctx, "Ping", "status", "failed", "error", err)
					return
				}

				// then send a ping message

				if err := v.Conn.WriteMessage(websocket.PingMessage, []byte("ping")); err != nil {

					cht.log.Info(ctx, "Ping", "status", "failed", "error", err)

				}
			}

			cht.log.Info(ctx, "Ping", "status", "completed", "connection length", len(cht.users.Connections()))

		}
	}()

}

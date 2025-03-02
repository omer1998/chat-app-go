package chat

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/omer1998/chat-app-go.git/chat/app/sdk/errs"
	"github.com/omer1998/chat-app-go.git/chat/foundation/logger"
)

type Chat struct {
	log   *logger.Logger
	mu    sync.RWMutex
	users map[uuid.UUID]User
}

func NewChat(log *logger.Logger) *Chat {

	cht := &Chat{
		log:   log,
		users: make(map[uuid.UUID]User),
	}
	cht.ping()
	return cht
}

func (cht *Chat) Handshake(cxt context.Context, w http.ResponseWriter, r *http.Request) (User, error) {
	// after connection established we send HELLO to the client, after client recieve the HELLO we expect to hear from him id and name
	upgrader := websocket.Upgrader{
		// CheckOrigin: func(r *http.Request) bool {
		// 	return true
		// },
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return User{}, errs.Newf(errs.Internal, "error upgrading server to websocket %s", err.Error())
	}
	err = conn.WriteMessage(websocket.TextMessage, []byte("HELLO"))
	if err != nil {
		return User{}, errs.Newf(errs.Internal, "error write %s ", err.Error())
	}

	cxt, cancel := context.WithTimeout(cxt, time.Second*10)
	defer cancel()

	usr := User{
		Conn: conn,
	}

	msg, err := cht.readMessages(cxt, usr)
	if err != nil {
		return User{}, errs.Newf(errs.Internal, "error reading message %s ", err.Error())
	}
	// here we expect the msg to be the id and name (which the User type)
	// the msg is of byte type we need to marshal it to the User type

	err = json.Unmarshal(msg, &usr)
	if err != nil {
		return User{}, errs.Newf(errs.Internal, "error unmarshal data %s: ", err.Error())
	}

	// after we recieve the user
	// add this user to our map
	if err = cht.addUser(usr); err != nil {
		err = conn.WriteMessage(websocket.TextMessage, []byte("already connected"))
		defer conn.Close()
		if err != nil {
			return User{}, fmt.Errorf("error writing: %w", err)
		}

		return User{}, fmt.Errorf("error adding user: %w", err)
	}
	// we need to send WELCOME user.name
	helloName := fmt.Sprintf("WELCOME %s ", usr.Name)
	err = conn.WriteMessage(websocket.TextMessage, []byte(helloName))
	if err != nil {
		return User{}, errs.Newf(errs.Internal, "error write %s ", err.Error())
	}

	cht.log.Info(cxt, "handshake complete", "user", usr)

	return usr, nil
}
func (cht *Chat) Listen(cxt context.Context, user User) {

	for {

		data, err := cht.readMessages(cxt, user) //in readmessgae we already handle the logic of removing user if connection closed
		// here we only need to decide if close error we need to return else continue
		if err != nil {
			if cht.isCriticalError(cxt, err) {
				return
			}
			continue

		}

		var msg InMessage
		err = json.Unmarshal(data, &msg)
		if err != nil {
			cht.log.Info(cxt, "listen msg unmarshal", "status", "failed", "err", err)
			continue
		}
		if err := cht.sendMessage(msg); err != nil {
			cht.log.Info(cxt, "listen send message ", "status", "failed", "err", err)

		}
	}

}

func (cht *Chat) sendMessage(message InMessage) error {
	cht.mu.RLock()
	defer cht.mu.RUnlock()

	to, exist := cht.users[message.ToId]
	if !exist {
		return fmt.Errorf("to user not exist")
	}
	// we need to write message to the connection of the to user
	msg := OutMessage{
		To:  User{Id: to.Id, Name: to.Name},
		Msg: message.Msg,
	}
	if err := to.Conn.WriteJSON(msg); err != nil {
		// here also if we can't send message
		// mostprobably the connection is problematic or may be closed
		return fmt.Errorf("error send msg to user: %w ", err)
	}
	return nil

}

func (cht *Chat) addUser(usr User) error {
	cht.mu.Lock()
	defer cht.mu.Unlock()
	_, exists := cht.users[usr.Id]
	if exists {
		return fmt.Errorf("user already exists")
	}

	cht.log.Info(context.Background(), "user added", "status", "success", "id", usr.Id, "name", usr.Name)
	cht.users[usr.Id] = usr
	return nil
}

func (cht *Chat) connections() map[uuid.UUID]User {
	newUsrs := make(map[uuid.UUID]User)
	for k, v := range cht.users {
		newUsrs[k] = v
	}
	return newUsrs
}

func (cht *Chat) removeUser(cxt context.Context, id uuid.UUID) {
	cht.mu.Lock()
	defer cht.mu.Unlock()
	v, exist := cht.users[id]
	if !exist {
		cht.log.Info(cxt, "remove user", "id", id, "status", "does not exist")
		return
	}
	v.Conn.Close()
	delete(cht.users, id)
	cht.log.Info(cxt, "remove user", "status", "success", "id", id, "name", v.Name)

}

// it will be an orfen go routine continously working to ping all connection in order to maintain only the
// available connecction
// failure to ping mean the connection is closed thus we need to remove this connenction from the users
// we will ping based on a new copy of our user:connection map
// this ping help us maintain only the available connection in our users
func (cht *Chat) ping() {
	ticker := time.NewTicker(time.Second * 10)

	go func() {
		// the aim is pind
		for {
			cht.log.Info(context.Background(), "Ping", "status", "started")

			<-ticker.C // this is blocking code every 10 seconds this loop run
			fmt.Println("users now ", cht.users, "length", len(cht.users))
			for _, v := range cht.connections() {
				if err := v.Conn.WriteMessage(websocket.PingMessage, []byte("ping")); err != nil {

					cht.log.Info(context.Background(), "Ping", "status", "failed", "error", err)

				}
			}

			cht.log.Info(context.Background(), "Ping", "status", "completed", "available connection length", len(cht.users))

		}
	}()

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
		cht.log.Info(cxt, "starting read")
		defer cht.log.Info(cxt, "end read")
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
			cht.removeUser(cxt, user.Id)
			return nil, msg.err
		}

		return msg.data, msg.err
	case <-cxt.Done():
		// when connection closed we need to remove this user
		cht.log.Info(cxt, "read message", "status", "client closed", "error", cxt.Err().Error())
		cht.removeUser(cxt, user.Id)
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

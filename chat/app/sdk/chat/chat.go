package chat

import (
	"context"
	"encoding/json"
	"fmt"
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
	users map[uuid.UUID]connection
}

func NewChat(log *logger.Logger) *Chat {

	cht := &Chat{
		log:   log,
		users: make(map[uuid.UUID]connection),
	}
	cht.ping()
	return cht
}

func (cht *Chat) Handshake(conn *websocket.Conn) error {
	// after connection established we send HELLO to the client, after client recieve the HELLO we expect to hear from him id and name
	err := conn.WriteMessage(websocket.TextMessage, []byte("HELLO"))
	if err != nil {
		return errs.Newf(errs.Internal, "error write %s ", err.Error())
	}
	cxt := context.Background()
	cxt, cancel := context.WithTimeout(cxt, time.Second*10)
	defer cancel()

	msg, err := cht.readMessages(cxt, conn)
	if err != nil {
		return errs.Newf(errs.Internal, "error reading message %s ", err.Error())
	}
	// here we expect the msg to be the id and name (which the User type)
	// the msg is of byte type we need to marshal it to the User type
	var usr user
	err = json.Unmarshal(msg, &usr)
	if err != nil {
		return errs.Newf(errs.Internal, "error unmarshal data %s: ", err.Error())
	}

	// after we recieve the user
	// add this user to our map
	if err = cht.addUser(usr, conn); err != nil {
		err = conn.WriteMessage(websocket.TextMessage, []byte("already connected"))
		defer conn.Close()
		if err != nil {
			return fmt.Errorf("error writing: %w", err)
		}

		return fmt.Errorf("error adding user: %w", err)
	}
	// we need to send WELCOME user.name
	helloName := fmt.Sprintf("WELCOME %s ", usr.Name)
	err = conn.WriteMessage(websocket.TextMessage, []byte(helloName))
	if err != nil {
		return errs.Newf(errs.Internal, "error write %s ", err.Error())
	}

	cht.log.Info(cxt, "handshake complete", "user", usr)

	return nil
}
func (cht *Chat) Listen(cxt context.Context, conn *websocket.Conn) {
	go func() {
		for {

			data, err := cht.readMessages(cxt, conn)
			if err != nil {
				cht.log.Info(cxt, "listen read msg", "status", "failed", "err", err)
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

	}()
}

func (cht *Chat) sendMessage(message InMessage) error {
	cht.mu.RLock()
	defer cht.mu.RUnlock()

	from, exist := cht.users[message.FromId]
	if !exist {
		return fmt.Errorf("from user not exist")
	}
	to, exist := cht.users[message.ToId]
	if !exist {
		return fmt.Errorf("to user not exist")
	}
	// we need to write message to the connection of the to user
	msg := OutMessage{
		From: user{Id: from.id, Name: from.name},
		To:   user{Id: to.id, Name: to.name},
		Msg:  message.Msg,
	}
	if err := to.conn.WriteJSON(msg); err != nil {
		return fmt.Errorf("error send msg to user: %w ", err)
	}
	return nil

}

func (cht *Chat) addUser(usr user, cn *websocket.Conn) error {
	cht.mu.Lock()
	defer cht.mu.Unlock()
	_, exists := cht.users[usr.Id]
	if exists {
		return fmt.Errorf("user already exists")
	}
	connect := connection{
		id:   usr.Id,
		name: usr.Name,
		conn: cn,
	}
	cht.log.Info(context.Background(), "user added", "status", "success", "id", connect.id, "user name", connect.name)
	cht.users[usr.Id] = connect
	return nil
}

func (cht *Chat) connections() map[uuid.UUID]connection {
	newUsrs := make(map[uuid.UUID]connection)
	for k, v := range cht.users {
		newUsrs[k] = v
	}
	return newUsrs
}

func (cht *Chat) removeUser(id uuid.UUID) {
	cht.mu.Lock()
	defer cht.mu.Unlock()
	v, exist := cht.users[id]
	if !exist {
		cht.log.Info(context.Background(), "remove user", "id", id, "status", "does not exist")
		return
	}
	v.conn.Close()
	delete(cht.users, id)
	cht.log.Info(context.Background(), "remove user", "status", "does success", "id", id, "name", v.name)

}

// it will be an orfen go routine continously working to ping all connection in order to maintain only the
// available connecction
// failure to ping mean the connection is closed thus we need to remove this connenction from the users
// we will ping based on a new copy of our user:connection map
// this ping help us maintain only the available connection in our users
func (cht *Chat) ping() {
	ticker := time.NewTicker(time.Second * 10)

	go func() {

		for {
			cht.log.Info(context.Background(), "Ping", "status", "started")

			<-ticker.C // this is blocking code every 10 seconds this loop run
			fmt.Println("users now ", cht.users, "length", len(cht.users))
			for k, v := range cht.connections() {
				if err := v.conn.WriteMessage(websocket.PingMessage, []byte("ping")); err != nil {
					// here we need to remove this connection it fails send msg because conn is closed
					cht.log.Info(context.Background(), "rPing", "status", "failed", "error", err)
					cht.log.Info(context.Background(), "remove user", "user id", k, "user name", v.name)
					cht.removeUser(k)

				}
			}

			cht.log.Info(context.Background(), "Ping", "status", "completed", "available connection", cht.users)

		}
	}()

}
func (cht *Chat) writeMessage(data []byte, conn *websocket.Conn, msgType int) error {
	if err := conn.WriteMessage(msgType, data); err != nil {
		return fmt.Errorf("error writing message: %w", err)
	}
	return nil
}
func (cht *Chat) readMessages(cxt context.Context, conn *websocket.Conn) ([]byte, error) {
	type Response struct {
		data []byte
		err  error
	}
	respChan := make(chan Response)

	go func() {
		cht.log.Info(cxt, "starting read")
		defer cht.log.Info(cxt, "end read")
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

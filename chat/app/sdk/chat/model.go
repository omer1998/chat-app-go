package chat

import (
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type connection struct {
	id   uuid.UUID
	name string
	conn *websocket.Conn
}

type User struct {
	Id   uuid.UUID
	Name string
	Conn *websocket.Conn
}

type InMessage struct {
	FromId uuid.UUID `json:"fromId"`
	ToId   uuid.UUID `json:"toId"`
	Msg    string    `json:"msg"`
}
type OutMessage struct {
	From User   `json:"from"`
	To   User   `json:"to"`
	Msg  string `json:"msg"`
}

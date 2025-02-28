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

type user struct {
	Id   uuid.UUID
	Name string
}

type InMessage struct {
	FromId uuid.UUID `json:"fromId"`
	ToId   uuid.UUID `json:"toId"`
	Msg    string    `json:"msg"`
}
type OutMessage struct {
	From user   `json:"from"`
	To   user   `json:"to"`
	Msg  string `json:"msg"`
}

package chat

import (
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type User struct {
	Id   uuid.UUID
	Name string
	Conn *websocket.Conn
}

type InMessage struct {
	ToId uuid.UUID `json:"toId"`
	Msg  string    `json:"msg"`
}
type OutMessage struct {
	To  User   `json:"to"`
	Msg string `json:"msg"`
}
